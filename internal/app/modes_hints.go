package app

import (
	"fmt"
	"image"
	"runtime"

	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/features/hints"
	"github.com/y3owk1n/neru/internal/infra/accessibility"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"go.uber.org/zap"
)

// activateMode activates a mode with a given action (for hints mode).
func (a *App) activateMode(mode Mode) {
	if mode == ModeIdle {
		// Explicit idle transition
		a.exitMode()
		return
	}
	if mode == ModeHints {
		a.activateHintMode()
		return
	}
	if mode == ModeGrid {
		a.activateGridMode()
		return
	}
	// Unknown or unsupported mode
	a.logger.Warn("Unknown mode", zap.String("mode", getModeString(mode)))
}

// activateHintMode activates hint mode, with an option to preserve action mode state.
func (a *App) activateHintMode() {
	a.activateHintModeInternal(false)
}

// activateHintModeInternal activates hint mode with option to preserve action mode state.
func (a *App) activateHintModeInternal(preserveActionMode bool) {
	if !a.state.IsEnabled() {
		a.logger.Debug("Neru is disabled, ignoring hint mode activation")
		return
	}
	// Respect mode enable flag
	if !a.config.Hints.Enabled {
		a.logger.Debug("Hints mode disabled by config, ignoring activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.captureInitialCursorPosition()

	action := domain.ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating hint mode", zap.String("action", actionString))

	// Only exit mode if not preserving action mode state
	if !preserveActionMode {
		a.exitMode()
	}

	if actionString == unknownAction {
		a.logger.Warn("Unknown action string, ignoring")
		return
	}

	// Always resize overlay to the active screen (where mouse is) before collecting elements
	// This ensures proper positioning when switching between multiple displays
	a.overlayManager.ResizeToActiveScreenSync()
	a.state.SetHintOverlayNeedsRefresh(false)

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Collect elements based on mode
	elements := a.collectElements()
	if len(elements) == 0 {
		a.logger.Warn("No elements found for action", zap.String("action", actionString))
		return
	}

	// Generate and setup hints
	err := a.setupHints(elements)
	if err != nil {
		a.logger.Error("Failed to setup hints", zap.Error(err), zap.String("action", actionString))
		return
	}

	a.hintsCtx.SetSelectedHint(nil)
	a.setModeHints()
}

// setupHints generates hints and draws them with appropriate styling.
func (a *App) setupHints(elements []*accessibility.TreeNode) error {
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)

	// Generate and normalize hints
	hintList, err := a.generateAndNormalizeHints(elements)
	if err != nil {
		return err
	}

	// Set up hints in the hint manager
	hintCollection := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hintCollection)

	// Draw hints with mode-specific styling
	drawErr := a.renderer.DrawHints(hintList)
	if drawErr != nil {
		return fmt.Errorf("failed to draw hints: %w", drawErr)
	}
	a.renderer.Show()

	// Log performance metrics
	a.logHintsSetupPerformance(hintList, msBefore)

	return nil
}

// generateAndNormalizeHints generates hints and normalizes their positions.
func (a *App) generateAndNormalizeHints(elements []*accessibility.TreeNode) ([]*hints.Hint, error) {
	// Get active screen bounds to calculate offset for normalization
	screenBounds := bridge.GetActiveScreenBounds()
	screenOffsetX := screenBounds.Min.X
	screenOffsetY := screenBounds.Min.Y

	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hints: %w", err)
	}

	// Normalize hint positions to window-local coordinates
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	for _, hint := range hintList {
		pos := hint.GetPosition()
		hint.Position.X = pos.X - screenOffsetX
		hint.Position.Y = pos.Y - screenOffsetY
	}

	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())
	filtered := make([]*hints.Hint, 0, len(hintList))
	for _, h := range hintList {
		if h.IsVisible(localBounds) {
			filtered = append(filtered, h)
		}
	}

	return filtered, nil
}

// logHintsSetupPerformance logs performance metrics for hint setup.
func (a *App) logHintsSetupPerformance(hintList []*hints.Hint, msBefore runtime.MemStats) {
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	a.logger.Info("Hints setup perf",
		zap.Int("hints", len(hintList)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))
}

// drawHintsActionHighlight draws a highlight border around the active screen for hints action mode.
func (a *App) drawHintsActionHighlight() {
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

	a.logger.Debug("Drawing hints action highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))
}

// handleHintsActionKey handles action keys when in hints action mode.
func (a *App) handleHintsActionKey(key string) {
	a.handleActionKey(key, "Hints")
}
