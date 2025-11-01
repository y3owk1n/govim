package hints

import (
	"image"
	"strings"
	"testing"

	"github.com/y3owk1n/govim/internal/accessibility"
	_ "github.com/y3owk1n/govim/internal/bridge" // Import for CGo compilation
)

func TestGenerateAlphabetLabels(t *testing.T) {
	gen := NewGenerator("asdf", "alphabet", 100)

	tests := []struct {
		count    int
		expected []string
	}{
		{0, []string{}},
		{1, []string{"a"}},
		{4, []string{"a", "s", "d", "f"}},
		{5, []string{"aa", "as", "ad", "af", "sa"}},                   // All 2-char when count > numChars
		{8, []string{"aa", "as", "ad", "af", "sa", "ss", "sd", "sf"}}, // All 2-char
	}

	for _, tt := range tests {
		labels := gen.generateAlphabetLabels(tt.count)
		if len(labels) != len(tt.expected) {
			t.Errorf("Expected %d labels, got %d", len(tt.expected), len(labels))
			continue
		}

		for i, label := range labels {
			if label != tt.expected[i] {
				t.Errorf("Label %d: expected %s, got %s", i, tt.expected[i], label)
			}
		}
	}
}

func TestGenerateNumericLabels(t *testing.T) {
	gen := NewGenerator("", "numeric", 100)

	tests := []struct {
		count    int
		expected []string
	}{
		{0, []string{}},
		{1, []string{"1"}},
		{5, []string{"1", "2", "3", "4", "5"}},
		{10, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
	}

	for _, tt := range tests {
		labels := gen.generateNumericLabels(tt.count)
		if len(labels) != len(tt.expected) {
			t.Errorf("Expected %d labels, got %d", len(tt.expected), len(labels))
			continue
		}

		for i, label := range labels {
			if label != tt.expected[i] {
				t.Errorf("Label %d: expected %s, got %s", i, tt.expected[i], label)
			}
		}
	}
}

func TestFindHintByLabel(t *testing.T) {
	hints := []*Hint{
		{Label: "a", Position: image.Point{X: 0, Y: 0}},
		{Label: "s", Position: image.Point{X: 10, Y: 10}},
		{Label: "d", Position: image.Point{X: 20, Y: 20}},
	}

	tests := []struct {
		label    string
		expected bool
	}{
		{"a", true},
		{"s", true},
		{"d", true},
		{"f", false},
		{"A", true}, // case insensitive
	}

	for _, tt := range tests {
		hint := FindHintByLabel(hints, tt.label)
		found := hint != nil

		if found != tt.expected {
			t.Errorf("Label %s: expected found=%v, got found=%v", tt.label, tt.expected, found)
		}
	}
}

func TestFilterHintsByPrefix(t *testing.T) {
	hints := []*Hint{
		{Label: "aa"},
		{Label: "ab"},
		{Label: "ba"},
		{Label: "bb"},
	}

	tests := []struct {
		prefix   string
		expected int
	}{
		{"", 4},
		{"a", 2},
		{"b", 2},
		{"aa", 1},
		{"c", 0},
	}

	for _, tt := range tests {
		filtered := FilterHintsByPrefix(hints, tt.prefix)
		if len(filtered) != tt.expected {
			t.Errorf("Prefix %s: expected %d hints, got %d", tt.prefix, tt.expected, len(filtered))
		}
	}
}

func TestHintGetBounds(t *testing.T) {
	hint := &Hint{
		Position: image.Point{X: 10, Y: 20},
		Size:     image.Point{X: 30, Y: 40},
	}

	bounds := hint.GetBounds()

	if bounds.Min.X != 10 || bounds.Min.Y != 20 {
		t.Errorf("Expected Min=(10,20), got Min=(%d,%d)", bounds.Min.X, bounds.Min.Y)
	}

	if bounds.Max.X != 40 || bounds.Max.Y != 60 {
		t.Errorf("Expected Max=(40,60), got Max=(%d,%d)", bounds.Max.X, bounds.Max.Y)
	}
}

func TestHintGetCenter(t *testing.T) {
	hint := &Hint{
		Position: image.Point{X: 10, Y: 20},
		Size:     image.Point{X: 30, Y: 40},
	}

	center := hint.GetCenter()

	expectedX := 10 + 30/2
	expectedY := 20 + 40/2

	if center.X != expectedX || center.Y != expectedY {
		t.Errorf("Expected center=(%d,%d), got center=(%d,%d)", expectedX, expectedY, center.X, center.Y)
	}
}

func TestHintIsVisible(t *testing.T) {
	hint := &Hint{
		Position: image.Point{X: 10, Y: 10},
		Size:     image.Point{X: 20, Y: 20},
	}

	tests := []struct {
		name         string
		screenBounds image.Rectangle
		expected     bool
	}{
		{
			name:         "fully visible",
			screenBounds: image.Rect(0, 0, 100, 100),
			expected:     true,
		},
		{
			name:         "partially visible",
			screenBounds: image.Rect(0, 0, 20, 20),
			expected:     true,
		},
		{
			name:         "not visible",
			screenBounds: image.Rect(50, 50, 100, 100),
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visible := hint.IsVisible(tt.screenBounds)
			if visible != tt.expected {
				t.Errorf("Expected visible=%v, got visible=%v", tt.expected, visible)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	// Create test elements
	elements := []*accessibility.TreeNode{
		{
			Info: &accessibility.ElementInfo{
				Position: image.Point{X: 10, Y: 10},
				Size:     image.Point{X: 20, Y: 20},
				Role:     "button",
			},
		},
		{
			Info: &accessibility.ElementInfo{
				Position: image.Point{X: 50, Y: 10},
				Size:     image.Point{X: 20, Y: 20},
				Role:     "link",
			},
		},
		{
			Info: &accessibility.ElementInfo{
				Position: image.Point{X: 10, Y: 50},
				Size:     image.Point{X: 20, Y: 20},
				Role:     "checkbox",
			},
		},
	}

	tests := []struct {
		name       string
		characters string
		style      string
		maxHints   int
		elements   []*accessibility.TreeNode
		wantCount  int
		wantLabels []string
	}{
		{
			name:       "alphabet style",
			characters: "asdf",
			style:      "alphabet",
			maxHints:   10,
			elements:   elements,
			wantCount:  3,
			wantLabels: []string{"A", "S", "D"},
		},
		{
			name:       "numeric style",
			characters: "asdf",
			style:      "numeric",
			maxHints:   10,
			elements:   elements,
			wantCount:  3,
			wantLabels: []string{"1", "2", "3"},
		},
		{
			name:       "limited hints",
			characters: "asdf",
			style:      "alphabet",
			maxHints:   2,
			elements:   elements,
			wantCount:  2,
			wantLabels: []string{"A", "S"},
		},
		{
			name:       "empty elements",
			characters: "asdf",
			style:      "alphabet",
			maxHints:   10,
			elements:   []*accessibility.TreeNode{},
			wantCount:  0,
			wantLabels: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewGenerator(tt.characters, tt.style, tt.maxHints)
			hints, err := generator.Generate(tt.elements)

			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if len(hints) != tt.wantCount {
				t.Errorf("Generate() returned %d hints, want %d", len(hints), tt.wantCount)
			}

			// Check labels
			for i, hint := range hints {
				if i < len(tt.wantLabels) && hint.Label != tt.wantLabels[i] {
					t.Errorf("Generate() hint[%d].Label = %s, want %s", i, hint.Label, tt.wantLabels[i])
				}
			}

			// Check positions are centered
			for i, hint := range hints {
				element := tt.elements[i]
				expectedX := element.Info.Position.X + (element.Info.Size.X / 2)
				expectedY := element.Info.Position.Y + (element.Info.Size.Y / 2)

				if hint.Position.X != expectedX || hint.Position.Y != expectedY {
					t.Errorf("Generate() hint[%d].Position = (%d,%d), want (%d,%d)",
						i, hint.Position.X, hint.Position.Y, expectedX, expectedY)
				}
			}
		})
	}
}

func TestHintCollection(t *testing.T) {
	hints := []*Hint{
		{Label: "a"},
		{Label: "s"},
		{Label: "d"},
	}

	collection := NewHintCollection(hints)

	if !collection.IsActive() {
		t.Error("Expected collection to be active")
	}

	if collection.Count() != 3 {
		t.Errorf("Expected count=3, got count=%d", collection.Count())
	}

	hint := collection.FindByLabel("s")
	if hint == nil || hint.Label != "s" {
		t.Error("Failed to find hint by label")
	}

	filtered := collection.FilterByPrefix("a")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 filtered hint, got %d", len(filtered))
	}

	collection.Deactivate()
	if collection.IsActive() {
		t.Error("Expected collection to be inactive")
	}
}

func createMockElement(x, y int) *accessibility.TreeNode {
	return &accessibility.TreeNode{
		Info: &accessibility.ElementInfo{
			Position: image.Point{X: x, Y: y},
			Size:     image.Point{X: 50, Y: 20},
		},
	}
}

func TestGenerateHints(t *testing.T) {
	gen := NewGenerator("asdf", "alphabet", 100)

	elements := []*accessibility.TreeNode{
		createMockElement(10, 10),
		createMockElement(100, 10),
		createMockElement(10, 100),
	}

	hints, err := gen.Generate(elements)
	if err != nil {
		t.Fatalf("Failed to generate hints: %v", err)
	}

	if len(hints) != 3 {
		t.Errorf("Expected 3 hints, got %d", len(hints))
	}

	// Check that hints are sorted by position
	if hints[0].Position.Y > hints[1].Position.Y {
		t.Error("Hints not sorted by Y position")
	}
}

func TestGenerateHints_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		elements    []*accessibility.TreeNode
		maxHints    int
		expectError bool
		expectedLen int
	}{
		{
			name:        "Empty elements",
			elements:    []*accessibility.TreeNode{},
			maxHints:    10,
			expectError: false,
			expectedLen: 0,
		},
		{
			name:        "Nil elements",
			elements:    nil,
			maxHints:    10,
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "Single element",
			elements: []*accessibility.TreeNode{
				createMockElement(10, 10),
			},
			maxHints:    10,
			expectError: false,
			expectedLen: 1,
		},
		{
			name: "Max hints limit",
			elements: []*accessibility.TreeNode{
				createMockElement(10, 10),
				createMockElement(20, 20),
				createMockElement(30, 30),
			},
			maxHints:    2,
			expectError: false,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator("asdf", "alphabet", tt.maxHints)
			hints, err := gen.Generate(tt.elements)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(hints) != tt.expectedLen {
				t.Errorf("Expected %d hints, got %d", tt.expectedLen, len(hints))
			}

			// Verify hints have proper properties
			for i, hint := range hints {
				if hint.Label == "" {
					t.Errorf("Hint %d has empty label", i)
				}
				if hint.Element == nil && tt.elements != nil {
					t.Errorf("Hint %d has nil element", i)
				}
			}
		})
	}
}

