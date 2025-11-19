package app

import (
	"fmt"
	"image"
	"strings"

	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/features/grid"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"go.uber.org/zap"
)

func (a *App) activateGridMode() {
	if !a.state.IsEnabled() {
		a.logger.Debug("Neru is disabled, ignoring grid mode activation")
		return
	}

	if !a.config.Grid.Enabled {
		a.logger.Debug("Grid mode disabled by config, ignoring activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	if a.scrollComponent.Context.GetIsActive() {
		// Reset scroll context to ensure clean transition
		a.scrollComponent.Context.SetIsActive(false)
		a.scrollComponent.Context.SetLastKey("")
		// Also reset the skip restore flag since we're transitioning from scroll to grid mode
		a.cursor.Reset()
	}

	a.captureInitialCursorPosition()

	action := domain.ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating grid mode", zap.String("action", actionString))

	a.exitMode()

	// Always resize overlay to the active screen (where mouse is) before drawing grid.
	// This ensures proper positioning when switching between multiple displays.
	a.renderer.ResizeActive()
	a.state.SetGridOverlayNeedsRefresh(false)

	err := a.setupGrid()
	if err != nil {
		a.logger.Error("Failed to setup grid",
			zap.Error(err),
			zap.String("action", actionString),
			zap.String("screen_bounds", fmt.Sprintf("%+v", bridge.GetActiveScreenBounds())))
		return
	}

	a.setModeGrid()

	a.logger.Info("Grid mode activated", zap.String("action", actionString))
	a.logger.Info("Type a grid label to select a location")
}

// setupGrid generates grid and draws it.
func (a *App) setupGrid() error {
	// Create grid with active screen bounds (screen containing mouse cursor).
	// This ensures proper multi-monitor support.
	gridInstance := a.createGridInstance()

	a.updateGridOverlayConfig()

	// Ensure the overlay is properly sized for the active screen
	a.overlayManager.ResizeToActiveScreenSync()

	// Reset the grid manager state when setting up the grid
	if a.GridManager() != nil {
		a.GridManager().Reset()
	}

	a.initializeGridManager(gridInstance)

	a.gridComponent.Router = grid.NewRouter(a.GridManager(), a.logger)

	initErr := a.renderer.DrawGrid(gridInstance, "")
	if initErr != nil {
		return fmt.Errorf("failed to draw grid: %w", initErr)
	}
	a.renderer.Show()

	return nil
}

// createGridInstance creates a new grid instance with proper bounds and characters.
func (a *App) createGridInstance() *grid.Grid {
	screenBounds := bridge.GetActiveScreenBounds()

	// Normalize bounds to window-local coordinates (0,0 origin).
	// The overlay window is positioned at the screen origin, but the view uses local coordinates.
	bounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	characters := a.config.Grid.Characters
	if strings.TrimSpace(characters) == "" {
		characters = a.config.Hints.HintCharacters
	}
	gridInstance := grid.NewGrid(characters, bounds, a.logger)
	a.gridComponent.Context.SetGridInstanceValue(gridInstance)

	return gridInstance
}

// updateGridOverlayConfig updates the grid overlay configuration.
func (a *App) updateGridOverlayConfig() {
	(*a.gridComponent.Context.GridOverlay).UpdateConfig(a.config.Grid)
}

// initializeGridManager initializes the grid manager with the new grid instance.
func (a *App) initializeGridManager(gridInstance *grid.Grid) {
	const defaultGridCharacters = "asdfghjkl"

	// Defensive check for grid instance
	if gridInstance == nil {
		a.logger.Warn("Grid instance is nil, creating with default bounds")
		screenBounds := bridge.GetActiveScreenBounds()
		bounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())
		gridInstance = grid.NewGrid(a.config.Grid.Characters, bounds, a.logger)
	}

	// Subgrid configuration and keys (fallback to grid characters): always 3x3
	keys := strings.TrimSpace(a.config.Grid.SublayerKeys)
	if keys == "" {
		keys = a.config.Grid.Characters
	}

	// Ensure we have valid keys for subgrid
	if keys == "" {
		a.logger.Warn("No subgrid keys configured, using grid characters as fallback")
		keys = a.config.Grid.Characters
	}

	// Final fallback
	if keys == "" {
		keys = defaultGridCharacters
		a.logger.Warn("No characters available for subgrid, using default")
	}

	const subRows = 3
	const subCols = 3

	a.gridComponent.Manager = grid.NewManager(
		gridInstance,
		subRows,
		subCols,
		keys,
		func(forceRedraw bool) {
			// Defensive check for grid manager
			if a.GridManager() == nil {
				a.logger.Error("Grid manager is nil during update callback")
				return
			}

			input := a.GridManager().GetInput()

			// special case to handle only when exiting subgrid
			if forceRedraw {
				a.renderer.Clear()
				gridErr := a.renderer.DrawGrid(gridInstance, input)
				if gridErr != nil {
					a.logger.Error("Failed to redraw grid", zap.Error(gridErr))
					return
				}
				a.renderer.Show()
			}

			// Set hideUnmatched based on whether we have input and the config setting
			hideUnmatched := a.config.Grid.HideUnmatched && len(input) > 0
			a.renderer.SetHideUnmatched(hideUnmatched)
			a.renderer.UpdateGridMatches(input)
		},
		func(cell *grid.Cell) {
			// Defensive check for cell
			if cell == nil {
				a.logger.Warn("Attempted to show subgrid for nil cell")
				return
			}

			// Draw 3x3 subgrid inside selected cell
			a.renderer.ShowSubgrid(cell)
		},
		a.logger,
	)
}

// drawGridActionHighlight draws a highlight border around the active screen for grid action mode.
func (a *App) drawGridActionHighlight() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	a.renderer.DrawActionHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)

	a.logger.Debug("Drawing grid action highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))
}

// handleGridActionKey handles action keys when in grid action mode.
func (a *App) handleGridActionKey(key string) {
	a.handleActionKey(key, "Grid")
}
