package grid

import (
	"image"
	"strings"

	"go.uber.org/zap"
)

// Manager handles variable-length grid input (2, 3, or 4 characters)
type Manager struct {
	grid         *Grid
	currentInput string
	labelLength  int // Length of labels (2, 3, or 4)
	onUpdate     func()
	onShowSub    func(cell *Cell)
	inSubgrid    bool
	selectedCell *Cell
	// Subgrid configuration
	subRows int
	subCols int
	subKeys string
	logger  *zap.Logger
}

// NewManager creates a new grid manager
func NewManager(grid *Grid, subRows int, subCols int, subKeys string, onUpdate func(), onShowSub func(cell *Cell), logger *zap.Logger) *Manager {
	// Determine label length from first cell (if grid exists)
	labelLength := 3 // Default
	if grid != nil && len(grid.cells) > 0 {
		labelLength = len(grid.cells[0].Coordinate)
	}

	return &Manager{
		grid:        grid,
		labelLength: labelLength,
		onUpdate:    onUpdate,
		onShowSub:   onShowSub,
		subRows:     subRows,
		subCols:     subCols,
		subKeys:     strings.ToUpper(strings.TrimSpace(subKeys)),
		logger:      logger,
	}
}

// HandleInput processes variable-length coordinate input
// Returns the target point when complete (labelLength characters entered or subgrid selection)
func (m *Manager) HandleInput(key string) (targetPoint image.Point, complete bool) {
	m.logger.Debug("Grid manager: Processing input", zap.String("key", key), zap.String("current_input", m.currentInput))

	// Handle backspace
	if key == "\x7f" || key == "delete" || key == "backspace" {
		if len(m.currentInput) > 0 {
			m.currentInput = m.currentInput[:len(m.currentInput)-1]
			m.logger.Debug("Grid manager: Backspace processed", zap.String("new_input", m.currentInput))
			if m.onUpdate != nil {
				m.onUpdate()
			}
			return image.Point{}, false
		}
		// If in subgrid, backspace exits subgrid
		if m.inSubgrid {
			m.logger.Debug("Grid manager: Exiting subgrid on backspace")
			m.inSubgrid = false
			m.selectedCell = nil
			if m.onUpdate != nil {
				m.onUpdate()
			}
		}
		return image.Point{}, false
	}

	// Ignore non-letter keys
	if len(key) != 1 || !isLetter(key[0]) {
		m.logger.Debug("Grid manager: Ignoring non-letter key", zap.String("key", key))
		return image.Point{}, false
	}

	key = strings.ToUpper(key)

	// If we're in subgrid selection, next key chooses a subcell
	if m.inSubgrid && m.selectedCell != nil {
		idx := strings.Index(m.subKeys, key)
		if idx < 0 {
			m.logger.Debug("Grid manager: Invalid subgrid key", zap.String("key", key))
			return image.Point{}, false
		}
		// Subgrid is always 3x3
		if idx >= 9 {
			m.logger.Debug("Grid manager: Subgrid index out of range", zap.Int("index", idx))
			return image.Point{}, false
		}
		row := idx / m.subCols
		col := idx % m.subCols
		b := m.selectedCell.Bounds
		// Compute breakpoints to match overlay splitting and cover full bounds
		xBreaks := make([]int, m.subCols+1)
		yBreaks := make([]int, m.subRows+1)
		xBreaks[0] = b.Min.X
		yBreaks[0] = b.Min.Y
		for i := 1; i <= m.subCols; i++ {
			val := float64(i) * float64(b.Dx()) / float64(m.subCols)
			xBreaks[i] = b.Min.X + int(val+0.5)
		}
		for j := 1; j <= m.subRows; j++ {
			val := float64(j) * float64(b.Dy()) / float64(m.subRows)
			yBreaks[j] = b.Min.Y + int(val+0.5)
		}
		// Ensure exact coverage
		xBreaks[m.subCols] = b.Max.X
		yBreaks[m.subRows] = b.Max.Y
		left := xBreaks[col]
		right := xBreaks[col+1]
		top := yBreaks[row]
		bottom := yBreaks[row+1]
		x := left + (right-left)/2
		y := top + (bottom-top)/2
		m.logger.Info("Grid manager: Subgrid selection complete",
			zap.Int("row", row), zap.Int("col", col),
			zap.Int("x", x), zap.Int("y", y))
		m.Reset()
		return image.Point{X: x, Y: y}, true
	}

	// Check if character is valid for grid
	if m.grid != nil && !strings.Contains(m.grid.characters, key) {
		m.logger.Debug("Grid manager: Character not valid for grid", zap.String("key", key))
		return image.Point{}, false
	}

	// Check if this key could potentially lead to a valid coordinate
	// by checking if there's any cell that starts with currentInput + key
	potentialInput := m.currentInput + key
	validPrefix := false
	for _, cell := range m.grid.cells {
		if len(cell.Coordinate) >= len(potentialInput) &&
			strings.HasPrefix(cell.Coordinate, potentialInput) {
			validPrefix = true
			break
		}
	}

	// If this key doesn't lead to any valid coordinate, ignore it
	if !validPrefix {
		m.logger.Debug("Grid manager: Key does not lead to valid coordinate", zap.String("input", potentialInput))
		return image.Point{}, false
	}

	m.currentInput += key
	m.logger.Debug("Grid manager: Input accumulated", zap.String("current_input", m.currentInput))

	// After reaching label length, show subgrid inside the selected cell
	if len(m.currentInput) >= m.labelLength {
		coord := m.currentInput[:m.labelLength]
		if m.grid != nil {
			cell := m.grid.GetCellByCoordinate(coord)
			if cell != nil {
				m.inSubgrid = true
				m.selectedCell = cell
				m.currentInput = ""
				m.logger.Debug("Grid manager: Showing subgrid for cell", zap.String("coordinate", coord))
				if m.onShowSub != nil {
					m.onShowSub(cell)
				}
				return image.Point{}, false
			}
		}
		// Invalid coordinate, reset
		m.logger.Debug("Grid manager: Invalid coordinate, resetting", zap.String("coordinate", coord))
		m.Reset()
		return image.Point{}, false
	}

	// Update overlay to show matched cells
	if m.onUpdate != nil {
		m.onUpdate()
	}
	return image.Point{}, false
}

// GetInput returns the current partial coordinate input
func (m *Manager) GetInput() string {
	return m.currentInput
}

// GetCurrentGrid returns the grid
func (m *Manager) GetCurrentGrid() *Grid {
	return m.grid
}

// Reset resets the input state
func (m *Manager) Reset() {
	m.currentInput = ""
	m.inSubgrid = false
	m.selectedCell = nil
	m.logger.Debug("Grid manager: Resetting input state")
	if m.onUpdate != nil {
		m.onUpdate()
	}
}

// GetGrid returns the grid
func (m *Manager) GetGrid() *Grid {
	return m.grid
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
