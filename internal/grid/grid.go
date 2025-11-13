package grid

import (
	"image"
	"math"
	"strings"

	"go.uber.org/zap"
)

// Grid represents a flat 3-character coordinate grid (all positions visible at once)
type Grid struct {
	characters string          // Characters used for coordinates (e.g., "asdfghjkl")
	bounds     image.Rectangle // Screen bounds
	cells      []*Cell         // All cells with 3-char coordinates
	index      map[string]*Cell
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
// Grid layout uses spatial regions for predictable navigation:
//   - Each region is identified by the first character (Region A, Region B, etc.)
//   - Within each region, coordinates flow left-to-right, top-to-bottom
//   - Region A: AAA, ABA, ACA (left-to-right), then AAB, ABB, ACB (next row)
//   - Regions flow left-to-right until screen width is filled
//   - Next region starts on new row below, continuing the pattern
//   - This allows users to think: "C** coordinates are in region C on the screen"
//
// Cell sizing is fully automatic based on screen characteristics:
//   - Very small screens (<1.5M pixels): 25-60px cells for maximum precision
//   - Small-medium screens (1.5-2.5M pixels): 30-80px cells
//   - Medium-large screens (2.5-4M pixels): 40-100px cells
//   - Very large screens (>4M pixels): 50-120px cells
func NewGrid(characters string, bounds image.Rectangle, logger *zap.Logger) *Grid {
	logger.Debug("Creating new grid",
		zap.String("characters", characters),
		zap.Int("bounds_width", bounds.Dx()),
		zap.Int("bounds_height", bounds.Dy()))

	if characters == "" {
		characters = "abcdefghijklmnopqrstuvwxyz"
	}
	characters = strings.ToUpper(characters)
	chars := []rune(characters)
	numChars := len(chars)

	// Ensure we have valid characters
	if numChars < 2 {
		characters = "abcdefghijklmnopqrstuvwxyz"
		chars = []rune(characters)
		numChars = len(chars)
	}

	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	logger.Debug("Grid dimensions calculated",
		zap.Int("width", width),
		zap.Int("height", height))

	if gridCacheEnabled {
		if cells, ok := gridCache.get(characters, bounds); ok {
			logger.Debug("Grid cache hit",
				zap.Int("cell_count", len(cells)))
			return &Grid{characters: characters, bounds: bounds, cells: cells}
		}
		logger.Debug("Grid cache miss")
	}

	if width <= 0 || height <= 0 {
		logger.Warn("Invalid grid bounds, creating minimal grid",
			zap.Int("width", width),
			zap.Int("height", height))
		return &Grid{
			characters: characters,
			bounds:     bounds,
			cells:      []*Cell{},
		}
	}

	// Automatically determine optimal cell size constraints based on screen characteristics
	screenArea := width * height
	screenAspect := float64(width) / float64(height)

	var minCellSize, maxCellSize int

	// Calculate optimal cell size ranges based on screen size and pixel density
	if screenArea < 1500000 {
		minCellSize = 30
		maxCellSize = 60
	} else if screenArea < 2500000 {
		minCellSize = 30
		maxCellSize = 80
	} else if screenArea < 4000000 {
		minCellSize = 40
		maxCellSize = 100
	} else {
		minCellSize = 50
		maxCellSize = 120
	}

	// Adjust cell size constraints for extreme aspect ratios
	if screenAspect > 2.5 || screenAspect < 0.4 {
		maxCellSize = int(float64(maxCellSize) * 1.2)
	}

	// Find all valid grid configurations and pick the one with best aspect ratio match
	type candidate struct {
		cols, rows   int
		cellW, cellH int
		score        float64
	}

	var candidates []candidate

	// Calculate search ranges
	minCols := width / maxCellSize
	if minCols < 1 {
		minCols = 1
	}
	maxCols := width / minCellSize
	if maxCols < 1 {
		maxCols = 1
	}

	minRows := height / maxCellSize
	if minRows < 1 {
		minRows = 1
	}
	maxRows := height / minCellSize
	if maxRows < 1 {
		maxRows = 1
	}

	// Search through all valid grid configurations
	for c := maxCols; c >= minCols && c >= 1; c-- {
		cellW := width / c
		if cellW < minCellSize || cellW > maxCellSize {
			continue
		}

		for r := maxRows; r >= minRows && r >= 1; r-- {
			cellH := height / r
			if cellH < minCellSize || cellH > maxCellSize {
				continue
			}

			// Calculate cell aspect ratio
			cellAspect := float64(cellW) / float64(cellH)

			// Score based on how close cell is to being square
			aspectDiff := cellAspect - 1.0
			if aspectDiff < 0 {
				aspectDiff = -aspectDiff
			}

			// Also consider cell count - prefer more cells for better precision
			totalCells := float64(c * r)
			maxCells := float64(maxCols * maxRows)
			cellScore := (maxCells - totalCells) / maxCells * 0.1

			aspectScore := aspectDiff + cellScore

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
		// Fallback: if no valid candidates, use simple best-fit approach
		findBestFit := func(dimension, minSize, maxSize int) int {
			count := max(dimension/minSize, 1)
			for dimension/count > maxSize {
				count++
			}
			return count
		}
		gridCols = findBestFit(width, minCellSize, maxCellSize)
		gridRows = findBestFit(height, minCellSize, maxCellSize)
	}

	// Safety check: ensure we always have at least a 2x2 grid
	if gridCols < 2 {
		gridCols = 2
	}
	if gridRows < 2 {
		gridRows = 2
	}

	// Calculate total cells needed to fill screen
	totalCells := gridRows * gridCols

	// Calculate maximum possible cells we can label
	maxPossibleCells := numChars * numChars * numChars * numChars

	// Cap totalCells to what we can actually label
	if totalCells > maxPossibleCells {
		totalCells = maxPossibleCells
		gridCols = max(int(math.Sqrt(float64(totalCells)*float64(width)/float64(height))), 1)
		gridRows = max(totalCells/gridCols, 1)
		totalCells = gridRows * gridCols
	}

	// Determine optimal label length based on total cells
	var labelLength int
	if totalCells <= numChars*numChars {
		labelLength = 2
	} else if totalCells <= numChars*numChars*numChars {
		labelLength = 3
	} else {
		labelLength = 4
	}

	// Calculate base cell sizes and remainders
	baseCellWidth := width / gridCols
	baseCellHeight := height / gridRows
	remainderWidth := width % gridCols
	remainderHeight := height % gridRows

	// Generate cells with spatial region logic
	cells := generateCellsWithRegions(chars, numChars, gridCols, gridRows, labelLength,
		bounds, baseCellWidth, baseCellHeight, remainderWidth, remainderHeight, logger)

	logger.Debug("Grid created successfully",
		zap.Int("cell_count", len(cells)),
		zap.Int("grid_cols", gridCols),
		zap.Int("grid_rows", gridRows),
		zap.Int("label_length", labelLength))

	if gridCacheEnabled {
		gridCache.put(characters, bounds, cells)
		logger.Debug("Grid cache store",
			zap.Int("cell_count", len(cells)))
	}

	idx := make(map[string]*Cell, len(cells))
	for _, c := range cells {
		idx[c.Coordinate] = c
	}

	return &Grid{
		characters: characters,
		bounds:     bounds,
		cells:      cells,
		index:      idx,
	}
}

// generateCellsWithRegions creates cells using spatial region logic
// Each region (identified by first char) fills left-to-right, top-to-bottom
// Regions flow across screen, wrapping to next row when width is exhausted
func generateCellsWithRegions(chars []rune, numChars, gridCols, gridRows, labelLength int,
	bounds image.Rectangle, baseCellWidth, baseCellHeight, remainderWidth, remainderHeight int,
	logger *zap.Logger,
) []*Cell {
	logger.Debug("Generating cells with regions",
		zap.Int("num_chars", numChars),
		zap.Int("grid_cols", gridCols),
		zap.Int("grid_rows", gridRows),
		zap.Int("label_length", labelLength))

	cells := make([]*Cell, 0, gridCols*gridRows)

	// Calculate region dimensions (how many cols/rows per region)
	// Each region is a sub-grid of size numChars x numChars
	var regionCols, regionRows int

	// Adjust region size based on label length
	switch labelLength {
	case 2:
		// For 2-char labels: each region is numChars x numChars
		regionCols = numChars
		regionRows = numChars
	case 3:
		// For 3-char labels: first char = region, next 2 chars = position
		// Region is numChars wide x numChars tall
		regionCols = numChars
		regionRows = numChars
	default:
		// For 4-char labels: first 2 chars could represent super-regions
		regionCols = numChars
		regionRows = numChars
	}

	// Track current position as we fill regions
	currentCol := 0
	currentRow := 0

	// Iterate through regions (first character)
	regionIndex := 0
	maxRegions := numChars * numChars // Maximum regions we might need

	// Precompute x/y starts to avoid inner summation loops
	xStarts := make([]int, gridCols)
	yStarts := make([]int, gridRows)
	for i := 0; i < gridCols; i++ {
		xStarts[i] = bounds.Min.X + i*baseCellWidth
		if i < remainderWidth {
			xStarts[i] += i
		} else {
			xStarts[i] += remainderWidth
		}
	}
	for j := 0; j < gridRows; j++ {
		yStarts[j] = bounds.Min.Y + j*baseCellHeight
		if j < remainderHeight {
			yStarts[j] += j
		} else {
			yStarts[j] += remainderHeight
		}
	}

	for regionIndex < maxRegions && currentRow < gridRows {
		// Determine region identifier (first character)
		var regionChar1, regionChar2 rune
		switch labelLength {
		case 2:
			regionChar1 = chars[regionIndex%numChars]
		case 3:
			regionChar1 = chars[regionIndex%numChars]
		default: // 4 chars
			regionChar1 = chars[regionIndex/numChars%numChars]
			regionChar2 = chars[regionIndex%numChars]
		}

		// Calculate how many columns this region can occupy
		colsAvailable := gridCols - currentCol
		colsForRegion := min(regionCols, colsAvailable)

		// Calculate how many rows this region can occupy
		rowsAvailable := gridRows - currentRow
		rowsForRegion := min(regionRows, rowsAvailable)

		// Fill this region
		for r := range rowsForRegion {
			for c := range colsForRegion {
				globalCol := currentCol + c
				globalRow := currentRow + r

				if globalCol >= gridCols || globalRow >= gridRows {
					break
				}

				// Generate coordinate for this cell
				// Second char = column within region, third char = row within region
				var coord string
				switch labelLength {
				case 2:
					coord = string(regionChar1) + string(chars[c])
				case 3:
					// First char = region, second char = column, third char = row
					char2 := chars[c%numChars] // column
					char3 := chars[r%numChars] // row
					coord = string(regionChar1) + string(char2) + string(char3)
				default: // 4 chars
					// First 2 chars = region, third char = column, fourth char = row
					char3 := chars[c%numChars] // column
					char4 := chars[r%numChars] // row
					coord = string(regionChar1) + string(regionChar2) + string(char3) + string(char4)
				}

				// Calculate cell dimensions with remainder distribution
				cellWidth := baseCellWidth
				if globalCol < remainderWidth {
					cellWidth++
				}
				cellHeight := baseCellHeight
				if globalRow < remainderHeight {
					cellHeight++
				}

				x := xStarts[globalCol]
				y := yStarts[globalRow]

				cell := &Cell{
					Coordinate: coord,
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
			}
		}

		// Move to next region position
		currentCol += colsForRegion

		// If we've filled the row width, move to next row
		if currentCol >= gridCols {
			currentCol = 0
			currentRow += rowsForRegion
		}

		regionIndex++

		// Stop if we've filled the entire screen
		if len(cells) >= gridCols*gridRows {
			break
		}
	}

	return cells
}

// GetAllCells returns all grid cells
func (g *Grid) GetAllCells() []*Cell {
	return g.cells
}

// GetCellByCoordinate returns the cell for a given coordinate (2, 3, or 4 characters)
func (g *Grid) GetCellByCoordinate(coord string) *Cell {
	coord = strings.ToUpper(coord)

	if g.index != nil {
		if cell, ok := g.index[coord]; ok {
			return cell
		}
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
