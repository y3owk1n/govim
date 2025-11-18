package app

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/accessibility"
)

// performActionAtPoint executes the specified action at the given point.
func performActionAtPoint(action string, pt image.Point) error {
	actionName := ActionName(action)
	switch actionName {
	case ActionNameLeftClick:
		return accessibility.LeftClickAtPoint(pt, false)
	case ActionNameRightClick:
		return accessibility.RightClickAtPoint(pt, false)
	case ActionNameMiddleClick:
		return accessibility.MiddleClickAtPoint(pt, false)
	case ActionNameMouseDown:
		return accessibility.LeftMouseDownAtPoint(pt)
	case ActionNameMouseUp:
		return accessibility.LeftMouseUpAtPoint(pt)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// isKnownAction checks if the specified action is a known action type.
func isKnownAction(action string) bool {
	return IsKnownActionName(ActionName(action))
}

// startInteractiveScroll initiates interactive scrolling mode with visual feedback.
func (a *App) startInteractiveScroll() {
	a.skipCursorRestoreOnce = true
	a.exitMode()

	if a.overlayManager != nil {
		a.overlayManager.ResizeToActiveScreenSync()
	}

	if a.config.Scroll.HighlightScrollArea {
		a.drawScrollHighlightBorder()
		if a.overlayManager != nil {
			a.overlayManager.Show()
		}
	}

	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.isScrollingActive = true

	a.logger.Info("Interactive scroll activated")
	a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom, Esc to exit")
}
