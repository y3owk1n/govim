package scroll

import (
	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

// Direction represents scroll direction
type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
)

// ScrollAmount represents the amount to scroll
type ScrollAmount int

const (
	AmountChar ScrollAmount = iota
	AmountHalfPage
	AmountFullPage
	AmountEnd
)

// Controller manages scrolling operations
type Controller struct {
	config *config.ScrollConfig
	logger *zap.Logger
}

// NewController creates a new scroll controller
func NewController(cfg config.ScrollConfig, logger *zap.Logger) *Controller {
	return &Controller{
		config: &cfg,
		logger: logger,
	}
}

// Scroll scrolls in the specified direction
func (c *Controller) Scroll(dir Direction, amount ScrollAmount) error {
	deltaX, deltaY := c.calculateDelta(dir, amount)

	c.logger.Debug("Scrolling",
		zap.String("direction", c.directionString(dir)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	if err := accessibility.ScrollAtCursor(deltaX, deltaY); err != nil {
		c.logger.Error("Scroll failed", zap.Error(err))
		return err
	}

	return nil
}

// calculateDelta calculates the scroll delta based on direction and amount
func (c *Controller) calculateDelta(dir Direction, amount ScrollAmount) (int, int) {
	var deltaX, deltaY int

	// Use config values for all scroll amounts
	var baseScroll int

	switch amount {
	case AmountChar:
		// Single scroll (j/k) - use scroll_speed from config
		baseScroll = c.config.ScrollSpeed
	case AmountHalfPage:
		// Half page (Ctrl+D/U) - use page_height * half_page_multiplier
		baseScroll = int(float64(c.config.PageHeight) * c.config.HalfPageMultiplier)
	case AmountFullPage:
		// Full page - use page_height * full_page_multiplier
		baseScroll = int(float64(c.config.PageHeight) * c.config.FullPageMultiplier)
	case AmountEnd:
		// Top/Bottom - use page_height * end_multiplier
		baseScroll = int(c.config.ScrollToEdgeDelta)
	}

	// Note: For CGEvent scroll wheel, positive = up/left, negative = down/right
	switch dir {
	case DirectionUp:
		deltaY = baseScroll // Positive for up
	case DirectionDown:
		deltaY = -baseScroll // Negative for down
	case DirectionLeft:
		deltaX = baseScroll // Positive for left
	case DirectionRight:
		deltaX = -baseScroll // Negative for right
	}

	return deltaX, deltaY
}

// directionString converts direction to string
func (c *Controller) directionString(dir Direction) string {
	switch dir {
	case DirectionUp:
		return "up"
	case DirectionDown:
		return "down"
	case DirectionLeft:
		return "left"
	case DirectionRight:
		return "right"
	default:
		return "unknown"
	}
}

// ScrollUp scrolls up by character
func (c *Controller) ScrollUp() error {
	// Use positive delta for up (scroll wheel convention)
	return c.Scroll(DirectionUp, AmountChar)
}

// ScrollDown scrolls down by character
func (c *Controller) ScrollDown() error {
	// Use negative delta for down (scroll wheel convention)
	return c.Scroll(DirectionDown, AmountChar)
}

// ScrollLeft scrolls left by character
func (c *Controller) ScrollLeft() error {
	return c.Scroll(DirectionLeft, AmountChar)
}

// ScrollRight scrolls right by character
func (c *Controller) ScrollRight() error {
	return c.Scroll(DirectionRight, AmountChar)
}

// ScrollUpHalfPage scrolls up by half page (Ctrl+U in Vim)
func (c *Controller) ScrollUpHalfPage() error {
	return c.Scroll(DirectionUp, AmountHalfPage)
}

// ScrollDownHalfPage scrolls down by half page (Ctrl+D in Vim)
func (c *Controller) ScrollDownHalfPage() error {
	return c.Scroll(DirectionDown, AmountHalfPage)
}

// ScrollToTop scrolls to the top (gg in Vim)
func (c *Controller) ScrollToTop() error {
	return c.Scroll(DirectionUp, AmountEnd)
}

// ScrollToBottom scrolls to the bottom (G in Vim)
func (c *Controller) ScrollToBottom() error {
	return c.Scroll(DirectionDown, AmountEnd)
}

// ScrollUpFullPage scrolls up by full page
func (c *Controller) ScrollUpFullPage() error {
	return c.Scroll(DirectionUp, AmountFullPage)
}

// ScrollDownFullPage scrolls down by full page
func (c *Controller) ScrollDownFullPage() error {
	return c.Scroll(DirectionDown, AmountFullPage)
}
