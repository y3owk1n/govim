package hints

import (
	"image"
	"sort"
	"strings"

	"github.com/y3owk1n/neru/internal/infra/accessibility"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"go.uber.org/zap"
)

// Hint represents a labeled UI element with positioning and sizing information for hint mode.
type Hint struct {
	Label         string
	Element       *accessibility.TreeNode
	Position      image.Point
	Size          image.Point
	MatchedPrefix string // Characters that have been typed
}

// Generator creates hint labels for UI elements based on their position and size.
type Generator struct {
	characters       string
	uppercaseChars   string // Cached uppercase version.
	maxHints         int
	uppercaseRuneMap map[rune]rune // Cache for uppercase rune conversions.
}

// NewGenerator initializes a new hint generator with the specified character set.
func NewGenerator(characters string) *Generator {
	// Ensure we have at least some characters
	if characters == "" {
		// Use home row characters by default
		characters = "asdfghjkl" // fallback to default
	}

	charCount := len(characters)
	maxHints := charCount * charCount * charCount

	logger.Debug(
		"Considered characters",
		zap.String("characters", characters),
		zap.Int("charCount", charCount),
	)
	logger.Debug("Setting maxHints", zap.Int("maxHints", maxHints))

	// Cache uppercase version and create rune map
	uppercaseChars := strings.ToUpper(characters)
	runeMap := make(map[rune]rune, len(characters))
	for i, r := range characters {
		runeMap[r] = rune(uppercaseChars[i])
	}

	return &Generator{
		characters:       characters,
		uppercaseChars:   uppercaseChars,
		maxHints:         maxHints,
		uppercaseRuneMap: runeMap,
	}
}

// GetCharacters returns the characters used for hint generation.
func (g *Generator) GetCharacters() string { return g.characters }

// GetMaxHints returns the maximum number of hints that can be generated.
func (g *Generator) GetMaxHints() int { return g.maxHints }

// UpdateCharacters updates the character set used for hint generation.
func (g *Generator) UpdateCharacters(characters string) {
	if characters == "" {
		characters = "asdfghjkl" // fallback to default
	}
	g.characters = characters
	charCount := len(characters)
	g.maxHints = charCount * charCount * charCount

	// Update cached uppercase version and rune map
	g.uppercaseChars = strings.ToUpper(characters)
	g.uppercaseRuneMap = make(map[rune]rune, len(characters))
	for i, r := range characters {
		g.uppercaseRuneMap[r] = rune(g.uppercaseChars[i])
	}

	logger.Debug("Updated hint characters",
		zap.String("characters", characters),
		zap.Int("maxHints", g.maxHints))
}

// Generate creates hints for the given UI elements, sorted by position and limited by maximum count.
func (g *Generator) Generate(elements []*accessibility.TreeNode) ([]*Hint, error) {
	if len(elements) == 0 {
		return []*Hint{}, nil
	}

	// Sort elements by position (top-to-bottom, left-to-right)
	// Use a sortable slice type to avoid closure allocations
	sortedElements := make(treeNodesByPosition, len(elements))
	copy(sortedElements, elements)
	sort.Sort(sortedElements)

	// Limit to max hints
	if g.maxHints > 0 && len(sortedElements) > g.maxHints {
		sortedElements = sortedElements[:g.maxHints]
	}

	// Generate labels (alphabet-only) - already uppercase
	labels := g.generateAlphabetLabels(len(sortedElements))

	// Pre-allocate hints with exact capacity
	hints := make([]*Hint, len(sortedElements))
	for elementIndex, element := range sortedElements {
		// Position hint at the center of the element (like Vimac does)
		centerX := element.Info.Position.X + (element.Info.Size.X / 2)
		centerY := element.Info.Position.Y + (element.Info.Size.Y / 2)

		hints[elementIndex] = &Hint{
			Label:    labels[elementIndex], // Already uppercase from generateAlphabetLabels
			Element:  element,
			Position: image.Point{X: centerX, Y: centerY},
			Size:     element.Info.Size,
		}
	}

	return hints, nil
}

// treeNodesByPosition implements sort.Interface for sorting TreeNodes by position.
type treeNodesByPosition []*accessibility.TreeNode

func (t treeNodesByPosition) Len() int { return len(t) }

func (t treeNodesByPosition) Less(i, j int) bool {
	posI := t[i].Info.Position
	posJ := t[j].Info.Position

	// Sort by Y first (top to bottom)
	if posI.Y != posJ.Y {
		return posI.Y < posJ.Y
	}
	// Then by X (left to right)
	return posI.X < posJ.X
}

