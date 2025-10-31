package hints

import (
	"image"
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
		{5, []string{"a", "s", "d", "f", "aa"}},
		{8, []string{"a", "s", "d", "f", "aa", "as", "ad", "af"}},
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

	labels := gen.generateNumericLabels(5)
	expected := []string{"1", "2", "3", "4", "5"}

	if len(labels) != len(expected) {
		t.Errorf("Expected %d labels, got %d", len(expected), len(labels))
	}

	for i, label := range labels {
		if label != expected[i] {
			t.Errorf("Label %d: expected %s, got %s", i, expected[i], label)
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
