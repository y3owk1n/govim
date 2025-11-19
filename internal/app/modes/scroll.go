package modes

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/features/scroll"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"github.com/y3owk1n/neru/internal/ui/overlay"
	"go.uber.org/zap"
)

// StartInteractiveScroll initiates interactive scrolling mode with visual feedback.
func (h *Handler) StartInteractiveScroll() {
	h.Cursor.SkipNextRestore()
	// Reset scroll context before exiting current mode to ensure clean state transition
	h.Scroll.Context.SetIsActive(false)
	h.Scroll.Context.SetLastKey("")
	h.ExitMode()

	h.DrawScrollHighlightBorder()
	if h.OverlayManager != nil {
		h.OverlayManager.Show()
	}

	if h.EnableEventTap != nil {
		h.EnableEventTap()
	}

	h.Scroll.Context.SetIsActive(true)

	h.Logger.Info("Interactive scroll activated")
	h.Logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom, Esc to exit")
}

// handleGenericScrollKey handles scroll keys in a generic way.
func (h *Handler) handleGenericScrollKey(key string, lastScrollKey *string) {
	var localLastKey string
	if lastScrollKey == nil {
		lastScrollKey = &localLastKey
	}

	bytes := []byte(key)
	h.Logger.Info("Scroll key pressed",
		zap.String("key", key),
		zap.Int("len", len(key)),
		zap.String("hex", fmt.Sprintf("%#v", key)),
		zap.Any("bytes", bytes))

	var err error

	if len(key) == 1 {
		if h.handleControlScrollKey(key, *lastScrollKey, lastScrollKey) {
			return
		}
	}

	h.Logger.Debug(
		"Entering switch statement",
		zap.String("key", key),
		zap.String("keyHex", fmt.Sprintf("%#v", key)),
	)
	switch key {
	case "j", "k", "h", "l":
		err = h.handleDirectionalScrollKey(key, *lastScrollKey)
	case "g":
		operation, newLast, ok := scroll.ParseKey(key, *lastScrollKey, h.Logger)
		if !ok {
			h.Logger.Info("First g pressed, press again for top")
			*lastScrollKey = newLast
			return
		}
		if operation == "top" {
			h.Logger.Info("gg detected - scroll to top")
			err = h.Scroll.Controller.ScrollToTop()
			*lastScrollKey = ""
			goto done
		}
	case "G":
		operation, _, ok := scroll.ParseKey(key, *lastScrollKey, h.Logger)
		if ok && operation == "bottom" {
			h.Logger.Info("G key detected - scroll to bottom")
			err = h.Scroll.Controller.ScrollToBottom()
			*lastScrollKey = ""
		}
	default:
		h.Logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		*lastScrollKey = ""
		return
	}

	*lastScrollKey = ""

done:
	if err != nil {
		h.Logger.Error("Scroll failed", zap.Error(err))
	}
}

// handleControlScrollKey handles control character scroll keys.
func (h *Handler) handleControlScrollKey(key string, lastKey string, lastScrollKey *string) bool {
	byteVal := key[0]
	h.Logger.Info("Checking control char", zap.Uint8("byte", byteVal))
	// Only handle Ctrl+D / Ctrl+U here; let Tab (9) and other keys fall through to switch
	if byteVal == 4 || byteVal == 21 {
		op, _, ok := scroll.ParseKey(key, lastKey, h.Logger)
		if ok {
			*lastScrollKey = ""
			switch op {
			case "half_down":
				h.Logger.Info("Ctrl+D detected - half page down")
				return h.Scroll.Controller.ScrollDownHalfPage() != nil
			case "half_up":
				h.Logger.Info("Ctrl+U detected - half page up")
				return h.Scroll.Controller.ScrollUpHalfPage() != nil
			}
		}
	}
	return false
}

// handleDirectionalScrollKey handles directional scroll keys (j, k, h, l).
func (h *Handler) handleDirectionalScrollKey(key string, lastKey string) error {
	op, _, ok := scroll.ParseKey(key, lastKey, h.Logger)
	if !ok {
		return nil
	}

	switch key {
	case "j":
		if op == "down" {
			return h.Scroll.Controller.ScrollDown()
		}
	case "k":
		if op == "up" {
			return h.Scroll.Controller.ScrollUp()
		}
	case "h":
		if op == "left" {
			return h.Scroll.Controller.ScrollLeft()
		}
	case "l":
		if op == "right" {
			return h.Scroll.Controller.ScrollRight()
		}
	}
	return nil
}

// DrawScrollHighlightBorder renders the highlight border for scroll mode on the active screen.
// This function resizes the overlay to the active screen for multi-monitor support,
// calculates the screen bounds, and draws the scroll highlight border using the renderer.
func (h *Handler) DrawScrollHighlightBorder() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	h.Renderer.ResizeActive()

	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	h.Renderer.DrawScrollHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)
	h.OverlayManager.SwitchTo(overlay.ModeScroll)
}
