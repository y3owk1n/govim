package scroll

import (
	"fmt"
	"image"

	"github.com/y3owk1n/govim/internal/accessibility"
)

// ScrollArea represents a scrollable area
type ScrollArea struct {
	Element *accessibility.TreeNode
	Bounds  image.Rectangle
	Index   int
	Active  bool
}

// Detector detects scrollable areas
type Detector struct {
	areas       []*ScrollArea
	activeIndex int
}

// NewDetector creates a new scroll detector
func NewDetector() *Detector {
	return &Detector{
		areas:       []*ScrollArea{},
		activeIndex: 0,
	}
}

const (
	// MinScrollAreaWidth is the minimum width for a valid scroll area
	MinScrollAreaWidth = 50
	// MinScrollAreaHeight is the minimum height for a valid scroll area
	MinScrollAreaHeight = 50
)

// DetectScrollAreas detects all scrollable areas in the frontmost window
func (d *Detector) DetectScrollAreas() ([]*ScrollArea, error) {
	elements, err := accessibility.GetScrollableElements()
	if err != nil {
		return nil, fmt.Errorf("failed to get scrollable elements: %w", err)
	}

	areas := make([]*ScrollArea, 0, len(elements))
	for i, element := range elements {
		bounds := element.Element.GetScrollBounds()

		// Skip very small scroll areas
		if bounds.Dx() < MinScrollAreaWidth || bounds.Dy() < MinScrollAreaHeight {
			continue
		}

		area := &ScrollArea{
			Element: element,
			Bounds:  bounds,
			Index:   i,
			Active:  false,
		}
		areas = append(areas, area)
	}

	d.areas = areas
	return areas, nil
}

// GetAreas returns all detected scroll areas
func (d *Detector) GetAreas() []*ScrollArea {
	return d.areas
}

// GetActiveArea returns the currently active scroll area
func (d *Detector) GetActiveArea() *ScrollArea {
	for _, area := range d.areas {
		if area.Active {
			return area
		}
	}
	return nil
}

// SetActiveArea sets the active scroll area
func (d *Detector) SetActiveArea(index int) error {
	if index < 0 || index >= len(d.areas) {
		return fmt.Errorf("invalid area index: %d", index)
	}

	// Deactivate all areas
	for _, area := range d.areas {
		area.Active = false
	}

	// Activate selected area
	d.areas[index].Active = true
	return nil
}

// GetLargestArea returns the largest scroll area
func (d *Detector) GetLargestArea() *ScrollArea {
	if len(d.areas) == 0 {
		return nil
	}

	largest := d.areas[0]
	maxArea := largest.Bounds.Dx() * largest.Bounds.Dy()

	for _, area := range d.areas[1:] {
		areaSize := area.Bounds.Dx() * area.Bounds.Dy()
		if areaSize > maxArea {
			largest = area
			maxArea = areaSize
		}
	}

	return largest
}

// GetAreaAtPosition returns the scroll area at the given position
func (d *Detector) GetAreaAtPosition(x, y int) *ScrollArea {
	point := image.Point{X: x, Y: y}

	for _, area := range d.areas {
		if point.In(area.Bounds) {
			return area
		}
	}

	return nil
}

// CycleActiveArea cycles to the next scroll area
func (d *Detector) CycleActiveArea() *ScrollArea {
	if len(d.areas) == 0 {
		return nil
	}

	currentIndex := -1
	for i, area := range d.areas {
		if area.Active {
			currentIndex = i
			area.Active = false
			break
		}
	}

	nextIndex := (currentIndex + 1) % len(d.areas)
	d.areas[nextIndex].Active = true
	d.areas[nextIndex].Index = nextIndex
	return d.areas[nextIndex]
}

// CyclePrevArea cycles to the previous scroll area
func (d *Detector) CyclePrevArea() *ScrollArea {
	if len(d.areas) == 0 {
		return nil
	}

	currentIndex := -1
	for i, area := range d.areas {
		if area.Active {
			currentIndex = i
			area.Active = false
			break
		}
	}

	prevIndex := currentIndex - 1
	if prevIndex < 0 {
		prevIndex = len(d.areas) - 1
	}
	d.areas[prevIndex].Active = true
	d.areas[prevIndex].Index = prevIndex
	return d.areas[prevIndex]
}

// SetActiveByNumber sets the active area by number (1-indexed)
func (d *Detector) SetActiveByNumber(number int) *ScrollArea {
	index := number - 1
	if index < 0 || index >= len(d.areas) {
		return nil
	}

	// Deactivate all
	for _, area := range d.areas {
		area.Active = false
	}

	// Activate selected
	d.areas[index].Active = true
	d.areas[index].Index = index
	return d.areas[index]
}

// Clear clears all detected areas
func (d *Detector) Clear() {
	d.areas = []*ScrollArea{}
}
