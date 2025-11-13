package hints

import (
	"image"
	"sort"
	"strings"
	"unicode"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

// Hint represents a hint label for a UI element
type Hint struct {
	Label         string
	Element       *accessibility.TreeNode
	Position      image.Point
	Size          image.Point
	MatchedPrefix string // Characters that have been typed
}

// Generator generates hints for UI elements
type Generator struct {
	characters string
	maxHints   int
}

// NewGenerator creates a new hint generator
func NewGenerator(characters string) *Generator {
	// Ensure we have at least some characters
	if characters == "" {
		// Use home row characters by default
		characters = "asdfghjkl" // fallback to default
	}

	charCount := len(characters)
	maxHints := charCount * charCount * charCount

	logger.Debug("Considered characters", zap.String("characters", characters), zap.Int("charCount", charCount))
	logger.Debug("Setting maxHints", zap.Int("maxHints", maxHints))

	return &Generator{
		characters: characters,
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

	// Generate labels (alphabet-only)
	labels := g.generateAlphabetLabels(len(sortedElements))

	// Generate hints
	hints := make([]*Hint, len(sortedElements))
	for i, element := range sortedElements {
		// Position hint at the center of the element (like Vimac does)
		centerX := element.Info.Position.X + (element.Info.Size.X / 2)
		centerY := element.Info.Position.Y + (element.Info.Size.Y / 2)

		hints[i] = &Hint{
			Label:    strings.ToUpper(labels[i]), // Convert to uppercase
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
	if count <= 0 {
		return []string{}
	}

	labels := make([]string, 0, count)
	chars := []rune(g.characters)
	numChars := len(chars)

	// Calculate minimum length needed to avoid prefix conflicts
	// For n chars: 1-char = n, 2-char = n², 3-char = n³
	var length int
	if count <= numChars {
		length = 1
	} else if count <= numChars*numChars {
		length = 2
	} else {
		length = 3
	}

	// Generate labels of the determined length
	switch length {
	case 1:
		// Single character labels
		for i := 0; i < count; i++ {
			labels = append(labels, string(chars[i]))
		}
	case 2:
		// All 2-char combinations
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				labels = append(labels, string(chars[i])+string(chars[j]))
			}
		}
	case 3:
		// All 3-char combinations
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				for k := 0; k < numChars && len(labels) < count; k++ {
					labels = append(labels, string(chars[i])+string(chars[j])+string(chars[k]))
				}
			}
		}
	}

	return labels[:count]
}

// GenerateSimpleLabels generates sequential single-character labels without sorting.
// This is used when we want to preserve the original order.
func (g *Generator) GenerateSimpleLabels(count int) []string {
	if count <= 0 {
		return []string{}
	}

	chars := []rune(g.characters)
	if count > len(chars) {
		count = len(chars) // Limit to available characters
	}

	labels := make([]string, count)
	for i := 0; i < count; i++ {
		labels[i] = strings.ToUpper(string(chars[i]))
	}

	return labels
}

// GenerateScrollLabels generates labels for scroll areas, excluding scroll control keys.
// This ensures we don't create hints that conflict with scroll navigation keys, and supports
// multi-character combinations like the regular hint system.
func (g *Generator) GenerateScrollLabels(count int) []string {
	if count <= 0 {
		return []string{}
	}

	// Define scroll control keys to exclude
	scrollKeys := map[rune]bool{
		'j': true, // down
		'k': true, // up
		'h': true, // left
		'l': true, // right
		'g': true, // top
		'G': true, // bottom
		'u': true, // unused but was a scroll key
		'd': true, // unused but was a scroll key
	}

	// Filter available characters
	var availableChars []rune
	for _, ch := range g.characters {
		ch = unicode.ToLower(ch)
		if !scrollKeys[ch] {
			availableChars = append(availableChars, ch)
		}
	}

	if len(availableChars) == 0 {
		logger.Warn("No available characters for scroll hints after filtering")
		return []string{}
	}

	numChars := len(availableChars)
	labels := make([]string, 0, count)

	// Calculate minimum length needed to avoid prefix conflicts
	// For n chars: 1-char = n, 2-char = n², 3-char = n³
	var length int
	if count <= numChars {
		length = 1
	} else if count <= numChars*numChars {
		length = 2
	} else {
		length = 3
	}

	// Generate labels of the determined length
	switch length {
	case 1:
		// Single character labels
		for i := 0; i < count; i++ {
			labels = append(labels, string(availableChars[i]))
		}
	case 2:
		// All 2-char combinations
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				labels = append(labels, string(availableChars[i])+string(availableChars[j]))
			}
		}
	case 3:
		// All 3-char combinations
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				for k := 0; k < numChars && len(labels) < count; k++ {
					labels = append(labels, string(availableChars[i])+string(availableChars[j])+string(availableChars[k]))
				}
			}
		}
	}

	// Convert all labels to uppercase
	for i := range labels {
		labels[i] = strings.ToUpper(labels[i])
	}

	return labels[:count]
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
	hints   []*Hint
	active  bool
	byLabel map[string]*Hint
	prefix1 map[byte][]*Hint
	prefix2 map[string][]*Hint
}

// NewHintCollection creates a new hint collection
func NewHintCollection(hints []*Hint) *HintCollection {
	hc := &HintCollection{
		hints:   hints,
		active:  true,
		byLabel: make(map[string]*Hint, len(hints)),
		prefix1: make(map[byte][]*Hint),
		prefix2: make(map[string][]*Hint),
	}
	for _, h := range hints {
		lbl := strings.ToUpper(h.Label)
		hc.byLabel[lbl] = h
		if len(lbl) >= 1 {
			b := lbl[0]
			hc.prefix1[b] = append(hc.prefix1[b], h)
		}
		if len(lbl) >= 2 {
			p2 := lbl[:2]
			hc.prefix2[p2] = append(hc.prefix2[p2], h)
		}
	}
	return hc
}

// GetHints returns all hints
func (hc *HintCollection) GetHints() []*Hint {
	return hc.hints
}

// FindByLabel finds a hint by label
func (hc *HintCollection) FindByLabel(label string) *Hint {
	return hc.byLabel[strings.ToUpper(label)]
}

// FilterByPrefix filters hints by prefix
func (hc *HintCollection) FilterByPrefix(prefix string) []*Hint {
	if prefix == "" {
		return hc.hints
	}
	p := strings.ToUpper(prefix)
	if len(p) == 1 {
		b := p[0]
		bucket := hc.prefix1[b]
		if len(bucket) == 0 {
			return []*Hint{}
		}
		out := make([]*Hint, 0, len(bucket))
		out = append(out, bucket...)
		return out
	}
	if len(p) >= 2 {
		if bucket, ok := hc.prefix2[p[:2]]; ok {
			out := make([]*Hint, 0, len(bucket))
			for _, h := range bucket {
				if strings.HasPrefix(h.Label, p) {
					out = append(out, h)
				}
			}
			return out
		}
		return []*Hint{}
	}
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
