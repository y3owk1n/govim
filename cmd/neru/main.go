package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/appwatcher"
	_ "github.com/y3owk1n/neru/internal/bridge" // Import for CGo compilation
	"github.com/y3owk1n/neru/internal/cli"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/electron"
	"github.com/y3owk1n/neru/internal/eventtap"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/hotkeys"
	"github.com/y3owk1n/neru/internal/ipc"
	"github.com/y3owk1n/neru/internal/logger"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
)

// Mode represents the current application mode
type Mode int

const (
	ModeIdle Mode = iota
	ModeHint
	ModeHintWithActions
	ModeScroll
)

// App represents the main application
type App struct {
	config            *config.Config
	ConfigPath        string
	logger            *zap.Logger
	hotkeyManager     *hotkeys.Manager
	hintGenerator     *hints.Generator
	hintOverlay       *hints.Overlay
	scrollController  *scroll.Controller
	eventTap          *eventtap.EventTap
	ipcServer         *ipc.Server
	electronManager   *electron.ElectronManager
	appWatcher        *appwatcher.Watcher
	currentMode       Mode
	hintManager       *hints.Manager
	lastScrollKey     string
	selectedHint      *hints.Hint
	canScroll         bool
	enabled           bool
	hotkeysRegistered bool
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	if err := logger.Init(cfg.Logging.LogLevel, cfg.Logging.LogFile, cfg.Logging.StructuredLogging); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Get()

	// Check accessibility permissions
	if cfg.Accessibility.AccessibilityCheckOnStart {
		if !accessibility.CheckAccessibilityPermissions() {
			log.Warn("Accessibility permissions not granted. Please grant permissions in System Settings.")
			fmt.Println("⚠️  Neru requires Accessibility permissions to function.")
			fmt.Println("Please go to: System Settings → Privacy & Security → Accessibility")
			fmt.Println("and enable Neru.")
			return nil, fmt.Errorf("accessibility permissions required")
		}
	}

	// Set global config
	config.SetGlobal(cfg)

	// Apply clickable and scrollable roles from config
	log.Info("Applying clickable roles",
		zap.Int("count", len(cfg.Accessibility.ClickableRoles)),
		zap.Strings("roles", cfg.Accessibility.ClickableRoles))
	accessibility.SetClickableRoles(cfg.Accessibility.ClickableRoles)

	log.Info("Applying scrollable roles",
		zap.Int("count", len(cfg.Accessibility.ScrollableRoles)),
		zap.Strings("roles", cfg.Accessibility.ScrollableRoles))
	accessibility.SetScrollableRoles(cfg.Accessibility.ScrollableRoles)

	// Create app watcher
	appWatcher := appwatcher.NewWatcher()

	// Create hotkey manager
	hotkeyMgr := hotkeys.NewManager(log)
	hotkeys.SetGlobalManager(hotkeyMgr)

	// Create hint generator
	hintGen := hints.NewGenerator(
		cfg.Hints.HintCharacters,
	)

	// Create hint overlay
	hintOverlay, err := hints.NewOverlay(cfg.Hints)
	if err != nil {
		return nil, fmt.Errorf("failed to create overlay: %w", err)
	}

	// Create scroll controller
	scrollCtrl := scroll.NewController(cfg.Scroll, log)

	app := &App{
		config:            cfg,
		logger:            log,
		hotkeyManager:     hotkeyMgr,
		appWatcher:        appWatcher,
		hintGenerator:     hintGen,
		hintOverlay:       hintOverlay,
		scrollController:  scrollCtrl,
		currentMode:       ModeIdle,
		enabled:           true,
		hotkeysRegistered: false,
	}

	// Initialize hint managers
	app.hintManager = hints.NewManager(func(hints []*hints.Hint) {
		switch app.currentMode {
		case ModeHintWithActions:
			style := app.getStyleForMode(ModeHintWithActions)
			if err := app.hintOverlay.DrawHintsWithStyle(hints, style); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		case ModeHint:
			style := app.getStyleForMode(ModeHint)
			if err := app.hintOverlay.DrawHintsWithStyle(hints, style); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		case ModeScroll:
			style := app.getStyleForMode(ModeScroll)
			if err := app.hintOverlay.DrawHintsWithStyle(hints, style); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}
	})

	// Create electron manager
	if cfg.Accessibility.AdditionalAXSupport.Enable {
		app.electronManager = electron.NewElectronManager(cfg.Accessibility.AdditionalAXSupport.AdditionalElectronBundles, cfg.Accessibility.AdditionalAXSupport.AdditionalChromiumBundles, cfg.Accessibility.AdditionalAXSupport.AdditionalFirefoxBundles)
	}

	// Create event tap for capturing keys in modes
	app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - key capture won't work")
	} else {

		keys := make([]string, 0, len(cfg.Hotkeys.Bindings))
		for k := range cfg.Hotkeys.Bindings {
			keys = append(keys, k)
		}

		// Configure hotkeys that should pass through to the global hotkey system
		app.eventTap.SetHotkeys(keys)

		// Ensure event tap is disabled initially (only enable in active modes)
		app.eventTap.Disable()
	}

	// Create IPC server
	ipcServer, err := ipc.NewServer(app.handleIPCCommand, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC server: %w", err)
	}
	app.ipcServer = ipcServer

	return app, nil
}

