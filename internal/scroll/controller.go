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

// Controller manages scrolling operations
type Controller struct {
	detector *Detector
	config   *config.ScrollConfig
	logger   *zap.Logger
}

// NewController creates a new scroll controller
func NewController(cfg config.ScrollConfig, logger *zap.Logger) *Controller {
	return &Controller{
		detector: NewDetector(),
		config:   &cfg,
		logger:   logger,
	}
}

// Initialize initializes the scroll controller
func (c *Controller) Initialize() error {
	c.logger.Info("Starting scroll area detection")
	
	// Detect scroll areas
	areas, err := c.detector.DetectScrollAreas()
	if err != nil {
		c.logger.Error("Failed to detect scroll areas", zap.Error(err))
		return fmt.Errorf("failed to detect scroll areas: %w", err)
	}

	c.logger.Info("Detected scroll areas", zap.Int("count", len(areas)))

	if len(areas) == 0 {
		return fmt.Errorf("no scroll areas found")
	}

	c.logger.Info("Getting largest area")
	// Activate the largest scroll area by default
	largestArea := c.detector.GetLargestArea()
	if largestArea != nil {
		c.logger.Info("Setting active area")
		c.detector.SetActiveArea(0)
		largestArea.Active = true

		c.logger.Info("Activated largest scroll area",
			zap.Int("width", largestArea.Bounds.Dx()),
			zap.Int("height", largestArea.Bounds.Dy()))
	}

	c.logger.Info("Scroll controller initialized successfully")
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

	c.logger.Debug("Scrolling",
		zap.String("direction", c.directionString(dir)),
		zap.Int("deltaX", deltaX),
		zap.Int("deltaY", deltaY))

	if err := area.Element.Element.Scroll(deltaX, deltaY); err != nil {
		c.logger.Error("Scroll failed", zap.Error(err))
		return err
	}
	return nil
}

// calculateDelta calculates the scroll delta based on direction and amount
func (c *Controller) calculateDelta(dir Direction, amount ScrollAmount, area *ScrollArea) (int, int) {
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

const (
	// SmoothScrollSteps is the number of steps for smooth scrolling
	SmoothScrollSteps = 10
	// SmoothScrollStepDuration is the duration between smooth scroll steps
	SmoothScrollStepDuration = 20 * time.Millisecond
)

// smoothScrollTo performs smooth scrolling
func (c *Controller) smoothScrollTo(area *ScrollArea, deltaX, deltaY int) error {
	stepX := deltaX / SmoothScrollSteps
	stepY := deltaY / SmoothScrollSteps

	for i := 0; i < SmoothScrollSteps; i++ {
		if err := area.Element.Element.Scroll(stepX, stepY); err != nil {
			return err
		}
		time.Sleep(SmoothScrollStepDuration)
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

// GetDetector returns the scroll detector
func (c *Controller) GetDetector() *Detector {
	return c.detector
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
	area := c.detector.GetActiveArea()
	if area == nil {
		return fmt.Errorf("no active scroll area")
	}

	// Use config values for scroll to edge
	c.logger.Info("Scrolling to top")
	for i := 0; i < c.config.ScrollToEdgeIterations; i++ {
		area.Element.Element.Scroll(0, c.config.ScrollToEdgeDelta)
	}
	return nil
}

// ScrollToBottom scrolls to the bottom (G in Vim)
func (c *Controller) ScrollToBottom() error {
	area := c.detector.GetActiveArea()
	if area == nil {
		return fmt.Errorf("no active scroll area")
	}

	// Use config values for scroll to edge
	c.logger.Info("Scrolling to bottom")
	for i := 0; i < c.config.ScrollToEdgeIterations; i++ {
		area.Element.Element.Scroll(0, -c.config.ScrollToEdgeDelta)
	}
	return nil
}

// ScrollUpFullPage scrolls up by full page
func (c *Controller) ScrollUpFullPage() error {
	return c.Scroll(DirectionUp, AmountFullPage)
}

// ScrollDownFullPage scrolls down by full page
func (c *Controller) ScrollDownFullPage() error {
	return c.Scroll(DirectionDown, AmountFullPage)
}
