// Package app contains the main application logic.
package app

import (
	"errors"
	"fmt"
	"image"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/action"
	"github.com/y3owk1n/neru/internal/appwatcher"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/eventtap"
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/hotkeys"
	"github.com/y3owk1n/neru/internal/ipc"
	"github.com/y3owk1n/neru/internal/logger"
	"github.com/y3owk1n/neru/internal/overlay"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
)

// Mode is the current mode of the application.
type Mode int

const (
	// ModeIdle is the default mode of the application.
	ModeIdle Mode = iota
	// ModeHints is used when the user is in hints mode.
	ModeHints
	// ModeGrid is used when the user is in grid mode.
	ModeGrid
)

// Action is the current action of the application.
type Action int

const (
	// ActionLeftClick is the action for left click.
	ActionLeftClick Action = iota
	// ActionRightClick is the action for right click.
	ActionRightClick
	// ActionMouseUp is the action for mouse up.
	ActionMouseUp
	// ActionMouseDown is the action for mouse down.
	ActionMouseDown
	// ActionMiddleClick is the action for middle click.
	ActionMiddleClick
	// ActionMoveMouse is the action for moving the mouse.
	ActionMoveMouse
	// ActionScroll is the action for scrolling.
	ActionScroll
)

// App is the main application struct.
type App struct {
	config           *config.Config
	ConfigPath       string
	logger           *zap.Logger
	hotkeyManager    *hotkeys.Manager
	hintGenerator    *hints.Generator
	hintOverlay      *hints.Overlay
	scrollOverlay    *scroll.Overlay
	actionOverlay    *action.Overlay
	scrollController *scroll.Controller
	eventTap         *eventtap.EventTap
	ipcServer        *ipc.Server
	appWatcher       *appwatcher.Watcher
	currentMode      Mode
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
	idleScrollLastKey       string

	initialCursorPos      image.Point
	initialScreenBounds   image.Rectangle
	cursorRestoreCaptured bool
	isScrollingActive     bool
	skipCursorRestoreOnce bool
}

