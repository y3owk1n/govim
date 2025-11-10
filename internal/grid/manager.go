package grid

import (
	"image"
	"strings"
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
}

// NewManager creates a new grid manager
func NewManager(grid *Grid, onUpdate func(), onShowSub func(cell *Cell)) *Manager {
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
	}
}

// HandleInput processes variable-length coordinate input
// Returns the target point when complete (labelLength characters entered or subgrid selection)
func (m *Manager) HandleInput(key string) (targetPoint image.Point, complete bool) {
	// Handle backspace
	if key == "\x7f" || key == "delete" || key == "backspace" {
		if len(m.currentInput) > 0 {
			m.currentInput = m.currentInput[:len(m.currentInput)-1]
			if m.onUpdate != nil {
				m.onUpdate()
			}
			return image.Point{}, false
		}
		// If in subgrid, backspace exits subgrid
		if m.inSubgrid {
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
		return image.Point{}, false
	}

	key = strings.ToUpper(key)

	// If we're in subgrid selection, next key chooses a subcell
	if m.inSubgrid && m.selectedCell != nil {
		idx := strings.Index(m.grid.characters, key)
		if idx < 0 {
			return image.Point{}, false
		}
		subRows, subCols := 3, 3
		if idx >= subRows*subCols {
			return image.Point{}, false
		}
		row := idx / subCols
		col := idx % subCols
		b := m.selectedCell.Bounds
		cw := b.Dx() / subCols
		ch := b.Dy() / subRows
		x := b.Min.X + col*cw + cw/2
		y := b.Min.Y + row*ch + ch/2
		m.Reset()
		return image.Point{X: x, Y: y}, true
	}

	// Check if character is valid for grid
	if m.grid != nil && !strings.Contains(m.grid.characters, key) {
		return image.Point{}, false
	}

	m.currentInput += key

	// After reaching label length, show subgrid inside the selected cell
	if len(m.currentInput) >= m.labelLength {
		coord := m.currentInput[:m.labelLength]
		if m.grid != nil {
			cell := m.grid.GetCellByCoordinate(coord)
			if cell != nil {
				m.inSubgrid = true
				m.selectedCell = cell
				m.currentInput = ""
				if m.onShowSub != nil {
					m.onShowSub(cell)
				}
				return image.Point{}, false
			}
		}
		// Invalid coordinate, reset
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