func (t treeNodesByPosition) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// generateAlphabetLabels generates alphabet-based hint labels using a prefix-avoidance strategy.
// Ensures no single character label conflicts with the start of a multi-character label.
// Returns uppercase labels.
func (g *Generator) generateAlphabetLabels(count int) []string {
	if count <= 0 {
		return []string{}
	}

	// Pre-allocate with exact capacity
	labels := make([]string, 0, count)
	chars := []rune(g.uppercaseChars) // Use cached uppercase characters
	numChars := len(chars)

	var length int
	switch {
	case count <= numChars:
		length = 1
	case count <= numChars*numChars:
		length = 2
	default:
		length = 3
	}

	// Generate labels of the determined length
	switch length {
	case 1:
		// Single character labels
		for i := range chars[:count] {
			labels = append(labels, string(chars[i]))
		}
	case 2:
		// All 2-char combinations using strings.Builder
		var b strings.Builder
		b.Grow(2)
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				b.Reset()
				b.WriteRune(chars[i])
				b.WriteRune(chars[j])
				labels = append(labels, b.String())
			}
		}
	case 3:
		// All 3-char combinations using strings.Builder
		var b strings.Builder
		b.Grow(3)
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				for k := 0; k < numChars && len(labels) < count; k++ {
					b.Reset()
					b.WriteRune(chars[i])
					b.WriteRune(chars[j])
					b.WriteRune(chars[k])
					labels = append(labels, b.String())
				}
			}
		}
	}

	return labels[:count]
}

// GetBounds returns the bounding rectangle for a hint.
func (h *Hint) GetBounds() image.Rectangle {
	return image.Rectangle{
		Min: h.GetPosition(),
		Max: image.Point{
			X: h.GetPosition().X + h.GetSize().X,
			Y: h.GetPosition().Y + h.GetSize().Y,
		},
	}
}

// IsVisible checks if the hint is visible on screen.
func (h *Hint) IsVisible(screenBounds image.Rectangle) bool {
	bounds := h.GetBounds()
	return bounds.Overlaps(screenBounds)
}

// GetLabel returns the hint label.
func (h *Hint) GetLabel() string { return h.Label }

// GetElement returns the hint element.
func (h *Hint) GetElement() *accessibility.TreeNode { return h.Element }

// GetPosition returns the hint position.
func (h *Hint) GetPosition() image.Point { return h.Position }

// GetSize returns the hint size.
func (h *Hint) GetSize() image.Point { return h.Size }

// GetMatchedPrefix returns the matched prefix.
func (h *Hint) GetMatchedPrefix() string { return h.MatchedPrefix }

// HintCollection manages a collection of hints.
type HintCollection struct {
	hints   []*Hint
	active  bool
	byLabel map[string]*Hint
	prefix1 map[byte][]*Hint
	prefix2 map[string][]*Hint
}

// NewHintCollection creates a new hint collection.
func NewHintCollection(hints []*Hint) *HintCollection {
	hintCollection := &HintCollection{
		hints:   hints,
		active:  true,
		byLabel: make(map[string]*Hint, len(hints)),
		prefix1: make(map[byte][]*Hint),
		prefix2: make(map[string][]*Hint),
	}
	for _, hint := range hints {
		// Labels are already uppercase from generation
		label := hint.GetLabel()
		hintCollection.byLabel[label] = hint
		if len(label) >= 1 {
			firstByte := label[0]
			hintCollection.prefix1[firstByte] = append(hintCollection.prefix1[firstByte], hint)
		}
		if len(label) >= 2 {
			prefix := label[:2]
			hintCollection.prefix2[prefix] = append(hintCollection.prefix2[prefix], hint)
		}
	}
	return hintCollection
}

// GetHints returns all hints.
func (hc *HintCollection) GetHints() []*Hint {
	return hc.hints
}

// FindByLabel finds a hint by label.
func (hc *HintCollection) FindByLabel(label string) *Hint {
	// Convert to uppercase once for lookup
	upperLabel := strings.ToUpper(label)
	return hc.byLabel[upperLabel]
}

// FilterByPrefix filters hints by prefix.
func (hc *HintCollection) FilterByPrefix(prefix string) []*Hint {
	if prefix == "" {
		return hc.hints
	}
	// Convert to uppercase once at the start
	upperPrefix := strings.ToUpper(prefix)
	prefixLen := len(upperPrefix)

	if prefixLen == 1 {
		firstByte := upperPrefix[0]
		bucket := hc.prefix1[firstByte]
		if len(bucket) == 0 {
			return []*Hint{}
		}
		// Pre-allocate with exact capacity
		out := make([]*Hint, len(bucket))
		copy(out, bucket)
		return out
	}
	if prefixLen >= 2 {
		if bucket, ok := hc.prefix2[upperPrefix[:2]]; ok {
			// Pre-allocate with capacity hint
			out := make([]*Hint, 0, len(bucket))
			for _, hint := range bucket {
				// Labels are already uppercase, direct comparison
				if strings.HasPrefix(hint.GetLabel(), upperPrefix) {
					out = append(out, hint)
				}
			}
			return out
		}
		return []*Hint{}
	}
	return []*Hint{}
}

// IsActive returns whether the collection is active.
func (hc *HintCollection) IsActive() bool {
	return hc.active
}

// Deactivate deactivates the collection.
func (hc *HintCollection) Deactivate() {
	hc.active = false
}

// Count returns the number of hints.
func (hc *HintCollection) Count() int {
	return len(hc.hints)
}

// GetByLabelMap returns the hint map by label.
func (hc *HintCollection) GetByLabelMap() map[string]*Hint { return hc.byLabel }

// GetPrefix1Map returns the prefix1 map.
func (hc *HintCollection) GetPrefix1Map() map[byte][]*Hint { return hc.prefix1 }

// GetPrefix2Map returns the prefix2 map.
func (hc *HintCollection) GetPrefix2Map() map[string][]*Hint { return hc.prefix2 }