// Run runs the application
func (a *App) Run() error {
	a.logger.Info("Starting Neru")

	// Start IPC server
	a.ipcServer.Start()
	a.logger.Info("IPC server started")

	// Start app watcher
	a.appWatcher.Start()
	a.logger.Info("App watcher started")

	// Initialize hotkeys based on current focused app and exclusion
	a.refreshHotkeysForAppOrCurrent("")
	a.logger.Info("Hotkeys initialized")

	// setup watcher callbacks
	a.appWatcher.OnActivate(func(appName, bundleID string) {
		a.logger.Debug("App activated", zap.String("bundle_id", bundleID))

		// refresh hotkeys for app
		go a.refreshHotkeysForAppOrCurrent(bundleID)
		a.logger.Debug("Handled hotkey refresh")

		if a.config.Accessibility.AdditionalAXSupport.Enable {
			// handle electrons
			if electron.ShouldEnableElectronSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalElectronBundles) {
				electron.EnsureElectronAccessibility(bundleID)
				a.logger.Debug("Handled electron accessibility")
			}

			// handle chromium
			if electron.ShouldEnableChromiumSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalChromiumBundles) {
				electron.EnsureChromiumAccessibility(bundleID)
				a.logger.Debug("Handled chromium accessibility")
			}

			// handle firefox
			if electron.ShouldEnableFirefoxSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalFirefoxBundles) {
				electron.EnsureFirefoxAccessibility(bundleID)
				a.logger.Debug("Handled firefox accessibility")
			}

		}

		a.logger.Debug("Done handling app activation")
	})

	a.logger.Info("Neru is running")
	fmt.Println("✓ Neru is running")

	for k, v := range a.config.Hotkeys.Bindings {
		toShow := v
		if strings.HasPrefix(v, "exec") {
			runes := []rune(v)
			if len(runes) > 30 {
				toShow = string(runes[:30]) + "..."
			}
		}
		fmt.Printf("  %s: %s\n", k, toShow)
	}

	// Wait for interrupt signal with force-quit support
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// First signal: graceful shutdown
	<-sigChan
	a.logger.Info("Received shutdown signal, starting graceful shutdown...")
	fmt.Println("\n⚠️  Shutting down gracefully... (press Ctrl+C again to force quit)")

	// Start cleanup in goroutine
	done := make(chan struct{})
	go func() {
		// Quit systray to exit the event loop
		systray.Quit()
		close(done)
	}()

	// Wait for cleanup or second signal
	select {
	case <-done:
		a.logger.Info("Graceful shutdown completed")
		return nil
	case <-sigChan:
		// Second signal: force quit
		a.logger.Warn("Received second signal, forcing shutdown")
		fmt.Println("⚠️  Force quitting...")
		os.Exit(1)
	case <-time.After(10 * time.Second):
		// Timeout: force quit
		a.logger.Error("Shutdown timeout exceeded, forcing shutdown")
		fmt.Println("⚠️  Shutdown timeout, force quitting...")
		os.Exit(1)
	}

	return nil
}

// Cleanup cleans up resources
func (a *App) Cleanup() {
	a.logger.Info("Cleaning up")

	// Exit any active mode first
	a.exitMode()

	// Stop IPC server first to prevent new requests
	if a.ipcServer != nil {
		if err := a.ipcServer.Stop(); err != nil {
			a.logger.Error("Failed to stop IPC server", zap.Error(err))
		}
	}

	// Unregister all hotkeys
	if a.hotkeyManager != nil {
		a.hotkeyManager.UnregisterAll()
	}

	// Clean up UI elements
	if a.hintOverlay != nil {
		a.hintOverlay.Destroy()
	}

	// Cleanup event tap
	if a.eventTap != nil {
		a.eventTap.Destroy()
	}

	// Sync and close logger at the end
	if err := logger.Sync(); err != nil {
		// Ignore "inappropriate ioctl for device" error which occurs when syncing stdout/stderr
		if !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			a.logger.Error("Failed to sync logger", zap.Error(err))
		}
	}

	// Stop app watcher
	a.appWatcher.Stop()

	// Close logger (syncs and closes log file)
	if err := logger.Close(); err != nil {
		// Can't log this since logger is being closed
		fmt.Fprintf(os.Stderr, "Warning: failed to close logger: %v\n", err)
	}
}