func TestGenerateHintsWithMaxLimit(t *testing.T) {
	gen := NewGenerator("asdf", "alphabet", 2)

	elements := []*accessibility.TreeNode{
		createMockElement(10, 10),
		createMockElement(100, 10),
		createMockElement(10, 100),
	}

	hints, err := gen.Generate(elements)
	if err != nil {
		t.Fatalf("Failed to generate hints: %v", err)
	}

	if len(hints) != 2 {
		t.Errorf("Expected 2 hints (max limit), got %d", len(hints))
	}
}

func TestGenerateAlphabetLabels_EdgeCases(t *testing.T) {
	// Test with different character sets
	tests := []struct {
		name      string
		chars     string
		count     int
		wantError bool
	}{
		{"Empty chars", "", 5, false},  // Should use fallback
		{"Single char", "a", 1, false}, // Single char, single label
		{"Many chars", "asdfghjkl", 100, false},
		{"Zero count", "asdf", 0, false},
		{"Negative count", "asdf", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.chars, "alphabet", 1000)
			labels := gen.generateAlphabetLabels(tt.count)

			if tt.count <= 0 {
				// Zero or negative counts should return empty slice
				if len(labels) != 0 {
					t.Errorf("Expected empty labels for count %d, got %d labels", tt.count, len(labels))
				}
				return
			}

			if len(labels) != tt.count {
				t.Errorf("Expected %d labels, got %d", tt.count, len(labels))
			}

			// Debug output
			t.Logf("Generated labels: %v", labels)

			// Verify no empty strings
			for i, label := range labels {
				if label == "" {
					t.Errorf("Label %d is empty", i)
				}
			}

			// Verify no prefix conflicts (only for multi-character labels)
			if len(labels) > 1 && len(labels[0]) > 1 {
				for i, label1 := range labels {
					for j, label2 := range labels {
						if i != j && strings.HasPrefix(label2, label1) {
							t.Errorf("Prefix conflict: '%s' is a prefix of '%s'", label1, label2)
						}
					}
				}
			}
		})
	}
}

