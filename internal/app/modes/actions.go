package modes

import (
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"github.com/y3owk1n/neru/internal/ui/coordinates"
	"go.uber.org/zap"
)

// drawActionHighlight draws a highlight border around the active screen for action mode.
// This is used by both hints and grid modes when in action mode.
func (h *Handler) drawActionHighlight() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	h.Renderer.ResizeActive()

	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := coordinates.NormalizeToLocalCoordinates(screenBounds)

	h.Renderer.DrawActionHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)

	h.Logger.Debug("Drawing action highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))
}
