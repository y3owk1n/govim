package hints

import "go.uber.org/zap"

// Router centralizes hint-mode key routing that is specific to hints.
// It parses keys like Escape/Tab and delegates incremental label input
// to the Manager, returning a concise result for the caller (App) to act on.

// Router handles key routing for hint mode operations.
type Router struct {
	manager *Manager
	logger  *zap.Logger
}

// KeyResult captures the results of key routing decisions in hint mode.
// The application interprets these flags to perform appropriate UI actions.
type KeyResult struct {
	Exit      bool  // Escape pressed -> exit mode
	ExactHint *Hint // Exact label match selected
}

// NewRouter initializes a new hints router with the specified manager and logger.
func NewRouter(m *Manager, logger *zap.Logger) *Router {
	return &Router{
		manager: m,
		logger:  logger,
	}
}

// RouteKey processes a keypress and determines the appropriate action in hint mode.
// RouteKey does NOT execute actions; it only returns routing decisions and exact matches.
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
		r.logger.Debug("Hints router: Exact hint match found", zap.String("label", hint.GetLabel()))
		res.ExactHint = hint
		return res
	}

	return res
}
