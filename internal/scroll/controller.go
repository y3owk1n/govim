package scroll

import (
	"fmt"
	"time"

	"github.com/y3owk1n/govim/internal/config"
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
)

// Controller controls scrolling behavior
type Controller struct {
	detector      *Detector
	config        config.ScrollConfig
	logger        *zap.Logger
	smoothScroll  bool
	scrollSpeed   int
}

// NewController creates a new scroll controller
func NewController(cfg config.ScrollConfig, logger *zap.Logger) *Controller {
	return &Controller{
		detector:     NewDetector(),
		config:       cfg,
		logger:       logger,
		smoothScroll: cfg.SmoothScroll,
		scrollSpeed:  cfg.ScrollSpeed,
	}
}

// Initialize initializes the scroll controller
func (c *Controller) Initialize() error {
	areas, err := c.detector.DetectScrollAreas()
	if err != nil {
		return fmt.Errorf("failed to detect scroll areas: %w", err)
	}

	c.logger.Info("Detected scroll areas", zap.Int("count", len(areas)))

	// Activate the largest area by default
	if len(areas) > 0 {
		largest := c.detector.GetLargestArea()
		if largest != nil {
			largest.Active = true
			c.logger.Info("Activated largest scroll area",
				zap.Int("width", largest.Bounds.Dx()),
				zap.Int("height", largest.Bounds.Dy()))
		}
	}

	return nil
}

// Scroll scrolls in the specified direction
func (c *Controller) Scroll(dir Direction, amount ScrollAmount) error {
	area := c.detector.GetActiveArea()
	if area == nil {
		c.logger.Error("No active scroll area")
		return fmt.Errorf("no active scroll area")
	}

	deltaX, deltaY := c.calculateDelta(dir, amount, area)

	c.logger.Info("Scrolling",
		zap.String("direction", c.directionString(dir)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY),
		zap.Int("areaX", area.Bounds.Min.X),
		zap.Int("areaY", area.Bounds.Min.Y),
		zap.Bool("smoothScroll", c.smoothScroll))

	// Disable smooth scroll for now - it's not working properly
	// if c.smoothScroll {
	// 	return c.smoothScrollTo(area, deltaX, deltaY)
	// }

	c.logger.Info("Calling Element.Scroll")
	err := area.Element.Element.Scroll(deltaX, deltaY)
	if err != nil {
		c.logger.Error("Scroll element failed", zap.Error(err))
	} else {
		c.logger.Info("Scroll element succeeded")
	}
	return err
}

// calculateDelta calculates the scroll delta based on direction and amount
func (c *Controller) calculateDelta(dir Direction, amount ScrollAmount, area *ScrollArea) (int, int) {
	var deltaX, deltaY int
	
	// Base scroll amount (in lines for scroll wheel events)
	baseScroll := 3 // 3 lines per scroll

	switch amount {
	case AmountChar:
		// Single scroll
		baseScroll = 3
	case AmountHalfPage:
		// Half page - more lines
		baseScroll = 10
	case AmountFullPage:
		// Full page - even more lines
		baseScroll = 20
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

// smoothScrollTo performs smooth scrolling
func (c *Controller) smoothScrollTo(area *ScrollArea, deltaX, deltaY int) error {
	steps := 10
	stepDuration := time.Millisecond * 20

	stepX := deltaX / steps
	stepY := deltaY / steps

	for i := 0; i < steps; i++ {
		if err := area.Element.Element.Scroll(stepX, stepY); err != nil {
			return err
		}
		time.Sleep(stepDuration)
	}

	return nil
}

// GetActiveArea returns the active scroll area
func (c *Controller) GetActiveArea() *ScrollArea {
	return c.detector.GetActiveArea()
}

// CycleArea cycles to the next scroll area
func (c *Controller) CycleArea() *ScrollArea {
	area := c.detector.CycleActiveArea()
	if area != nil {
		c.logger.Info("Cycled to scroll area",
			zap.Int("width", area.Bounds.Dx()),
			zap.Int("height", area.Bounds.Dy()))
	}
	return area
}

// GetAllAreas returns all detected scroll areas
func (c *Controller) GetAllAreas() []*ScrollArea {
	return c.detector.GetAreas()
}

// Cleanup cleans up the controller
func (c *Controller) Cleanup() {
	c.detector.Clear()
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

// ScrollUpHalfPage scrolls up by half page
func (c *Controller) ScrollUpHalfPage() error {
	return c.Scroll(DirectionUp, AmountHalfPage)
}

// ScrollDownHalfPage scrolls down by half page
func (c *Controller) ScrollDownHalfPage() error {
	return c.Scroll(DirectionDown, AmountHalfPage)
}

// ScrollUpFullPage scrolls up by full page
func (c *Controller) ScrollUpFullPage() error {
	return c.Scroll(DirectionUp, AmountFullPage)
}

// ScrollDownFullPage scrolls down by full page
func (c *Controller) ScrollDownFullPage() error {
	return c.Scroll(DirectionDown, AmountFullPage)
}
