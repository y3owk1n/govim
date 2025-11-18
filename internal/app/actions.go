package app

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/accessibility"
)

const (
	actionNameLeftClick   = "left_click"
	actionNameRightClick  = "right_click"
	actionNameMiddleClick = "middle_click"
	actionNameMouseDown   = "mouse_down"
	actionNameMouseUp     = "mouse_up"
)

func performActionAtPoint(action string, pt image.Point) error {
	switch action {
	case actionNameLeftClick:
		return accessibility.LeftClickAtPoint(pt, false)
	case actionNameRightClick:
		return accessibility.RightClickAtPoint(pt, false)
	case actionNameMiddleClick:
		return accessibility.MiddleClickAtPoint(pt, false)
	case actionNameMouseDown:
		return accessibility.LeftMouseDownAtPoint(pt)
	case actionNameMouseUp:
		return accessibility.LeftMouseUpAtPoint(pt)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func isKnownAction(action string) bool {
	switch action {
	case actionNameLeftClick,
		actionNameRightClick,
		actionNameMiddleClick,
		actionNameMouseDown,
		actionNameMouseUp:
		return true
	default:
		return false
	}
}
