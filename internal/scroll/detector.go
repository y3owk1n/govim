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
	Active  bool
}

// Detector detects scrollable areas
type Detector struct {
	areas []*ScrollArea
}

// NewDetector creates a new scroll detector
func NewDetector() *Detector {
	return &Detector{
		areas: []*ScrollArea{},
	}
}

// DetectScrollAreas detects all scrollable areas in the frontmost window
func (d *Detector) DetectScrollAreas() ([]*ScrollArea, error) {
	elements, err := accessibility.GetScrollableElements()
	if err != nil {
		return nil, fmt.Errorf("failed to get scrollable elements: %w", err)
	}

	areas := make([]*ScrollArea, 0, len(elements))
	for _, element := range elements {
		bounds := element.Element.GetScrollBounds()
		
		// Skip very small scroll areas
		if bounds.Dx() < 50 || bounds.Dy() < 50 {
			continue
		}

		area := &ScrollArea{
			Element: element,
			Bounds:  bounds,
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
	return d.areas[nextIndex]
}

// Clear clears all detected areas
func (d *Detector) Clear() {
	d.areas = []*ScrollArea{}
}