// handleIPCCommand handles IPC commands from the CLI
func (a *App) handleIPCCommand(cmd ipc.Command) ipc.Response {
	a.logger.Info("Handling IPC command", zap.String("action", cmd.Action))

	switch cmd.Action {
	case "ping":
		return ipc.Response{Success: true, Message: "pong"}

	case "start":
		if a.enabled {
			return ipc.Response{Success: false, Message: "neru is already running"}
		}
		a.enabled = true
		return ipc.Response{Success: true, Message: "neru started"}

	case "stop":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is already stopped"}
		}
		a.enabled = false
		a.exitMode()
		return ipc.Response{Success: true, Message: "neru stopped"}

	case "hints":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateHintMode(ModeHint)
		return ipc.Response{Success: true, Message: "hint mode activated"}

	case "hints_action":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateHintMode(ModeHintWithActions)
		return ipc.Response{Success: true, Message: "hint mode with actions activated"}

	case "scroll":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateHintMode(ModeScroll)
		return ipc.Response{Success: true, Message: "scroll mode activated"}

	case "idle":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.exitMode()
		return ipc.Response{Success: true, Message: "mode set to idle"}

	case "status":
		cfgPath := a.ConfigPath

		if cfgPath == "" {
			// Fallback to the standard config path if daemon wasn't started
			// with an explicit --config
			cfgPath = config.FindConfigFile()
		}

		// If config file doesn't exist, return default config
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			cfgPath = "No config file found, using default config without config file"
		} else {
			// Expand ~ to home dir and resolve relative paths to absolute
			if strings.HasPrefix(cfgPath, "~") {
				if home, err := os.UserHomeDir(); err == nil {
					cfgPath = filepath.Join(home, cfgPath[1:])
				}
			}
			if abs, err := filepath.Abs(cfgPath); err == nil {
				cfgPath = abs
			}
		}

		statusData := map[string]any{
			"enabled": a.enabled,
			"mode":    a.getCurrModeString(),
			"config":  cfgPath,
		}
		return ipc.Response{Success: true, Data: statusData}

	default:
		return ipc.Response{Success: false, Message: fmt.Sprintf("unknown command: %s", cmd.Action)}
	}
}

var globalApp *App

func onReady() {
	systray.SetTitle("⌨️")
	systray.SetTooltip("Neru - Keyboard Navigation")

	// Status submenu for version
	mVersion := systray.AddMenuItem(fmt.Sprintf("Version %s", cli.Version), "Show version")
	mVersion.Disable()
	mVersionCopy := systray.AddMenuItem("Copy version", "Copy version to clipboard")

	// Status toggle
	systray.AddSeparator()
	mStatus := systray.AddMenuItem("Status: Running", "Show current status")
	mStatus.Disable()
	mToggle := systray.AddMenuItem("Disable", "Disable/Enable Neru without quitting")

	// Control actions
	systray.AddSeparator()
	mHints := systray.AddMenuItem("Hints Mode", "Show hints")
	mHintsWithActions := systray.AddMenuItem("Hints Mode with Actions", "Show hints with actions")
	mScroll := systray.AddMenuItem("Scroll Mode", "Show scroll hints")

	// Quit option
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Neru", "Exit the application")

	// Handle clicks in a separate goroutine
	go func() {
		for {
			select {
			case <-mVersionCopy.ClickedCh:
				err := clipboard.WriteAll(cli.Version)
				if err != nil {
					logger.Error("Error copying version to clipboard", zap.Error(err))
				}
			case <-mToggle.ClickedCh:
				if globalApp != nil {
					if globalApp.enabled {
						globalApp.enabled = false
						mStatus.SetTitle("Status: Disabled")
						mToggle.SetTitle("Enable")
					} else {
						globalApp.enabled = true
						mStatus.SetTitle("Status: Enabled")
						mToggle.SetTitle("Disable")
					}
				}
			case <-mHints.ClickedCh:
				if globalApp != nil {
					globalApp.activateHintMode(ModeHint)
				}
			case <-mHintsWithActions.ClickedCh:
				if globalApp != nil {
					globalApp.activateHintMode(ModeHintWithActions)
				}
			case <-mScroll.ClickedCh:
				if globalApp != nil {
					globalApp.activateHintMode(ModeScroll)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	if globalApp != nil {
		globalApp.Cleanup()
	}
}

func main() {
	// Set the launch function for CLI
	cli.LaunchFunc = LaunchDaemon

	// Let cobra handle all command parsing
	cli.Execute()
}

// LaunchDaemon is called by the CLI to launch the daemon
func LaunchDaemon(configPath string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create app
	app, err := NewApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}
	// Record the config path the daemon was started with so status can report
	// the same path (if provided via --config)
	app.ConfigPath = configPath
	globalApp = app

	// Start app in goroutine
	go func() {
		if err := app.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		}
	}()

	// Run systray (blocks until quit)
	systray.Run(onReady, onExit)
}
