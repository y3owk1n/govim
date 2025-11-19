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
	// Respect mode enable flag
	if !a.config.Grid.Enabled {
		a.logger.Debug("Grid mode disabled by config, ignoring activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.captureInitialCursorPosition()

	action := domain.ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating grid mode", zap.String("action", actionString))

	a.exitMode() // Exit current mode first

	// Always resize overlay to the active screen (where mouse is) before drawing grid
	// This ensures proper positioning when switching between multiple displays
	a.renderer.ResizeActive()
	a.state.SetGridOverlayNeedsRefresh(false)

	err := a.setupGrid()
	if err != nil {
		a.logger.Error("Failed to setup grid", zap.Error(err), zap.String("action", actionString))
		return
	}

	a.setModeGrid()

	a.logger.Info("Grid mode activated", zap.String("action", actionString))
	a.logger.Info("Type a grid label to select a location")
}

// setupGrid generates grid and draws it.
func (a *App) setupGrid() error {
	// Create grid with active screen bounds (screen containing mouse cursor)
	// This ensures proper multi-monitor support
	gridInstance := a.createGridInstance()

	// Update grid overlay configuration
	a.updateGridOverlayConfig()

	// Ensure the overlay is properly sized for the active screen
	a.overlayManager.ResizeToActiveScreenSync()

	// Reset the grid manager state when setting up the grid
	if a.gridManager != nil {
		a.gridManager.Reset()
	}

	// Initialize manager with the new grid
	a.initializeGridManager(gridInstance)

	a.gridRouter = grid.NewRouter(a.gridManager, a.logger)

	// Draw initial grid
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

	// Normalize bounds to window-local coordinates (0,0 origin)
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	bounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	characters := a.config.Grid.Characters
	if strings.TrimSpace(characters) == "" {
		characters = a.config.Hints.HintCharacters
	}
	gridInstance := grid.NewGrid(characters, bounds, a.logger)
	*a.gridCtx.gridInstance = gridInstance

	return gridInstance
}

// updateGridOverlayConfig updates the grid overlay configuration.
func (a *App) updateGridOverlayConfig() {
	// Grid overlay already created in NewApp - update its config and use it
	(*a.gridCtx.gridOverlay).UpdateConfig(a.config.Grid)
}

// initializeGridManager initializes the grid manager with the new grid instance.
func (a *App) initializeGridManager(gridInstance *grid.Grid) {
	// Subgrid configuration and keys (fallback to grid characters): always 3x3
	keys := strings.TrimSpace(a.config.Grid.SublayerKeys)
	if keys == "" {
		keys = a.config.Grid.Characters
	}
	const subRows = 3
	const subCols = 3

	// Initialize manager with the new grid
	a.gridManager = grid.NewManager(gridInstance, subRows, subCols, keys, func(forceRedraw bool) {
		// Update matches only (no full redraw)
		input := a.gridManager.GetInput()

		// special case to handle only when exiting subgrid
		if forceRedraw {
			a.renderer.Clear()
			gridErr := a.renderer.DrawGrid(gridInstance, input)
			if gridErr != nil {
				return
			}
			a.renderer.Show()
		}

		// Set hideUnmatched based on whether we have input and the config setting
		hideUnmatched := a.config.Grid.HideUnmatched && len(input) > 0
		a.renderer.SetHideUnmatched(hideUnmatched)
		a.renderer.UpdateGridMatches(input)
	}, func(cell *grid.Cell) {
		// Draw 3x3 subgrid inside selected cell
		a.renderer.ShowSubgrid(cell)
	}, a.logger)
}

// drawGridActionHighlight draws a highlight border around the active screen for grid action mode.
func (a *App) drawGridActionHighlight() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	// Get active screen bounds
	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	// Draw action highlight using renderer
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

// cleanupGridMode handles cleanup for grid mode.
func (a *App) cleanupGridMode() {
	// Reset action mode state
	a.gridCtx.inActionMode = false

	if a.gridManager != nil {
		a.gridManager.Reset()
	}
	// Hide overlays
	a.logger.Info("Hiding grid overlay")
	a.overlayManager.Hide()

	// Also clear and hide action overlay
	a.overlayManager.Clear()
	a.overlayManager.Hide()
}

// handleGridActionKey handles action keys when in grid action mode.
func (a *App) handleGridActionKey(key string) {
	a.handleActionKey(key, "Grid")
}
