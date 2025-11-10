package grid

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/y3owk1n/neru/internal/config"
)

// GridOverlay manages grid-specific overlay rendering
type GridOverlay struct {
	window C.OverlayWindow
	cfg    config.GridConfig
}

// NewGridOverlay creates a new grid overlay with its own window
func NewGridOverlay(cfg config.GridConfig) *GridOverlay {
	window := C.createOverlayWindow()
	return &GridOverlay{
		window: window,
		cfg:    cfg,
	}
}

// UpdateConfig updates the overlay's config (e.g., after config reload)
func (o *GridOverlay) UpdateConfig(cfg config.GridConfig) {
	o.cfg = cfg
}

// Show displays the grid overlay
func (o *GridOverlay) Show() {
	C.showOverlayWindow(o.window)
}

// Hide hides the grid overlay
func (o *GridOverlay) Hide() {
	C.hideOverlayWindow(o.window)
}

// Clear clears the grid overlay
func (o *GridOverlay) Clear() {
	C.clearOverlay(o.window)
}

// Destroy destroys the grid overlay window
func (o *GridOverlay) Destroy() {
	C.destroyOverlayWindow(o.window)
}

// Draw renders the flat grid with all 3-char cells visible
func (o *GridOverlay) Draw(grid *Grid, currentInput string) error {
	// Clear existing content
	o.Clear()

	cells := grid.GetAllCells()
	if len(cells) == 0 {
		return nil
	}

	// Draw grid cells with labels (no grid lines for dense flat view)
	o.drawGridCells(cells, currentInput)

	return nil
}

// UpdateMatches updates matched state without redrawing all cells
func (o *GridOverlay) UpdateMatches(prefix string) {
	cPrefix := C.CString(prefix)
	defer C.free(unsafe.Pointer(cPrefix))
	C.updateGridMatchPrefix(o.window, cPrefix)
}

// ShowSubgrid draws a 3x3 subgrid inside the selected cell
func (o *GridOverlay) ShowSubgrid(cell *Cell) {
	keys := o.cfg.SublayerKeys
	if strings.TrimSpace(keys) == "" {
		keys = o.cfg.Characters
	}
	chars := []rune(keys)
	rows, cols := o.cfg.SubgridRows, o.cfg.SubgridCols
	if rows < 1 {
		rows = 3
	}
	if cols < 1 {
		cols = 3
	}
	count := rows * cols
	if len(chars) < count {
		count = len(chars)
	}

	cells := make([]C.GridCell, count)
	labels := make([]*C.char, count)

	b := cell.Bounds
	cw := b.Dx() / cols
	ch := b.Dy() / rows

	for i := 0; i < count; i++ {
		r := i / cols
		c := i % cols
		label := string(chars[i])
		labels[i] = C.CString(label)
		cells[i] = C.GridCell{
			label: labels[i],
			bounds: C.CGRect{
				origin: C.CGPoint{x: C.double(b.Min.X + c*cw), y: C.double(b.Min.Y + r*ch)},
				size:   C.CGSize{width: C.double(cw), height: C.double(ch)},
			},
			isMatched: C.int(0),
		}
	}

	fontFamily := C.CString(o.cfg.FontFamily)
	backgroundColor := C.CString(o.cfg.BackgroundColor)
	textColor := C.CString(o.cfg.TextColor)
	matchedTextColor := C.CString(o.cfg.MatchedTextColor)
	matchedBackgroundColor := C.CString(o.cfg.MatchedBackgroundColor)
	matchedBorderColor := C.CString(o.cfg.MatchedBorderColor)
	borderColor := C.CString(o.cfg.BorderColor)

	style := C.GridCellStyle{
		fontSize:               C.int(o.cfg.FontSize),
		fontFamily:             fontFamily,
		backgroundColor:        backgroundColor,
		textColor:              textColor,
		matchedTextColor:       matchedTextColor,
		matchedBackgroundColor: matchedBackgroundColor,
		matchedBorderColor:     matchedBorderColor,
		borderColor:            borderColor,
		borderWidth:            C.int(o.cfg.BorderWidth),
		backgroundOpacity:      C.double(o.cfg.Opacity),
		textOpacity:            C.double(1.0),
	}

	C.clearOverlay(o.window)
	C.drawGridCells(o.window, &cells[0], C.int(len(cells)), style)

	for i := range labels {
		C.free(unsafe.Pointer(labels[i]))
	}
	C.free(unsafe.Pointer(fontFamily))
	C.free(unsafe.Pointer(backgroundColor))
	C.free(unsafe.Pointer(textColor))
	C.free(unsafe.Pointer(matchedTextColor))
	C.free(unsafe.Pointer(matchedBackgroundColor))
	C.free(unsafe.Pointer(matchedBorderColor))
	C.free(unsafe.Pointer(borderColor))
}

