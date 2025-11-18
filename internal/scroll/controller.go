// Package scroll provides scrolling functionality for the Neru application.
package scroll

import (
	"fmt"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

// Direction represents the direction of a scrolling operation.
type Direction int

const (
	// DirectionUp represents upward scrolling.
	DirectionUp Direction = iota
	// DirectionDown represents downward scrolling.
	DirectionDown
	// DirectionLeft represents leftward scrolling.
	DirectionLeft
	// DirectionRight represents rightward scrolling.
	DirectionRight
)

// Amount represents the magnitude of a scrolling operation.
type Amount int

const (
	// AmountChar represents character-level scrolling.
	AmountChar Amount = iota
	// AmountHalfPage represents half-page scrolling.
	AmountHalfPage
	// AmountEnd represents scrolling to the end.
	AmountEnd
)

// Controller orchestrates scrolling operations using accessibility APIs.
type Controller struct {
	config *config.ScrollConfig
	logger *zap.Logger
}

// NewController initializes a new scroll controller with the specified configuration.
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

// Scroll performs a scrolling operation in the specified direction and magnitude.
func (c *Controller) Scroll(dir Direction, amount Amount) error {
	deltaX, deltaY := c.calculateDelta(dir, amount)

	c.logger.Debug("Scrolling",
		zap.String("direction", c.directionString(dir)),
		zap.String("amount", c.amountString(amount)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	err := accessibility.ScrollAtCursor(deltaX, deltaY)
	if err != nil {
		c.logger.Error("Scroll failed", zap.Error(err))
		return fmt.Errorf("failed to scroll at cursor: %w", err)
	}

	c.logger.Info("Scroll completed successfully",
		zap.String("direction", c.directionString(dir)),
		zap.String("amount", c.amountString(amount)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	return nil
}

// ScrollUp performs a character-level upward scroll.
func (c *Controller) ScrollUp() error {
	c.logger.Debug("ScrollUp called")
	// Use positive delta for up (scroll wheel convention)
	return c.Scroll(DirectionUp, AmountChar)
}

// ScrollDown performs a character-level downward scroll.
func (c *Controller) ScrollDown() error {
	c.logger.Debug("ScrollDown called")
	// Use negative delta for down (scroll wheel convention)
	return c.Scroll(DirectionDown, AmountChar)
}

// ScrollLeft performs a character-level leftward scroll.
func (c *Controller) ScrollLeft() error {
	c.logger.Debug("ScrollLeft called")
	return c.Scroll(DirectionLeft, AmountChar)
}

// ScrollRight performs a character-level rightward scroll.
func (c *Controller) ScrollRight() error {
	c.logger.Debug("ScrollRight called")
	return c.Scroll(DirectionRight, AmountChar)
}

// ScrollUpHalfPage performs a half-page upward scroll (equivalent to Ctrl+U in Vim).
func (c *Controller) ScrollUpHalfPage() error {
	c.logger.Debug("ScrollUpHalfPage called")
	return c.Scroll(DirectionUp, AmountHalfPage)
}

// ScrollDownHalfPage performs a half-page downward scroll (equivalent to Ctrl+D in Vim).
func (c *Controller) ScrollDownHalfPage() error {
	c.logger.Debug("ScrollDownHalfPage called")
	return c.Scroll(DirectionDown, AmountHalfPage)
}

// ScrollToTop scrolls to the top of the document (equivalent to gg in Vim).
func (c *Controller) ScrollToTop() error {
	c.logger.Debug("ScrollToTop called")
	return c.Scroll(DirectionUp, AmountEnd)
}

// ScrollToBottom scrolls to the bottom of the document (equivalent to G in Vim).
func (c *Controller) ScrollToBottom() error {
	c.logger.Debug("ScrollToBottom called")
	return c.Scroll(DirectionDown, AmountEnd)
}

// calculateDelta computes the scroll delta values based on direction and magnitude.
func (c *Controller) calculateDelta(dir Direction, amount Amount) (int, int) {
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

// directionString converts a Direction value to its string representation.
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

// amountString converts an Amount value to its string representation.
func (c *Controller) amountString(amount Amount) string {
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
