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
	Exit      bool  // Escape pressed -> exit mode
	ExactHint *Hint // Exact label match selected
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
func (r *Router) RouteKey(key string, selectedHintPresent bool) KeyResult {
	var res KeyResult

	r.logger.Debug("Hints router processing key",
		zap.String("key", key),
		zap.Bool("selected_hint_present", selectedHintPresent),
	)

	// Exit any mode with Escape
	if key == "\x1b" || key == "escape" {
		r.logger.Debug("Hints router: Exit key pressed")
		res.Exit = true
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
