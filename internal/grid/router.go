package grid

import "image"

// Router handles grid mode key routing
type Router struct {
	manager *Manager
}

// KeyResult captures routing decisions for grid mode
type KeyResult struct {
	Exit           bool        // Escape pressed -> exit mode
	SwitchToScroll bool        // Tab pressed when in scroll action but not yet scrolling
	TargetPoint    image.Point // Complete coordinate entered
	Complete       bool        // Coordinate selection complete
}

func NewRouter(m *Manager) *Router {
	return &Router{manager: m}
}

// RouteKey processes a keypress for grid mode
func (r *Router) RouteKey(key string, canScroll bool, isScrollAction bool) KeyResult {
	var res KeyResult

	// Exit grid mode with Escape
	if key == "\x1b" || key == "escape" {
		res.Exit = true
		return res
	}

	// Tab: when current action is scroll but not yet scrolling, switch to scroll sub-state
	if key == "\t" && isScrollAction && !canScroll {
		res.SwitchToScroll = true
		return res
	}

	// Delegate coordinate input to the grid manager
	if point, complete := r.manager.HandleInput(key); complete {
		res.TargetPoint = point
		res.Complete = true
	}

	return res
}
