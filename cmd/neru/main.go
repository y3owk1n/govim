package main

import (
	"fmt"
	"os"

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
	ModeHints
)

// Action represents the current action within hints mode
type Action int

const (
	ActionLeftClick Action = iota
	ActionRightClick
	ActionDoubleClick
	ActionTripleClick
	ActionMouseUp
	ActionMouseDown
	ActionMiddleClick
	ActionMoveMouse
	ActionScroll
	ActionContextMenu
)

// App represents the main application
type App struct {
	config           *config.Config
	ConfigPath       string
	logger           *zap.Logger
	hotkeyManager    *hotkeys.Manager
	hintGenerator    *hints.Generator
	hintOverlay      *hints.Overlay
	scrollController *scroll.Controller
	eventTap         *eventtap.EventTap
	ipcServer        *ipc.Server
	electronManager  *electron.ElectronManager
	appWatcher       *appwatcher.Watcher
	currentMode      Mode
	currentAction    Action
	hintsRouter      *hints.Router
	hintManager      *hints.Manager
	hintsCtx         *HintsContext

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
		currentAction:     ActionLeftClick,
		enabled:           true,
		hotkeysRegistered: false,
	}

	// Initialize hint managers
	app.hintManager = hints.NewManager(func(hs []*hints.Hint) {
		style := hints.BuildStyleForAction(app.config.Hints, getActionString(app.currentAction))
		if err := app.hintOverlay.DrawHintsWithStyle(hs, style); err != nil {
			app.logger.Error("Failed to redraw hints", zap.Error(err))
		}
	})
	// Initialize hints router
	app.hintsRouter = hints.NewRouter(app.hintManager)
	app.hintsCtx = &HintsContext{}

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

var globalApp *App

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
