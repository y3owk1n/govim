package domain

// ActionName represents a named action that can be performed by the application.
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

// KnownActionNames returns a slice containing all supported action names.
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

// IsKnownActionName determines whether the specified action name is supported.
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

// Action is the current action of the application.
type Action int

const (
	// ActionLeftClick is the action for left click.
	ActionLeftClick Action = iota
	// ActionRightClick is the action for right click.
	ActionRightClick
	// ActionMouseUp is the action for mouse up.
	ActionMouseUp
	// ActionMouseDown is the action for mouse down.
	ActionMouseDown
	// ActionMiddleClick is the action for middle click.
	ActionMiddleClick
	// ActionMoveMouse is the action for moving the mouse.
	ActionMoveMouse
	// ActionScroll is the action for scrolling.
	ActionScroll
)
