package modes

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// validateModeActivation performs common validation checks before mode activation.
// Returns an error if the mode cannot be activated.
func (h *Handler) validateModeActivation(modeName string, modeEnabled bool) error {
	if !h.State.IsEnabled() {
		h.Logger.Debug("Neru is disabled, ignoring mode activation",
			zap.String("mode", modeName))
		return errors.New("neru is disabled")
	}

	if !modeEnabled {
		h.Logger.Debug("Mode disabled by config, ignoring activation",
			zap.String("mode", modeName))
		return fmt.Errorf("mode %s is disabled", modeName)
	}

	if h.Accessibility.IsFocusedAppExcluded() {
		// isFocusedAppExcluded already logs the exclusion
		return errors.New("focused app is excluded")
	}

	return nil
}

// prepareForModeActivation performs common preparation steps before activating a mode.
// This includes resetting scroll context and capturing the initial cursor position.
func (h *Handler) prepareForModeActivation() {
	h.resetScrollContext()
	h.CaptureInitialCursorPosition()
}

// resetScrollContext resets scroll-related state to ensure clean mode transitions.
func (h *Handler) resetScrollContext() {
	if h.Scroll.Context.GetIsActive() {
		// Reset scroll context to ensure clean transition
		h.Scroll.Context.SetIsActive(false)
		h.Scroll.Context.SetLastKey("")
		// Also reset the skip restore flag since we're transitioning from scroll mode
		h.Cursor.Reset()
	}
}
