package hints

import (
	"fmt"
	"image"
	"sort"
	"strings"

	"github.com/y3owk1n/govim/internal/accessibility"
)

// Hint represents a hint label for a UI element
type Hint struct {
	Label    string
	Element  *accessibility.TreeNode
	Position image.Point
	Size     image.Point
}

// Generator generates hints for UI elements
type Generator struct {
	characters string
	style      string
	maxHints   int
}

// NewGenerator creates a new hint generator
func NewGenerator(characters, style string, maxHints int) *Generator {
	return &Generator{
		characters: characters,
		style:      style,
		maxHints:   maxHints,
	}
}

// Generate generates hints for the given elements
func (g *Generator) Generate(elements []*accessibility.TreeNode) ([]*Hint, error) {
	if len(elements) == 0 {
		return []*Hint{}, nil
	}

	// Sort elements by position (top-to-bottom, left-to-right)
	sortedElements := make([]*accessibility.TreeNode, len(elements))
	copy(sortedElements, elements)
	sort.Slice(sortedElements, func(i, j int) bool {
		posI := sortedElements[i].Info.Position
		posJ := sortedElements[j].Info.Position

		// Sort by Y first (top to bottom)
		if posI.Y != posJ.Y {
			return posI.Y < posJ.Y
		}
		// Then by X (left to right)
		return posI.X < posJ.X
	})

	// Limit to max hints
	if g.maxHints > 0 && len(sortedElements) > g.maxHints {
		sortedElements = sortedElements[:g.maxHints]
	}

	// Generate labels
	var labels []string
	if g.style == "alphabet" {
		labels = g.generateAlphabetLabels(len(sortedElements))
	} else {
		labels = g.generateNumericLabels(len(sortedElements))
	}

	// Generate hints
	hints := make([]*Hint, len(sortedElements))
	for i, element := range sortedElements {
		// Position hint at the center of the element (like Vimac does)
		centerX := element.Info.Position.X + (element.Info.Size.X / 2)
		centerY := element.Info.Position.Y + (element.Info.Size.Y / 2)
		
		hints[i] = &Hint{
			Label:    labels[i],
			Element:  element,
			Position: image.Point{X: centerX, Y: centerY},
			Size:     element.Info.Size,
		}
	}

	return hints, nil
}

// generateAlphabetLabels generates alphabet-based hint labels
// Uses a strategy that avoids prefixes (no "a" if "aa" exists)
func (g *Generator) generateAlphabetLabels(count int) []string {
	if count == 0 {
		return []string{}
	}

	labels := make([]string, 0, count)
	chars := []rune(g.characters)
	numChars := len(chars)

	// Calculate how many characters we need
	// For n chars, we can make: n + n^2 + n^3 + ... labels
	// We want to use the minimum length that fits all hints
	
	if count <= numChars {
		// Single character labels
		for i := 0; i < count; i++ {
			labels = append(labels, string(chars[i]))
		}
	} else {
		// Multi-character labels - use all 2-char combinations
		// This avoids the prefix problem (no single chars if we have 2-char)
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				labels = append(labels, string(chars[i])+string(chars[j]))
			}
		}
		
		// If we need more, add 3-char combinations
		if len(labels) < count {
			for i := 0; i < numChars && len(labels) < count; i++ {
				for j := 0; j < numChars && len(labels) < count; j++ {
					for k := 0; k < numChars && len(labels) < count; k++ {
						labels = append(labels, string(chars[i])+string(chars[j])+string(chars[k]))
					}
				}
			}
		}
	}

	return labels[:count]
}

// numberToAlphabet is no longer used but kept for reference
func (g *Generator) numberToAlphabet(num int, chars []rune, base int) string {
	if num < base {
		return string(chars[num])
	}

	result := ""
	for num >= 0 {
		result = string(chars[num%base]) + result
		num = num/base - 1
		if num < 0 {
			break
		}
	}

	return result
}

// generateNumericLabels generates numeric labels (1, 2, 3, ...)
func (g *Generator) generateNumericLabels(count int) []string {
	labels := make([]string, count)
	for i := 0; i < count; i++ {
		labels[i] = fmt.Sprintf("%d", i+1)
	}
	return labels
}

// FindHintByLabel finds a hint by its label
func FindHintByLabel(hints []*Hint, label string) *Hint {
	label = strings.ToLower(label)
	for _, hint := range hints {
		if strings.ToLower(hint.Label) == label {
			return hint
		}
	}
	return nil
}

// FilterHintsByPrefix filters hints by label prefix
func FilterHintsByPrefix(hints []*Hint, prefix string) []*Hint {
	if prefix == "" {
		return hints
	}

	prefix = strings.ToLower(prefix)
	filtered := make([]*Hint, 0)

	for _, hint := range hints {
		if strings.HasPrefix(strings.ToLower(hint.Label), prefix) {
			filtered = append(filtered, hint)
		}
	}

	return filtered
}

// GetHintBounds returns the bounding rectangle for a hint
func (h *Hint) GetBounds() image.Rectangle {
	return image.Rectangle{
		Min: h.Position,
		Max: image.Point{
			X: h.Position.X + h.Size.X,
			Y: h.Position.Y + h.Size.Y,
		},
	}
}

// GetCenter returns the center point of the hint
func (h *Hint) GetCenter() image.Point {
	return image.Point{
		X: h.Position.X + h.Size.X/2,
		Y: h.Position.Y + h.Size.Y/2,
	}
}

// IsVisible checks if the hint is visible on screen
func (h *Hint) IsVisible(screenBounds image.Rectangle) bool {
	bounds := h.GetBounds()
	return bounds.Overlaps(screenBounds)
}

// HintCollection manages a collection of hints
type HintCollection struct {
	hints  []*Hint
	active bool
}

// NewHintCollection creates a new hint collection
func NewHintCollection(hints []*Hint) *HintCollection {
	return &HintCollection{
		hints:  hints,
		active: true,
	}
}

// GetHints returns all hints
func (hc *HintCollection) GetHints() []*Hint {
	return hc.hints
}

// FindByLabel finds a hint by label
func (hc *HintCollection) FindByLabel(label string) *Hint {
	return FindHintByLabel(hc.hints, label)
}

// FilterByPrefix filters hints by prefix
func (hc *HintCollection) FilterByPrefix(prefix string) []*Hint {
	return FilterHintsByPrefix(hc.hints, prefix)
}

// IsActive returns whether the collection is active
func (hc *HintCollection) IsActive() bool {
	return hc.active
}

// Deactivate deactivates the collection
func (hc *HintCollection) Deactivate() {
	hc.active = false
}

// Count returns the number of hints
func (hc *HintCollection) Count() int {
	return len(hc.hints)
}