// DrawScrollHighlight draws a scroll highlight
func (o *GridOverlay) DrawScrollHighlight(x, y, w, h int, color string, width int) {
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	rect := C.CGRect{
		origin: C.CGPoint{x: C.double(x), y: C.double(y)},
		size:   C.CGSize{width: C.double(w), height: C.double(h)},
	}
	C.drawScrollHighlight(o.window, rect, cColor, C.int(width))
}

// drawGridCells draws all grid cells with their labels
func (o *GridOverlay) drawGridCells(cellsGo []*Cell, currentInput string) {
	cGridCells := make([]C.GridCell, len(cellsGo))
	cLabels := make([]*C.char, len(cellsGo))

	for i, cell := range cellsGo {
		cLabels[i] = C.CString(cell.Coordinate)

		isMatched := 0
		if currentInput != "" && len(cell.Coordinate) >= len(currentInput) {
			cellPrefix := cell.Coordinate[:len(currentInput)]
			if cellPrefix == currentInput {
				isMatched = 1
			}
		}

		cGridCells[i] = C.GridCell{
			label: cLabels[i],
			bounds: C.CGRect{
				origin: C.CGPoint{x: C.double(cell.Bounds.Min.X), y: C.double(cell.Bounds.Min.Y)},
				size:   C.CGSize{width: C.double(cell.Bounds.Dx()), height: C.double(cell.Bounds.Dy())},
			},
			isMatched: C.int(isMatched),
		}
	}

	fontFamily := C.CString(o.cfg.FontFamily)
	backgroundColor := C.CString(o.cfg.BackgroundColor)
	textColor := C.CString(o.cfg.TextColor)
	matchedTextColor := C.CString(o.cfg.MatchedTextColor)
	matchedBackgroundColor := C.CString(o.cfg.MatchedBackgroundColor)
	matchedBorderColor := C.CString(o.cfg.MatchedBorderColor)
	borderColor := C.CString(o.cfg.BorderColor)

	style := C.GridCellStyle{
		fontSize:               C.int(o.cfg.FontSize),
		fontFamily:             fontFamily,
		backgroundColor:        backgroundColor,
		textColor:              textColor,
		matchedTextColor:       matchedTextColor,
		matchedBackgroundColor: matchedBackgroundColor,
		matchedBorderColor:     matchedBorderColor,
		borderColor:            borderColor,
		borderWidth:            C.int(o.cfg.BorderWidth),
		backgroundOpacity:      C.double(o.cfg.Opacity),
		textOpacity:            C.double(1.0),
	}

	C.drawGridCells(o.window, &cGridCells[0], C.int(len(cGridCells)), style)

	for _, cLabel := range cLabels {
		C.free(unsafe.Pointer(cLabel))
	}
	C.free(unsafe.Pointer(fontFamily))
	C.free(unsafe.Pointer(backgroundColor))
	C.free(unsafe.Pointer(textColor))
	C.free(unsafe.Pointer(matchedTextColor))
	C.free(unsafe.Pointer(matchedBackgroundColor))
	C.free(unsafe.Pointer(matchedBorderColor))
	C.free(unsafe.Pointer(borderColor))
}
