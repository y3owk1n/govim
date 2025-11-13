package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/appwatcher"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/cli"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/electron"
	"github.com/y3owk1n/neru/internal/eventtap"
	"github.com/y3owk1n/neru/internal/grid"
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
	ModeGrid
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
	gridManager      *grid.Manager
	gridRouter       *grid.Router
	gridCtx          *GridContext

	enabled                 bool
	hotkeysRegistered       bool
	screenChangeProcessing  bool
	gridOverlayNeedsRefresh bool
	hintOverlayNeedsRefresh bool
	hotkeyRefreshPending    bool
	idleScrollLastKey       string // Track last scroll key in idle mode for gg support
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	if err := logger.Init(cfg.Logging.LogLevel, cfg.Logging.LogFile, cfg.Logging.StructuredLogging); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Get()

	// Initialize bridge logger
	bridge.InitializeLogger(log)

	// Check accessibility permissions
	if cfg.General.AccessibilityCheckOnStart { // Changed from cfg.Accessibility.AccessibilityCheckOnStart
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

	if cfg.Hints.Enabled {
		// Apply clickable and scrollable roles from config
		log.Info("Applying clickable roles",
			zap.Int("count", len(cfg.Hints.ClickableRoles)), // Changed from cfg.Accessibility.ClickableRoles
			zap.Strings("roles", cfg.Hints.ClickableRoles))  // Changed from cfg.Accessibility.ClickableRoles
		accessibility.SetClickableRoles(cfg.Hints.ClickableRoles) // Changed from cfg.Accessibility.ClickableRoles

		log.Info("Applying scrollable roles",
			zap.Int("count", len(cfg.Hints.ScrollableRoles)), // Changed from cfg.Accessibility.ScrollableRoles
			zap.Strings("roles", cfg.Hints.ScrollableRoles))  // Changed from cfg.Accessibility.ScrollableRoles
		accessibility.SetScrollableRoles(cfg.Hints.ScrollableRoles) // Changed from cfg.Accessibility.ScrollableRoles
	}

	// Create app watcher
	appWatcher := appwatcher.NewWatcher(log)

	// Create hotkey manager
	hotkeyMgr := hotkeys.NewManager(log)
	hotkeys.SetGlobalManager(hotkeyMgr)

	// Create hint generator only if hints are enabled
	var hintGen *hints.Generator
	if cfg.Hints.Enabled {
		hintGen = hints.NewGenerator(
			cfg.Hints.HintCharacters,
		)
	}

	// Create hint overlay if either hints or grid requires overlay drawing
	var hintOverlay *hints.Overlay
	if cfg.Hints.Enabled || cfg.Grid.Enabled {
		var err error
		hintOverlay, err = hints.NewOverlay(cfg.Hints, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create overlay: %w", err)
		}
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

	// Initialize hints components only if enabled
	if cfg.Hints.Enabled {
		app.hintManager = hints.NewManager(func(hs []*hints.Hint) {
			style := hints.BuildStyleForAction(app.config.Hints, getActionString(app.currentAction))
			if err := app.hintOverlay.DrawHintsWithStyle(hs, style); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}, app.logger)
		// Initialize hints router
		app.hintsRouter = hints.NewRouter(app.hintManager, app.logger)
		app.hintsCtx = &HintsContext{}
	}

	// Create grid components only if enabled
	if cfg.Grid.Enabled {
		// Create grid overlay upfront (like hints overlay) to avoid thread issues
		gridOverlay := grid.NewGridOverlay(cfg.Grid, log)

		// Grid instance will be created when activated (screen bounds may change)
		var gridInstance *grid.Grid
		// Subgrid config: always 3x3
		keys := strings.TrimSpace(cfg.Grid.SublayerKeys)
		if keys == "" {
			keys = cfg.Grid.Characters
		}
		const subRows = 3
		const subCols = 3
		app.gridManager = grid.NewManager(nil, subRows, subCols, keys, func(forceRedraw bool) {
			// Redraw grid overlay when input changes
			if gridInstance == nil {
				return
			}
			gridOverlay.UpdateMatches(app.gridManager.GetInput())
		}, func(cell *grid.Cell) {
			// Show subgrid in selected cell
			// Use default left_click style during initialization
			gridStyle := grid.BuildStyleForAction(cfg.Grid, "left_click")
			gridOverlay.ShowSubgrid(cell, gridStyle)
		}, log)

		// Store grid config for later creation
		app.gridCtx = &GridContext{
			currentAction: ActionLeftClick,
			gridInstance:  &gridInstance,
			gridOverlay:   &gridOverlay,
		}
	} else {
		// Minimal grid context without overlay/manager
		var gridInstance *grid.Grid
		app.gridCtx = &GridContext{
			currentAction: ActionLeftClick,
			gridInstance:  &gridInstance,
		}
	}

	// Create electron manager
	if cfg.Hints.Enabled {
		if cfg.Hints.AdditionalAXSupport.Enable { // Changed from cfg.Accessibility.AdditionalAXSupport.Enable
			app.electronManager = electron.NewElectronManager(cfg.Hints.AdditionalAXSupport.AdditionalElectronBundles, cfg.Hints.AdditionalAXSupport.AdditionalChromiumBundles, cfg.Hints.AdditionalAXSupport.AdditionalFirefoxBundles) // Changed from cfg.Accessibility.AdditionalAXSupport
		}
	}

	// Create event tap for capturing keys in modes
	app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - key capture won't work")
	} else {
		keys := make([]string, 0, len(cfg.Hotkeys.Bindings))
		for k, v := range cfg.Hotkeys.Bindings {
			mode := v
			if parts := strings.Split(v, " "); len(parts) > 0 {
				mode = parts[0]
			}
			if mode == "hints" && !cfg.Hints.Enabled {
				continue
			}
			if mode == "grid" && !cfg.Grid.Enabled {
				continue
			}
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
