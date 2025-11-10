package grid

import (
	"image"
	"strings"
)

// Grid represents a flat 3-character coordinate grid (all positions visible at once)
type Grid struct {
	characters string          // Characters used for coordinates (e.g., "asdfghjkl")
	bounds     image.Rectangle // Screen bounds
	cells      []*Cell         // All cells with 3-char coordinates
}

// Cell represents a grid cell with a 3-character coordinate
type Cell struct {
	Coordinate string          // 3-character coordinate (e.g., "AAA", "ABC")
	Bounds     image.Rectangle // Cell bounds
	Center     image.Point     // Center point
}

// NewGrid creates a grid with practical cell sizes that completely fill the screen
func NewGrid(characters string, minCellSize int, bounds image.Rectangle) *Grid {
	if characters == "" {
		characters = "asdfghjkl"
	}
	characters = strings.ToUpper(characters)
	chars := []rune(characters)
	numChars := len(chars)

	// Define minimum comfortable click target size
	if minCellSize <= 0 {
		minCellSize = 40 // default
	}

	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Calculate how many cells fit in each dimension (uniform sizes that divide screen exactly)
	targetCols := width / minCellSize
	targetRows := height / minCellSize
	if targetCols < 1 {
		targetCols = 1
	}
	if targetRows < 1 {
		targetRows = 1
	}
	findDiv := func(total, target, minSize int) int {
		for d := target; d >= 1; d-- {
			if total%d == 0 && total/d >= minSize {
				return d
			}
		}
		return 1
	}
	gridCols := findDiv(width, targetCols, minCellSize)
	gridRows := findDiv(height, targetRows, minCellSize)

	// Calculate total cells needed to fill screen
	totalCells := gridRows * gridCols

	// Determine optimal label length based on total cells
	// numChars^2 = 2-char labels (e.g., 9^2 = 81)
	// numChars^3 = 3-char labels (e.g., 9^3 = 729)
	// numChars^4 = 4-char labels (e.g., 9^4 = 6561)
	var labelLength int
	if totalCells <= numChars*numChars {
		labelLength = 2
	} else if totalCells <= numChars*numChars*numChars {
		labelLength = 3
	} else {
		labelLength = 4
	}

	// Generate exactly totalCells labels
	labels := generateLabels(chars, numChars, totalCells, labelLength)

	// Calculate actual cell size (distribute evenly to fill screen)
	cellWidth := width / gridCols
	cellHeight := height / gridRows

	// Create cells distributed across the entire screen
	cells := make([]*Cell, 0, totalCells)
	idx := 0

	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			if idx >= len(labels) {
				break
			}

			x := bounds.Min.X + col*cellWidth
			y := bounds.Min.Y + row*cellHeight

			cell := &Cell{
				Coordinate: labels[idx],
				Bounds: image.Rect(
					x, y,
					x+cellWidth, y+cellHeight,
				),
				Center: image.Point{
					X: x + cellWidth/2,
					Y: y + cellHeight/2,
				},
			}
			cells = append(cells, cell)
			idx++
		}
	}

	return &Grid{
		characters: characters,
		bounds:     bounds,
		cells:      cells,
	}
}

// generateLabels creates labels of the specified length for the given number of cells
func generateLabels(chars []rune, numChars, count, labelLength int) []string {
	labels := make([]string, 0, count)

	switch labelLength {
	case 2:
		// Generate 2-char labels: AA, AB, AC, ...
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				label := string(chars[i]) + string(chars[j])
				labels = append(labels, label)
			}
		}

	case 3:
		// Generate 3-char labels: AAA, AAB, AAC, ...
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				for k := 0; k < numChars && len(labels) < count; k++ {
					label := string(chars[i]) + string(chars[j]) + string(chars[k])
					labels = append(labels, label)
				}
			}
		}

	case 4:
		// Generate 4-char labels: AAAA, AAAB, AAAC, ...
		for i := 0; i < numChars && len(labels) < count; i++ {
			for j := 0; j < numChars && len(labels) < count; j++ {
				for k := 0; k < numChars && len(labels) < count; k++ {
					for l := 0; l < numChars && len(labels) < count; l++ {
						label := string(chars[i]) + string(chars[j]) + string(chars[k]) + string(chars[l])
						labels = append(labels, label)
					}
				}
			}
		}
	}

	return labels
}

// GetAllCells returns all grid cells
func (g *Grid) GetAllCells() []*Cell {
	return g.cells
}

// GetCellByCoordinate returns the cell for a given coordinate (2, 3, or 4 characters)
func (g *Grid) GetCellByCoordinate(coord string) *Cell {
	coord = strings.ToUpper(coord)

	// Linear search through cells
	for _, cell := range g.cells {
		if cell.Coordinate == coord {
			return cell
		}
	}

	return nil
}

// CalculateOptimalGrid calculates optimal character count for coverage
func CalculateOptimalGrid(characters string) (rows, cols int) {
	// For flat 3-char grid, we don't use rows/cols
	// Just return sensible defaults (will be ignored)
	numChars := len(characters)
	if numChars < 2 {
		numChars = 9
	}
	return numChars, numChars
}
