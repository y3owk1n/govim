package hints

import "go.uber.org/zap"

// Router centralizes hint-mode key routing that is specific to hints.
// It parses keys like Escape/Tab and delegates incremental label input
// to the Manager, returning a concise result for the caller (App) to act on.

type Router struct {
	manager *Manager
	logger  *zap.Logger
}

// KeyResult captures routing decisions for a single keypress.
// The App should interpret these flags and perform UI/actions.
type KeyResult struct {
	Exit           bool  // Escape pressed -> exit mode
	SwitchToScroll bool  // Tab pressed when in scroll action but not yet scrolling
	ExactHint      *Hint // Exact label match selected
}

func NewRouter(m *Manager, logger *zap.Logger) *Router {
	return &Router{
		manager: m,
		logger:  logger,
	}
}

// RouteKey processes a keypress for hints.
// - selectedHintPresent: whether a hint is already selected (context menu active)
// - canScroll: whether we are already in active scroll sub-state
// - isScrollAction: whether current action is the scroll action
// RouteKey does NOT execute actions; it only returns routing decisions and exact match.
func (r *Router) RouteKey(key string, selectedHintPresent bool, canScroll bool, isScrollAction bool) KeyResult {
	var res KeyResult

	r.logger.Debug("Hints router processing key",
		zap.String("key", key),
		zap.Bool("selected_hint_present", selectedHintPresent),
		zap.Bool("can_scroll", canScroll),
		zap.Bool("is_scroll_action", isScrollAction))

	// Exit any mode with Escape
	if key == "\x1b" || key == "escape" {
		r.logger.Debug("Hints router: Exit key pressed")
		res.Exit = true
		return res
	}

	// Tab: when current action is scroll but not yet scrolling, switch to scroll sub-state
	if key == "\t" && isScrollAction && !canScroll {
		r.logger.Debug("Hints router: Tab pressed for scroll action")
		res.SwitchToScroll = true
		return res
	}

	// If context menu is active or scroll is active, App will handle separately
	if selectedHintPresent || canScroll {
		r.logger.Debug("Hints router: Context menu or scroll active, skipping hint processing")
		return res
	}

	// Delegate label input to the hint manager
	if hint, ok := r.manager.HandleInput(key); ok {
		r.logger.Debug("Hints router: Exact hint match found", zap.String("label", hint.Label))
		res.ExactHint = hint
		return res
	}

	return res
}

// ParseContextMenuKey maps context menu key to action name (string).
// Returns empty string if unknown.
func ParseContextMenuKey(key string) string {
	switch key {
	case "l":
		return "left_click"
	case "r":
		return "right_click"
	case "d":
		return "double_click"
	case "t":
		return "triple_click"
	case "h":
		return "mouse_down"
	case "H":
		return "mouse_up"
	case "m":
		return "middle_click"
	case "g":
		return "move_mouse"
	default:
		return ""
	}
}

// ContextMenuItems returns key/label/action entries for the context menu
func ContextMenuItems() []struct{ Key, Label, Action string } {
	return []struct{ Key, Label, Action string }{
		{Key: "l", Label: "left", Action: "left_click"},
		{Key: "r", Label: "right", Action: "right_click"},
		{Key: "d", Label: "double", Action: "double_click"},
		{Key: "t", Label: "triple", Action: "triple_click"},
		{Key: "h", Label: "hold", Action: "mouse_down"},
		{Key: "H", Label: " unhold", Action: "mouse_up"},
		{Key: "m", Label: "middle", Action: "middle_click"},
		{Key: "g", Label: "oto pos", Action: "move_mouse"},
	}
}
