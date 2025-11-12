package scroll

import (
	"image"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/hints"
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
	AmountEnd
)

// Controller manages scrolling operations
type Controller struct {
	config *config.ScrollConfig
	logger *zap.Logger
}

// NewController creates a new scroll controller
func NewController(cfg config.ScrollConfig, logger *zap.Logger) *Controller {
	logger.Debug("Scroll controller: Initializing",
		zap.Int("scroll_step", cfg.ScrollStep),
		zap.Int("scroll_step_half", cfg.ScrollStepHalf),
		zap.Int("scroll_step_full", cfg.ScrollStepFull))

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
		zap.String("amount", c.amountString(amount)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	if err := accessibility.ScrollAtCursor(deltaX, deltaY); err != nil {
		c.logger.Error("Scroll failed", zap.Error(err))
		return err
	}

	c.logger.Info("Scroll completed successfully",
		zap.String("direction", c.directionString(dir)),
		zap.String("amount", c.amountString(amount)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	return nil
}

// calculateDelta calculates the scroll delta based on direction and amount
func (c *Controller) calculateDelta(dir Direction, amount ScrollAmount) (int, int) {
	var deltaX, deltaY int

	// Use config values for all scroll amounts
	var baseScroll int

	switch amount {
	case AmountChar:
		// Single scroll (j/k)
		baseScroll = c.config.ScrollStep
	case AmountHalfPage:
		// Half page (Ctrl+D/U)
		baseScroll = c.config.ScrollStepHalf
	case AmountEnd:
		// Top/Bottom (gg/G)
		baseScroll = c.config.ScrollStepFull
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

	c.logger.Debug("Calculated scroll delta",
		zap.String("direction", c.directionString(dir)),
		zap.String("amount", c.amountString(amount)),
		zap.Int("base_scroll", baseScroll),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

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

// amountString converts scroll amount to string
func (c *Controller) amountString(amount ScrollAmount) string {
	switch amount {
	case AmountChar:
		return "character"
	case AmountHalfPage:
		return "half_page"
	case AmountEnd:
		return "end"
	default:
		return "unknown"
	}
}

// ScrollUp scrolls up by character
func (c *Controller) ScrollUp() error {
	c.logger.Debug("ScrollUp called")
	// Use positive delta for up (scroll wheel convention)
	return c.Scroll(DirectionUp, AmountChar)
}

// ScrollDown scrolls down by character
func (c *Controller) ScrollDown() error {
	c.logger.Debug("ScrollDown called")
	// Use negative delta for down (scroll wheel convention)
	return c.Scroll(DirectionDown, AmountChar)
}

// ScrollLeft scrolls left by character
func (c *Controller) ScrollLeft() error {
	c.logger.Debug("ScrollLeft called")
	return c.Scroll(DirectionLeft, AmountChar)
}

// ScrollRight scrolls right by character
func (c *Controller) ScrollRight() error {
	c.logger.Debug("ScrollRight called")
	return c.Scroll(DirectionRight, AmountChar)
}

// ScrollUpHalfPage scrolls up by half page (Ctrl+U in Vim)
func (c *Controller) ScrollUpHalfPage() error {
	c.logger.Debug("ScrollUpHalfPage called")
	return c.Scroll(DirectionUp, AmountHalfPage)
}

// ScrollDownHalfPage scrolls down by half page (Ctrl+D in Vim)
func (c *Controller) ScrollDownHalfPage() error {
	c.logger.Debug("ScrollDownHalfPage called")
	return c.Scroll(DirectionDown, AmountHalfPage)
}

// ScrollToTop scrolls to the top (gg in Vim)
func (c *Controller) ScrollToTop() error {
	c.logger.Debug("ScrollToTop called")
	return c.Scroll(DirectionUp, AmountEnd)
}

// ScrollToBottom scrolls to the bottom (G in Vim)
func (c *Controller) ScrollToBottom() error {
	c.logger.Debug("ScrollToBottom called")
	return c.Scroll(DirectionDown, AmountEnd)
}

// DrawHighlightBorder draws a highlight around the current scroll area
func (c *Controller) DrawHighlightBorder(overlay *hints.Overlay) {
	c.logger.Debug("DrawHighlightBorder called")
	window := accessibility.GetFrontmostWindow()
	if window == nil {
		c.logger.Debug("No frontmost window")
		return
	}
	defer window.Release()

	// Get scroll bounds in absolute screen coordinates
	bounds := window.GetScrollBounds()

	// Get active screen bounds to normalize coordinates
	screenBounds := bridge.GetActiveScreenBounds()

	// Convert absolute bounds to window-local coordinates
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	localBounds := image.Rect(
		bounds.Min.X-screenBounds.Min.X,
		bounds.Min.Y-screenBounds.Min.Y,
		bounds.Max.X-screenBounds.Min.X,
		bounds.Max.Y-screenBounds.Min.Y,
	)

	c.logger.Debug("Drawing scroll highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))

	overlay.DrawScrollHighlight(
		localBounds.Min.X, localBounds.Min.Y,
		localBounds.Dx(), localBounds.Dy(),
		c.config.HighlightColor,
		c.config.HighlightWidth,
	)
}
