package app

// ActionName represents the string name of an action.
type ActionName string

const (
	// ActionNameLeftClick is the action name for left click.
	ActionNameLeftClick ActionName = "left_click"
	// ActionNameRightClick is the action name for right click.
	ActionNameRightClick ActionName = "right_click"
	// ActionNameMiddleClick is the action name for middle click.
	ActionNameMiddleClick ActionName = "middle_click"
	// ActionNameMouseDown is the action name for mouse down.
	ActionNameMouseDown ActionName = "mouse_down"
	// ActionNameMouseUp is the action name for mouse up.
	ActionNameMouseUp ActionName = "mouse_up"
	// ActionNameScroll is the action name for scrolling.
	ActionNameScroll ActionName = "scroll"
)

// KnownActionNames returns a slice of all known action names.
func KnownActionNames() []ActionName {
	return []ActionName{
		ActionNameLeftClick,
		ActionNameRightClick,
		ActionNameMiddleClick,
		ActionNameMouseDown,
		ActionNameMouseUp,
		ActionNameScroll,
	}
}

// IsKnownActionName checks if an action name is known.
func IsKnownActionName(action ActionName) bool {
	switch action {
	case ActionNameLeftClick,
		ActionNameRightClick,
		ActionNameMiddleClick,
		ActionNameMouseDown,
		ActionNameMouseUp,
		ActionNameScroll:
		return true
	default:
		return false
	}
}
