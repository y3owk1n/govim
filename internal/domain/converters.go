package domain

// GetModeString converts a Mode to its string representation.
func GetModeString(mode Mode) string {
	switch mode {
	case ModeIdle:
		return "idle"
	case ModeHints:
		return "hints"
	case ModeGrid:
		return "grid"
	default:
		return UnknownMode
	}
}

// GetActionString converts an Action to its string representation.
func GetActionString(action Action) string {
	switch action {
	case ActionLeftClick:
		return "left_click"
	case ActionRightClick:
		return "right_click"
	case ActionMouseUp:
		return "mouse_up"
	case ActionMouseDown:
		return "mouse_down"
	case ActionMiddleClick:
		return "middle_click"
	case ActionMoveMouse:
		return "move_mouse"
	case ActionScroll:
		return "scroll"
	default:
		return UnknownAction
	}
}
