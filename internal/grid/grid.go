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

// NewGrid creates a grid with automatically optimized cell sizes for the screen.
// Cell sizes are dynamically calculated based on screen dimensions, resolution, and aspect ratio
// to ensure optimal precision and usability across all display types.
//
// Cell sizing is fully automatic based on screen characteristics:
//   - Very small screens (<1.5M pixels): 25-60px cells for maximum precision
//   - Small-medium screens (1.5-2.5M pixels): 30-80px cells
//   - Medium-large screens (2.5-4M pixels): 40-100px cells
//   - Very large screens (>4M pixels): 50-120px cells
//
// The algorithm ensures cells maintain aspect ratios matching the screen to provide
// consistent rectangular proportions across all monitors.
func NewGrid(characters string, bounds image.Rectangle) *Grid {
	if characters == "" {
		characters = "asdfghjkl"
	}
	characters = strings.ToUpper(characters)
	chars := []rune(characters)
	numChars := len(chars)

	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Automatically determine optimal cell size constraints based on screen characteristics
	// This ensures consistent precision and usability across all display types
	screenArea := width * height
	screenAspect := float64(width) / float64(height)

	// Declare cell size constraints
	var minCellSize, maxCellSize int

	// Calculate optimal cell size ranges based on screen size and pixel density
	// Smaller screens need smaller cells for precision, larger screens can have bigger cells
	if screenArea < 1500000 { // Very small screens (< ~1500x1000)
		minCellSize = 30
		maxCellSize = 60
	} else if screenArea < 2500000 { // Small to medium screens (~1500x1000 to ~2000x1250)
		minCellSize = 30
		maxCellSize = 80
	} else if screenArea < 4000000 { // Medium to large screens (~1920x1080 to ~2560x1600)
		minCellSize = 40
		maxCellSize = 100
	} else { // Very large screens (4K+, ultra-wide)
		minCellSize = 50
		maxCellSize = 120
	}

	// Calculate target grid dimensions based on optimal cell sizes
	targetCols := width / minCellSize
	targetRows := height / minCellSize
	if targetCols < 1 {
		targetCols = 1
	}
	if targetRows < 1 {
		targetRows = 1
	}
	// Ensure cells don't exceed maximum size by increasing grid density
	if width/targetCols > maxCellSize {
		targetCols = width / maxCellSize
		if targetCols < 1 {
			targetCols = 1
		}
	}
	if height/targetRows > maxCellSize {
		targetRows = height / maxCellSize
		if targetRows < 1 {
			targetRows = 1
		}
	}

	// Find all valid grid configurations and pick the one with best aspect ratio match
	type candidate struct {
		cols, rows   int
		cellW, cellH int
		score        float64
	}

	var candidates []candidate

	// Search through all valid grid configurations
	// Don't require exact divisibility - use integer division for best fit
	for c := targetCols; c >= 1; c-- {
		cellW := width / c
		if cellW < minCellSize || cellW > maxCellSize {
			continue
		}

		for r := targetRows; r >= 1; r-- {
			cellH := height / r
			if cellH < minCellSize || cellH > maxCellSize {
				continue
			}

			// Calculate cell aspect ratio
			cellAspect := float64(cellW) / float64(cellH)

			// Score based on how close cell aspect matches screen aspect
			// Lower score is better
			aspectDiff := cellAspect - screenAspect
			if aspectDiff < 0 {
				aspectDiff = -aspectDiff
			}

			// Normalize by screen aspect to make it scale-independent
			aspectScore := aspectDiff / screenAspect

			candidates = append(candidates, candidate{
				cols:  c,
				rows:  r,
				cellW: cellW,
				cellH: cellH,
				score: aspectScore,
			})
		}
	}

	// Pick the candidate with the best (lowest) score
	var gridCols, gridRows int
	if len(candidates) > 0 {
		best := candidates[0]
		for _, cand := range candidates[1:] {
			if cand.score < best.score {
				best = cand
			}
		}
		gridCols = best.cols
		gridRows = best.rows
	} else {
		// Fallback: use simple divisor finding if no valid candidates
		findDiv := func(total, target, minSize int) int {
			for d := target; d >= 1; d-- {
				if total%d == 0 && total/d >= minSize {
					return d
				}
			}
			return 1
		}
		gridCols = findDiv(width, targetCols, minCellSize)
		gridRows = findDiv(height, targetRows, minCellSize)
	}

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

	// Calculate base cell sizes and remainders to ensure complete screen coverage
	// We need to distribute remainder pixels to avoid gaps at screen edges
	baseCellWidth := width / gridCols
	baseCellHeight := height / gridRows
	remainderWidth := width % gridCols   // Extra pixels that need distributing
	remainderHeight := height % gridRows // Extra pixels that need distributing

	// Create cells distributed across the entire screen with remainder distribution
	cells := make([]*Cell, 0, totalCells)
	idx := 0

	for row := 0; row < gridRows; row++ {
		for col := 0; col < gridCols; col++ {
			if idx >= len(labels) {
				break
			}

			// Calculate cell dimensions with remainder distribution
			// Distribute extra pixels evenly across first N cells
			cellWidth := baseCellWidth
			if col < remainderWidth {
				cellWidth++ // Give this column an extra pixel
			}
			cellHeight := baseCellHeight
			if row < remainderHeight {
				cellHeight++ // Give this row an extra pixel
			}

			// Calculate x position by summing widths of all previous columns
			x := bounds.Min.X
			for c := 0; c < col; c++ {
				if c < remainderWidth {
					x += baseCellWidth + 1
				} else {
					x += baseCellWidth
				}
			}

			// Calculate y position by summing heights of all previous rows
			y := bounds.Min.Y
			for r := 0; r < row; r++ {
				if r < remainderHeight {
					y += baseCellHeight + 1
				} else {
					y += baseCellHeight
				}
			}

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