func TestNoPrefixConflicts(t *testing.T) {
	tests := []struct {
		name       string
		chars      string
		count      int
		shouldFail bool
	}{
		{"9 chars, 81 hints (2-char)", "asdfghjkl", 81, false},
		{"9 chars, 82 hints (3-char)", "asdfghjkl", 82, false},
		{"9 chars, 100 hints (3-char)", "asdfghjkl", 100, false},
		{"4 chars, 16 hints (2-char)", "asdf", 16, false},
		{"4 chars, 17 hints (3-char)", "asdf", 17, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tt.chars, "alphabet", 1000)
			labels := gen.generateAlphabetLabels(tt.count)

			if len(labels) != tt.count {
				t.Fatalf("Expected %d labels, got %d", tt.count, len(labels))
			}

			// Check for prefix conflicts
			for i, label1 := range labels {
				for j, label2 := range labels {
					if i != j && strings.HasPrefix(label2, label1) {
						t.Errorf("Prefix conflict: '%s' is a prefix of '%s'", label1, label2)
					}
				}
			}

			// All labels should have the same length (except when count <= numChars)
			if tt.count > len(tt.chars) {
				expectedLen := labels[0]
				for _, label := range labels {
					if len(label) != len(expectedLen) {
						t.Errorf("Inconsistent label length: got '%s' (len=%d) and '%s' (len=%d)",
							expectedLen, len(expectedLen), label, len(label))
						break
					}
				}
			}
		})
	}
}