// New creates a new App instance.
func New(cfg *config.Config) (*App, error) {
	var err error
	err = logger.Init(
		cfg.Logging.LogLevel,
		cfg.Logging.LogFile,
		cfg.Logging.StructuredLogging,
		cfg.Logging.DisableFileLogging,
		cfg.Logging.MaxFileSize,
		cfg.Logging.MaxBackups,
		cfg.Logging.MaxAge,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Get()

	bridge.InitializeLogger(log)

	overlayManager := overlay.Init(log)

	if cfg.General.AccessibilityCheckOnStart {
		if !accessibility.CheckAccessibilityPermissions() {
			log.Warn(
				"Accessibility permissions not granted. Please grant permissions in System Settings.",
			)
			log.Info("⚠️  Neru requires Accessibility permissions to function.")
			log.Info("Please go to: System Settings → Privacy & Security → Accessibility")
			log.Info("and enable Neru.")
			return nil, errors.New("accessibility permissions required")
		}
	}

	config.SetGlobal(cfg)

	if cfg.Hints.Enabled {
		log.Info("Applying clickable roles",
			zap.Int("count", len(cfg.Hints.ClickableRoles)),
			zap.Strings("roles", cfg.Hints.ClickableRoles))
		accessibility.SetClickableRoles(cfg.Hints.ClickableRoles)
	}

	appWatcher := appwatcher.NewWatcher(log)

	hotkeyMgr := hotkeys.NewManager(log)
	hotkeys.SetGlobalManager(hotkeyMgr)

	var hintGen *hints.Generator
	if cfg.Hints.Enabled {
		hintGen = hints.NewGenerator(cfg.Hints.HintCharacters)
	}

	var hintOverlay *hints.Overlay

	scrollCtrl := scroll.NewController(cfg.Scroll, log)

	scrollOverlay, err := scroll.NewOverlayWithWindow(
		cfg.Scroll,
		log,
		overlayManager.GetWindowPtr(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scroll overlay: %w", err)
	}

	app := &App{
		config:            cfg,
		logger:            log,
		hotkeyManager:     hotkeyMgr,
		appWatcher:        appWatcher,
		hintGenerator:     hintGen,
		hintOverlay:       hintOverlay,
		scrollOverlay:     scrollOverlay,
		scrollController:  scrollCtrl,
		currentMode:       ModeIdle,
		enabled:           true,
		hotkeysRegistered: false,
	}

	if cfg.Hints.Enabled {
		app.hintManager = hints.NewManager(func(hs []*hints.Hint) {
			style := hints.BuildStyle(app.config.Hints)
			err := app.hintOverlay.DrawHintsWithStyle(hs, style)
			if err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}, app.logger)
		app.hintsRouter = hints.NewRouter(app.hintManager, app.logger)
		app.hintsCtx = &HintsContext{}

		var err error
		hintOverlay, err = hints.NewOverlayWithWindow(cfg.Hints, log, overlayManager.GetWindowPtr())
		if err != nil {
			return nil, fmt.Errorf("failed to create overlay: %w", err)
		}
		app.hintOverlay = hintOverlay
	}

	actionOverlay, err := action.NewOverlayWithWindow(
		cfg.Action,
		log,
		overlayManager.GetWindowPtr(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create action overlay: %w", err)
	}
	app.actionOverlay = actionOverlay

	if cfg.Grid.Enabled {
		gridOverlay := grid.NewOverlayWithWindow(cfg.Grid, log, overlayManager.GetWindowPtr())
		var gridInstance *grid.Grid
		keys := strings.TrimSpace(cfg.Grid.SublayerKeys)
		if keys == "" {
			keys = cfg.Grid.Characters
		}
		const subRows = 3
		const subCols = 3
		app.gridManager = grid.NewManager(nil, subRows, subCols, keys, func(_ bool) {
			if gridInstance == nil {
				return
			}
			gridOverlay.UpdateMatches(app.gridManager.GetInput())
		}, func(cell *grid.Cell) {
			gridStyle := grid.BuildStyle(cfg.Grid)
			gridOverlay.ShowSubgrid(cell, gridStyle)
		}, log)

		app.gridCtx = &GridContext{
			gridInstance: &gridInstance,
			gridOverlay:  &gridOverlay,
		}
	} else {
		var gridInstance *grid.Grid
		app.gridCtx = &GridContext{
			gridInstance: &gridInstance,
		}
	}

	app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - key capture won't work")
	} else {
		keys := make([]string, 0, len(cfg.Hotkeys.Bindings))
		for key, value := range cfg.Hotkeys.Bindings {
			mode := value
			if parts := strings.Split(value, " "); len(parts) > 0 {
				mode = parts[0]
			}
			if mode == "hints" && !cfg.Hints.Enabled {
				continue
			}
			if mode == "grid" && !cfg.Grid.Enabled {
				continue
			}
			keys = append(keys, key)
		}
		app.eventTap.SetHotkeys(keys)
		app.eventTap.Disable()
	}

	ipcServer, err := ipc.NewServer(app.handleIPCCommand, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC server: %w", err)
	}
	app.ipcServer = ipcServer

	overlayManager.UseScrollOverlay(scrollOverlay)
	overlayManager.UseActionOverlay(actionOverlay)
	if app.hintOverlay != nil {
		overlayManager.UseHintOverlay(app.hintOverlay)
	}
	if app.gridCtx != nil && app.gridCtx.gridOverlay != nil {
		overlayManager.UseGridOverlay(*app.gridCtx.gridOverlay)
	}

	return app, nil
}

// ActivateMode activates the specified mode.
func (a *App) ActivateMode(mode Mode) { a.activateMode(mode) }

// SetEnabled sets the enabled state of the application.
func (a *App) SetEnabled(v bool) { a.enabled = v }

// IsEnabled returns the enabled state of the application.
func (a *App) IsEnabled() bool { return a.enabled }

// HintsEnabled returns true if hints are enabled.
func (a *App) HintsEnabled() bool { return a.config != nil && a.config.Hints.Enabled }

// GridEnabled returns true if grid is enabled.
func (a *App) GridEnabled() bool { return a.config != nil && a.config.Grid.Enabled }
