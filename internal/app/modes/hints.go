package modes

import (
	"fmt"
	"image"
	"runtime"

	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/features/hints"
	infra "github.com/y3owk1n/neru/internal/infra/accessibility"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"go.uber.org/zap"
)

// ActivateMode activates a mode with a given action (for hints mode).
func (h *Handler) ActivateMode(mode domain.Mode) {
	if mode == domain.ModeIdle {
		h.ExitMode()
		return
	}
	if mode == domain.ModeHints {
		h.activateHintMode()
		return
	}
	if mode == domain.ModeGrid {
		h.activateGridMode()
		return
	}
	h.Logger.Warn("Unknown mode", zap.String("mode", domain.GetModeString(mode)))
}

// activateHintMode activates hint mode, with an option to preserve action mode state.
func (h *Handler) activateHintMode() {
	h.activateHintModeInternal(false)
}

// activateHintModeInternal activates hint mode with option to preserve action mode state.
func (h *Handler) activateHintModeInternal(preserveActionMode bool) {
	// Validate mode activation
	err := h.validateModeActivation("hints", h.Config.Hints.Enabled)
	if err != nil {
		return
	}

	// Prepare for mode activation (reset scroll, capture cursor)
	h.prepareForModeActivation()

	action := domain.ActionMoveMouse
	actionString := domain.GetActionString(action)
	h.Logger.Info("Activating hint mode", zap.String("action", actionString))

	if !preserveActionMode {
		h.ExitMode()
	}

	if actionString == domain.UnknownAction {
		h.Logger.Warn("Unknown action string, ignoring")
		return
	}

	// Always resize overlay to the active screen (where mouse is) before collecting elements.
	// This ensures proper positioning when switching between multiple displays.
	h.OverlayManager.ResizeToActiveScreenSync()
	h.State.SetHintOverlayNeedsRefresh(false)

	h.Accessibility.UpdateRolesForCurrentApp()

	elements := h.Accessibility.CollectElements()
	if len(elements) == 0 {
		h.Logger.Warn("No elements found for action", zap.String("action", actionString))
		return
	}

	err = h.SetupHints(elements)
	if err != nil {
		h.Logger.Error("Failed to setup hints", zap.Error(err), zap.String("action", actionString))
		return
	}

	h.Hints.Context.SetSelectedHint(nil)
	h.SetModeHints()
}

// SetupHints generates hints and draws them with appropriate styling.
func (h *Handler) SetupHints(elements []*infra.TreeNode) error {
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)

	hintList, err := h.generateAndNormalizeHints(elements)
	if err != nil {
		return err
	}

	hintCollection := hints.NewHintCollection(hintList)
	h.Hints.Manager.SetHints(hintCollection)

	drawErr := h.Renderer.DrawHints(hintList)
	if drawErr != nil {
		return fmt.Errorf("failed to draw hints: %w", drawErr)
	}
	h.Renderer.Show()

	h.logHintsSetupPerformance(hintList, msBefore)

	return nil
}

// generateAndNormalizeHints generates hints and normalizes their positions.
func (h *Handler) generateAndNormalizeHints(elements []*infra.TreeNode) ([]*hints.Hint, error) {
	// Get active screen bounds to calculate offset for normalization
	screenBounds := bridge.GetActiveScreenBounds()
	screenOffsetX := screenBounds.Min.X
	screenOffsetY := screenBounds.Min.Y

	hintList, err := h.Hints.Generator.Generate(elements)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hints: %w", err)
	}

	// Check if we have any hints
	if len(hintList) == 0 {
		h.Logger.Warn("No hints generated",
			zap.Int("elements_count", len(elements)),
			zap.String("screen_bounds", fmt.Sprintf("%+v", screenBounds)))
		return hintList, nil
	}

	// Normalize hint positions to window-local coordinates.
	// The overlay window is positioned at the screen origin, but the view uses local coordinates.
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

	h.Logger.Debug("Hints generated and normalized",
		zap.Int("generated", len(hintList)),
		zap.Int("visible", len(filtered)),
		zap.Int("elements", len(elements)))

	return filtered, nil
}

// logHintsSetupPerformance logs performance metrics for hint setup.
func (h *Handler) logHintsSetupPerformance(hintList []*hints.Hint, msBefore runtime.MemStats) {
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	h.Logger.Info("Hints setup perf",
		zap.Int("hints", len(hintList)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))
}

// handleHintsActionKey handles action keys when in hints action mode.
func (h *Handler) handleHintsActionKey(key string) {
	h.handleActionKey(key, "Hints")
}
