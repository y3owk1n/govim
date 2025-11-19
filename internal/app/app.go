package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/domain/state"
	"github.com/y3owk1n/neru/internal/features/action"
	"github.com/y3owk1n/neru/internal/features/grid"
	"github.com/y3owk1n/neru/internal/features/hints"
	"github.com/y3owk1n/neru/internal/features/scroll"
	"github.com/y3owk1n/neru/internal/infra/accessibility"
	"github.com/y3owk1n/neru/internal/infra/appwatcher"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"github.com/y3owk1n/neru/internal/infra/eventtap"
	"github.com/y3owk1n/neru/internal/infra/hotkeys"
	"github.com/y3owk1n/neru/internal/infra/ipc"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"github.com/y3owk1n/neru/internal/ui"
	"github.com/y3owk1n/neru/internal/ui/overlay"
	"go.uber.org/zap"
)

// Mode is the current mode of the application.
type Mode = state.Mode

// Mode constants from state package.
const (
	ModeIdle  = state.ModeIdle
	ModeHints = state.ModeHints
	ModeGrid  = state.ModeGrid
)

// App represents the main application instance containing all state and dependencies.
type App struct {
	config     *config.Config
	ConfigPath string
	logger     *zap.Logger

	state  *state.AppState
	cursor *state.CursorState

	overlayManager   *overlay.Manager
	hotkeyManager    hotkeyService
	eventTap         eventTap
	ipcServer        ipcServer
	appWatcher       *appwatcher.Watcher
	scrollController *scroll.Controller

	hintGenerator *hints.Generator
	hintOverlay   *hints.Overlay
	hintManager   *hints.Manager
	hintsRouter   *hints.Router
	hintsCtx      *hints.Context
	hintStyle     hints.StyleMode

	gridManager *grid.Manager
	gridRouter  *grid.Router
	gridCtx     *grid.Context
	gridStyle   grid.Style

	scrollOverlay *scroll.Overlay
	actionOverlay *action.Overlay
	renderer      *ui.OverlayRenderer

	scrollCtx *scroll.Context

	cmdHandlers map[string]func(ipc.Command) ipc.Response
}

type hotkeyService interface {
	Register(keyString string, callback hotkeys.Callback) (hotkeys.HotkeyID, error)
	UnregisterAll()
}

type eventTap interface {
	Enable()
	Disable()
	Destroy()
	SetHotkeys(hotkeys []string)
}

type ipcServer interface {
	Start()
	Stop() error
}

type eventTapFactory interface {
	New(callback func(string), logger *zap.Logger) eventTap
}

type ipcServerFactory interface {
	New(handler func(ipc.Command) ipc.Response, logger *zap.Logger) (ipcServer, error)
}

type deps struct {
	Hotkeys          hotkeyService
	EventTapFactory  eventTapFactory
	IPCServerFactory ipcServerFactory
}

// New creates a new App instance.
func New(cfg *config.Config) (*App, error) {
	return newWithDeps(cfg, nil)
}

func newWithDeps(cfg *config.Config, deps *deps) (*App, error) {
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

	var hotkeySvc hotkeyService
	if deps != nil && deps.Hotkeys != nil {
		hotkeySvc = deps.Hotkeys
	} else {
		mgr := hotkeys.NewManager(log)
		hotkeys.SetGlobalManager(mgr)
		hotkeySvc = mgr
	}

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
		config:           cfg,
		logger:           log,
		state:            state.NewAppState(),
		cursor:           state.NewCursorState(cfg.General.RestoreCursorPosition),
		overlayManager:   overlayManager,
		hotkeyManager:    hotkeySvc,
		appWatcher:       appWatcher,
		hintGenerator:    hintGen,
		hintOverlay:      hintOverlay,
		scrollOverlay:    scrollOverlay,
		scrollController: scrollCtrl,
		scrollCtx:        &scroll.Context{},
		cmdHandlers:      make(map[string]func(ipc.Command) ipc.Response),
	}

	if cfg.Hints.Enabled {
		app.hintStyle = hints.BuildStyle(cfg.Hints)
		app.hintManager = hints.NewManager(func(hs []*hints.Hint) {
			err := app.hintOverlay.DrawHintsWithStyle(hs, app.hintStyle)
			if err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}, app.logger)
		app.hintsRouter = hints.NewRouter(app.hintManager, app.logger)
		app.hintsCtx = &hints.Context{}

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
		app.gridStyle = grid.BuildStyle(cfg.Grid)
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
			gridOverlay.ShowSubgrid(cell, app.gridStyle)
		}, log)

		app.gridCtx = &grid.Context{
			GridInstance: &gridInstance,
			GridOverlay:  &gridOverlay,
		}
	} else {
		var gridInstance *grid.Grid
		app.gridCtx = &grid.Context{
			GridInstance: &gridInstance,
		}
	}

	app.renderer = ui.NewOverlayRenderer(overlayManager, app.hintStyle, app.gridStyle)

	if deps != nil && deps.EventTapFactory != nil {
		app.eventTap = deps.EventTapFactory.New(app.handleKeyPress, log)
	} else {
		app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	}
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

	if deps != nil && deps.IPCServerFactory != nil {
		srv, srvErr := deps.IPCServerFactory.New(app.handleIPCCommand, log)
		if srvErr != nil {
			return nil, fmt.Errorf("failed to create IPC server: %w", srvErr)
		}
		app.ipcServer = srv
	} else {
		srv, srvErr := ipc.NewServer(app.handleIPCCommand, log)
		if srvErr != nil {
			return nil, fmt.Errorf("failed to create IPC server: %w", srvErr)
		}
		app.ipcServer = srv
	}

	overlayManager.UseScrollOverlay(scrollOverlay)
	overlayManager.UseActionOverlay(actionOverlay)
	if app.hintOverlay != nil {
		overlayManager.UseHintOverlay(app.hintOverlay)
	}
	if app.gridCtx != nil && app.gridCtx.GetGridOverlay() != nil {
		overlayManager.UseGridOverlay(*app.gridCtx.GetGridOverlay())
	}

	app.cmdHandlers[domain.CommandPing] = app.handlePing
	app.cmdHandlers[domain.CommandStart] = app.handleStart
	app.cmdHandlers[domain.CommandStop] = app.handleStop
	app.cmdHandlers[domain.ModeHints] = app.handleHints
	app.cmdHandlers[domain.ModeGrid] = app.handleGrid
	app.cmdHandlers[domain.CommandAction] = app.handleAction
	app.cmdHandlers[domain.ModeIdle] = app.handleIdle
	app.cmdHandlers[domain.CommandStatus] = app.handleStatus
	app.cmdHandlers[domain.CommandConfig] = app.handleConfig

	return app, nil
}

