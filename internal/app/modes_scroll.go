package app

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/features/scroll"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"github.com/y3owk1n/neru/internal/ui/overlay"
	"go.uber.org/zap"
)

// handleGenericScrollKey handles scroll keys in a generic way.
func (a *App) handleGenericScrollKey(key string, lastScrollKey *string) {
	var localLastKey string
	if lastScrollKey == nil {
		lastScrollKey = &localLastKey
	}

	bytes := []byte(key)
	a.logger.Info("Scroll key pressed",
		zap.String("key", key),
		zap.Int("len", len(key)),
		zap.String("hex", fmt.Sprintf("%#v", key)),
		zap.Any("bytes", bytes))

	var err error

	if len(key) == 1 {
		if a.handleControlScrollKey(key, *lastScrollKey, lastScrollKey) {
			return
		}
	}

	a.logger.Debug(
		"Entering switch statement",
		zap.String("key", key),
		zap.String("keyHex", fmt.Sprintf("%#v", key)),
	)
	switch key {
	case "j", "k", "h", "l":
		err = a.handleDirectionalScrollKey(key, *lastScrollKey)
	case "g":
		operation, newLast, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			a.logger.Info("First g pressed, press again for top")
			*lastScrollKey = newLast
			return
		}
		if operation == "top" {
			a.logger.Info("gg detected - scroll to top")
			err = a.ScrollController().ScrollToTop()
			*lastScrollKey = ""
			goto done
		}
	case "G":
		operation, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if ok && operation == "bottom" {
			a.logger.Info("G key detected - scroll to bottom")
			err = a.ScrollController().ScrollToBottom()
			*lastScrollKey = ""
		}
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		*lastScrollKey = ""
		return
	}

	*lastScrollKey = ""

done:
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
}

// handleControlScrollKey handles control character scroll keys.
func (a *App) handleControlScrollKey(key string, lastKey string, lastScrollKey *string) bool {
	byteVal := key[0]
	a.logger.Info("Checking control char", zap.Uint8("byte", byteVal))
	// Only handle Ctrl+D / Ctrl+U here; let Tab (9) and other keys fall through to switch
	if byteVal == 4 || byteVal == 21 {
		op, _, ok := scroll.ParseKey(key, lastKey, a.logger)
		if ok {
			*lastScrollKey = ""
			switch op {
			case "half_down":
				a.logger.Info("Ctrl+D detected - half page down")
				return a.ScrollController().ScrollDownHalfPage() != nil
			case "half_up":
				a.logger.Info("Ctrl+U detected - half page up")
				return a.ScrollController().ScrollUpHalfPage() != nil
			}
		}
	}
	return false
}

// handleDirectionalScrollKey handles directional scroll keys (j, k, h, l).
func (a *App) handleDirectionalScrollKey(key string, lastKey string) error {
	op, _, ok := scroll.ParseKey(key, lastKey, a.logger)
	if !ok {
		return nil
	}

	switch key {
	case "j":
		if op == "down" {
			return a.ScrollController().ScrollDown()
		}
	case "k":
		if op == "up" {
			return a.ScrollController().ScrollUp()
		}
	case "h":
		if op == "left" {
			return a.ScrollController().ScrollLeft()
		}
	case "l":
		if op == "right" {
			return a.ScrollController().ScrollRight()
		}
	}
	return nil
}

func (a *App) drawScrollHighlightBorder() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	a.renderer.DrawScrollHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)
	a.overlayManager.SwitchTo(overlay.ModeScroll)
}
