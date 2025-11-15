package grid

import (
	"image"

	"go.uber.org/zap"
)

// Router handles grid mode key routing
type Router struct {
	manager *Manager
	logger  *zap.Logger
}

// KeyResult captures routing decisions for grid mode
type KeyResult struct {
	Exit        bool        // Escape pressed -> exit mode
	TargetPoint image.Point // Complete coordinate entered
	Complete    bool        // Coordinate selection complete
}

// NewRouter creates a new grid router.
func NewRouter(m *Manager, logger *zap.Logger) *Router {
	return &Router{
		manager: m,
		logger:  logger,
	}
}

// RouteKey processes a keypress for grid mode
func (r *Router) RouteKey(key string) KeyResult {
	var res KeyResult

	r.logger.Debug("Grid router processing key",
		zap.String("key", key),
	)

	// Exit grid mode with Escape
	if key == "\x1b" || key == "escape" {
		r.logger.Debug("Grid router: Exit key pressed")
		res.Exit = true
		return res
	}

	// Delegate coordinate input to the grid manager
	if point, complete := r.manager.HandleInput(key); complete {
		r.logger.Debug("Grid router: Coordinate selection complete",
			zap.Int("x", point.X),
			zap.Int("y", point.Y))
		res.TargetPoint = point
		res.Complete = true
	}

	return res
}