// ActivateMode activates the specified mode.
func (a *App) ActivateMode(mode Mode) { a.activateMode(mode) }

// SetEnabled sets the enabled state of the application.
func (a *App) SetEnabled(v bool) { a.state.SetEnabled(v) }

// IsEnabled returns the enabled state of the application.
func (a *App) IsEnabled() bool { return a.state.IsEnabled() }

// HintsEnabled returns true if hints are enabled.
func (a *App) HintsEnabled() bool { return a.config != nil && a.config.Hints.Enabled }

// GridEnabled returns true if grid is enabled.
func (a *App) GridEnabled() bool { return a.config != nil && a.config.Grid.Enabled }

// Config returns the application configuration.
func (a *App) Config() *config.Config { return a.config }

// Logger returns the application logger.
func (a *App) Logger() *zap.Logger { return a.logger }

// OverlayManager returns the overlay manager.
func (a *App) OverlayManager() *overlay.Manager { return a.overlayManager }

// HintGenerator returns the hint generator.
func (a *App) HintGenerator() *hints.Generator { return a.hintGenerator }

// HintManager returns the hint manager.
func (a *App) HintManager() *hints.Manager { return a.hintManager }

// HintsContext returns the hints context.
func (a *App) HintsContext() *hints.Context { return a.hintsCtx }

// Renderer returns the overlay renderer.
func (a *App) Renderer() *ui.OverlayRenderer { return a.renderer }

// GetConfigPath returns the config path.
func (a *App) GetConfigPath() string { return a.ConfigPath }

// SetHintOverlayNeedsRefresh sets the hint overlay needs refresh flag.
func (a *App) SetHintOverlayNeedsRefresh(value bool) { a.state.SetHintOverlayNeedsRefresh(value) }

// CaptureInitialCursorPosition captures the initial cursor position.
func (a *App) CaptureInitialCursorPosition() { a.captureInitialCursorPosition() }

// UpdateRolesForCurrentApp updates roles for the current app.
func (a *App) UpdateRolesForCurrentApp() { a.updateRolesForCurrentApp() }

// CollectElements collects elements.
func (a *App) CollectElements() []*accessibility.TreeNode { return a.collectElements() }

// IsFocusedAppExcluded checks if the focused app is excluded.
func (a *App) IsFocusedAppExcluded() bool { return a.isFocusedAppExcluded() }

// ExitMode exits the current mode.
func (a *App) ExitMode() { a.exitMode() }

// GridManager returns the grid manager.
func (a *App) GridManager() *grid.Manager { return a.gridManager }

// GridContext returns the grid context.
func (a *App) GridContext() *grid.Context { return a.gridCtx }

// GridRouter returns the grid router.
func (a *App) GridRouter() *grid.Router { return a.gridRouter }

// ScrollContext returns the scroll context.
func (a *App) ScrollContext() *scroll.Context { return a.scrollCtx }

// HintsRouter returns the hints router.
func (a *App) HintsRouter() *hints.Router { return a.hintsRouter }

// EventTap returns the event tap.
func (a *App) EventTap() eventTap { return a.eventTap }

// ScrollController returns the scroll controller.
func (a *App) ScrollController() *scroll.Controller { return a.scrollController }

// CurrentMode returns the current mode.
func (a *App) CurrentMode() Mode { return a.state.CurrentMode() }

// SetModeHints sets the mode to hints.
func (a *App) SetModeHints() { a.setModeHints() }

// SetModeGrid sets the mode to grid.
func (a *App) SetModeGrid() { a.setModeGrid() }

// SetModeIdle sets the mode to idle.
func (a *App) SetModeIdle() { a.setModeIdle() }

// EnableEventTap enables the event tap.
func (a *App) EnableEventTap() { a.enableEventTap() }

// DisableEventTap disables the event tap.
func (a *App) DisableEventTap() { a.disableEventTap() }

// overlaySwitch switches the overlay mode.
func (a *App) overlaySwitch(m overlay.Mode) {
	if a.overlayManager != nil {
		a.overlayManager.SwitchTo(m)
	}
}

func (a *App) enableEventTap() {
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
}

func (a *App) disableEventTap() {
	if a.eventTap != nil {
		a.eventTap.Disable()
	}
}

func (a *App) setModeHints() {
	a.state.SetMode(ModeHints)
	a.enableEventTap()
	a.overlaySwitch(overlay.ModeHints)
}

func (a *App) setModeGrid() {
	a.state.SetMode(ModeGrid)
	a.enableEventTap()
	a.overlaySwitch(overlay.ModeGrid)
}

func (a *App) setModeIdle() {
	a.state.SetMode(ModeIdle)
	a.disableEventTap()
	a.overlaySwitch(overlay.ModeIdle)
}
