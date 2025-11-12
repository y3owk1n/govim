package grid

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>

// Callback function that Go can reference
extern void gridResizeCompletionCallback(void* context);
*/
import "C"

import (
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

var (
	gridCallbackID   uint64
	gridCallbackMap  = make(map[uint64]chan struct{})
	gridCallbackLock sync.Mutex
)

//export gridResizeCompletionCallback
func gridResizeCompletionCallback(context unsafe.Pointer) {
	// Convert context to callback ID
	id := uint64(uintptr(context))

	gridCallbackLock.Lock()
	if done, ok := gridCallbackMap[id]; ok {
		close(done)
		delete(gridCallbackMap, id)
	}
	gridCallbackLock.Unlock()
}

// GridOverlay manages grid-specific overlay rendering
type GridOverlay struct {
	window C.OverlayWindow
	cfg    config.GridConfig
	logger *zap.Logger
}

// NewGridOverlay creates a new grid overlay with its own window
func NewGridOverlay(cfg config.GridConfig, logger *zap.Logger) *GridOverlay {
	window := C.createOverlayWindow()
	return &GridOverlay{
		window: window,
		cfg:    cfg,
		logger: logger,
	}
}

// UpdateConfig updates the overlay's config (e.g., after config reload)
func (o *GridOverlay) UpdateConfig(cfg config.GridConfig) {
	o.cfg = cfg
}

// SetHideUnmatched sets whether to hide unmatched cells
func (o *GridOverlay) SetHideUnmatched(hide bool) {
	C.setHideUnmatched(o.window, C.int(boolToInt(hide)))
}

// boolToInt converts a boolean to an integer (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
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

// ReplaceWindow atomically replaces the underlying overlay window on the main thread
func (o *GridOverlay) ReplaceWindow() {
	C.replaceOverlayWindow(&o.window)
}

// ResizeToMainScreen resizes the overlay window to the current main screen
func (o *GridOverlay) ResizeToMainScreen() {
	C.resizeOverlayToMainScreen(o.window)
}

// ResizeToActiveScreen resizes the overlay window to the screen containing the mouse cursor
func (o *GridOverlay) ResizeToActiveScreen() {
	C.resizeOverlayToActiveScreen(o.window)
}

// ResizeToActiveScreenSync resizes the overlay window synchronously with callback notification
func (o *GridOverlay) ResizeToActiveScreenSync() {
	done := make(chan struct{})

	// Generate unique ID for this callback
	id := atomic.AddUint64(&gridCallbackID, 1)

	// Store channel in map
	gridCallbackLock.Lock()
	gridCallbackMap[id] = done
	gridCallbackLock.Unlock()

	// Pass ID as context (safe - no Go pointers)
	// Note: uintptr conversion must happen in same expression to satisfy go vet
	C.resizeOverlayToActiveScreenWithCallback(
		o.window,
		(C.ResizeCompletionCallback)(unsafe.Pointer(C.gridResizeCompletionCallback)),
		*(*unsafe.Pointer)(unsafe.Pointer(&id)),
	)

	<-done
}

// Draw renders the flat grid with all 3-char cells visible
func (o *GridOverlay) Draw(grid *Grid, currentInput string, style GridStyle) error {
	o.logger.Debug("Drawing grid overlay",
		zap.Int("cell_count", len(grid.GetAllCells())),
		zap.String("current_input", currentInput))

	// Clear existing content
	o.Clear()

	cells := grid.GetAllCells()
	if len(cells) == 0 {
		o.logger.Debug("No cells to draw in grid overlay")
		return nil
	}

	// Draw grid cells with labels (no grid lines for dense flat view)
	o.drawGridCells(cells, currentInput, style)

	o.logger.Debug("Grid overlay drawn successfully")
	return nil
}

// UpdateMatches updates matched state without redrawing all cells
func (o *GridOverlay) UpdateMatches(prefix string) {
	o.logger.Debug("Updating grid matches", zap.String("prefix", prefix))

	cPrefix := C.CString(prefix)
	defer C.free(unsafe.Pointer(cPrefix))
	C.updateGridMatchPrefix(o.window, cPrefix)

	o.logger.Debug("Grid matches updated successfully")
}

// ShowSubgrid draws a 3x3 subgrid inside the selected cell
func (o *GridOverlay) ShowSubgrid(cell *Cell, style GridStyle) {
	o.logger.Debug("Showing subgrid",
		zap.Int("cell_x", cell.Bounds.Min.X),
		zap.Int("cell_y", cell.Bounds.Min.Y),
		zap.Int("cell_width", cell.Bounds.Dx()),
		zap.Int("cell_height", cell.Bounds.Dy()))

	keys := o.cfg.SublayerKeys
	if strings.TrimSpace(keys) == "" {
		keys = o.cfg.Characters
	}
	chars := []rune(keys)
	// Subgrid is always 3x3
	const rows = 3
	const cols = 3
	count := 9
	if len(chars) < count {
		// If not enough characters, adjust count to available characters
		count = len(chars)
	}

	cells := make([]C.GridCell, count)
	labels := make([]*C.char, count)

	b := cell.Bounds
	// Build breakpoints that evenly distribute remainders to fully cover the cell
	xBreaks := make([]int, cols+1)
	yBreaks := make([]int, rows+1)
	xBreaks[0] = b.Min.X
	yBreaks[0] = b.Min.Y
	for i := 1; i <= cols; i++ {
		// round(i * width / cols)
		val := float64(i) * float64(b.Dx()) / float64(cols)
		xBreaks[i] = b.Min.X + int(val+0.5)
	}
	for j := 1; j <= rows; j++ {
		val := float64(j) * float64(b.Dy()) / float64(rows)
		yBreaks[j] = b.Min.Y + int(val+0.5)
	}

	// Ensure last break exactly matches bounds max to avoid 1px drift
	xBreaks[cols] = b.Max.X
	yBreaks[rows] = b.Max.Y

	for i := 0; i < count; i++ {
		r := i / cols
		c := i % cols
		label := strings.ToUpper(string(chars[i]))
		labels[i] = C.CString(label)
		left := xBreaks[c]
		right := xBreaks[c+1]
		top := yBreaks[r]
		bottom := yBreaks[r+1]
		cells[i] = C.GridCell{
			label: labels[i],
			bounds: C.CGRect{
				origin: C.CGPoint{x: C.double(left), y: C.double(top)},
				size:   C.CGSize{width: C.double(right - left), height: C.double(bottom - top)},
			},
			isMatched: C.int(0),
			isSubgrid: C.int(1), // Mark as subgrid cell
		}
	}

	fontFamily := C.CString(style.FontFamily)
	backgroundColor := C.CString(style.BackgroundColor)
	textColor := C.CString(style.TextColor)
	matchedTextColor := C.CString(style.MatchedTextColor)
	matchedBackgroundColor := C.CString(style.MatchedBackgroundColor)
	matchedBorderColor := C.CString(style.MatchedBorderColor)
	borderColor := C.CString(style.BorderColor)

	finalStyle := C.GridCellStyle{
		fontSize:               C.int(style.FontSize),
		fontFamily:             fontFamily,
		backgroundColor:        backgroundColor,
		textColor:              textColor,
		matchedTextColor:       matchedTextColor,
		matchedBackgroundColor: matchedBackgroundColor,
		matchedBorderColor:     matchedBorderColor,
		borderColor:            borderColor,
		borderWidth:            C.int(style.BorderWidth),
		backgroundOpacity:      C.double(style.Opacity),
		textOpacity:            C.double(1.0),
	}

	C.clearOverlay(o.window)
	C.drawGridCells(o.window, &cells[0], C.int(len(cells)), finalStyle)

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

	o.logger.Debug("Subgrid shown successfully")
}

// DrawScrollHighlight draws a scroll highlight
func (o *GridOverlay) DrawScrollHighlight(x, y, w, h int, color string, width int) {
	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))
	// Build 4 border lines around the rectangle
	lines := make([]C.CGRect, 4)
	// Bottom
	lines[0] = C.CGRect{
		origin: C.CGPoint{x: C.double(x), y: C.double(y)},
		size:   C.CGSize{width: C.double(w), height: C.double(width)},
	}
	// Top
	lines[1] = C.CGRect{
		origin: C.CGPoint{x: C.double(x), y: C.double(y + h - width)},
		size:   C.CGSize{width: C.double(w), height: C.double(width)},
	}
	// Left
	lines[2] = C.CGRect{
		origin: C.CGPoint{x: C.double(x), y: C.double(y)},
		size:   C.CGSize{width: C.double(width), height: C.double(h)},
	}
	// Right
	lines[3] = C.CGRect{
		origin: C.CGPoint{x: C.double(x + w - width), y: C.double(y)},
		size:   C.CGSize{width: C.double(width), height: C.double(h)},
	}
	C.drawGridLines(o.window, &lines[0], C.int(4), cColor, C.int(width), C.double(1.0))
}

// drawGridCells draws all grid cells with their labels
func (o *GridOverlay) drawGridCells(cellsGo []*Cell, currentInput string, style GridStyle) {
	o.logger.Debug("Drawing grid cells",
		zap.Int("cell_count", len(cellsGo)),
		zap.String("current_input", currentInput))

	cGridCells := make([]C.GridCell, len(cellsGo))
	cLabels := make([]*C.char, len(cellsGo))

	matchedCount := 0
	for i, cell := range cellsGo {
		cLabels[i] = C.CString(cell.Coordinate)

		isMatched := 0
		if currentInput != "" && len(cell.Coordinate) >= len(currentInput) {
			cellPrefix := cell.Coordinate[:len(currentInput)]
			if cellPrefix == currentInput {
				isMatched = 1
				matchedCount++
			}
		}

		cGridCells[i] = C.GridCell{
			label: cLabels[i],
			bounds: C.CGRect{
				origin: C.CGPoint{x: C.double(cell.Bounds.Min.X), y: C.double(cell.Bounds.Min.Y)},
				size:   C.CGSize{width: C.double(cell.Bounds.Dx()), height: C.double(cell.Bounds.Dy())},
			},
			isMatched: C.int(isMatched),
			isSubgrid: C.int(0), // Mark as regular grid cell
		}
	}

	o.logger.Debug("Grid cell match statistics",
		zap.Int("total_cells", len(cellsGo)),
		zap.Int("matched_cells", matchedCount))

	fontFamily := C.CString(style.FontFamily)
	backgroundColor := C.CString(style.BackgroundColor)
	textColor := C.CString(style.TextColor)
	matchedTextColor := C.CString(style.MatchedTextColor)
	matchedBackgroundColor := C.CString(style.MatchedBackgroundColor)
	matchedBorderColor := C.CString(style.MatchedBorderColor)
	borderColor := C.CString(style.BorderColor)

	finalStyle := C.GridCellStyle{
		fontSize:               C.int(style.FontSize),
		fontFamily:             fontFamily,
		backgroundColor:        backgroundColor,
		textColor:              textColor,
		matchedTextColor:       matchedTextColor,
		matchedBackgroundColor: matchedBackgroundColor,
		matchedBorderColor:     matchedBorderColor,
		borderColor:            borderColor,
		borderWidth:            C.int(style.BorderWidth),
		backgroundOpacity:      C.double(style.Opacity),
		textOpacity:            C.double(1.0),
	}

	C.drawGridCells(o.window, &cGridCells[0], C.int(len(cGridCells)), finalStyle)

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

	o.logger.Debug("Grid cells drawn successfully")
}

// GridStyle represents the visual style for grid cells
type GridStyle struct {
	FontSize               int
	FontFamily             string
	Opacity                float64
	BorderWidth            int
	BackgroundColor        string
	TextColor              string
	MatchedTextColor       string
	MatchedBackgroundColor string
	MatchedBorderColor     string
	BorderColor            string
}

// BuildStyleForAction returns GridStyle based on action name using the provided config
func BuildStyleForAction(cfg config.GridConfig, action string) GridStyle {
	style := GridStyle{
		FontSize:    cfg.FontSize,
		FontFamily:  cfg.FontFamily,
		Opacity:     cfg.Opacity,
		BorderWidth: cfg.BorderWidth,
	}

	// Override with per-action colors if available
	switch action {
	case "left_click":
		style.BackgroundColor = cfg.LeftClick.BackgroundColor
		style.TextColor = cfg.LeftClick.TextColor
		style.MatchedTextColor = cfg.LeftClick.MatchedTextColor
		style.MatchedBackgroundColor = cfg.LeftClick.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.LeftClick.MatchedBorderColor
		style.BorderColor = cfg.LeftClick.BorderColor
	case "right_click":
		style.BackgroundColor = cfg.RightClick.BackgroundColor
		style.TextColor = cfg.RightClick.TextColor
		style.MatchedTextColor = cfg.RightClick.MatchedTextColor
		style.MatchedBackgroundColor = cfg.RightClick.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.RightClick.MatchedBorderColor
		style.BorderColor = cfg.RightClick.BorderColor
	case "double_click":
		style.BackgroundColor = cfg.DoubleClick.BackgroundColor
		style.TextColor = cfg.DoubleClick.TextColor
		style.MatchedTextColor = cfg.DoubleClick.MatchedTextColor
		style.MatchedBackgroundColor = cfg.DoubleClick.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.DoubleClick.MatchedBorderColor
		style.BorderColor = cfg.DoubleClick.BorderColor
	case "triple_click":
		style.BackgroundColor = cfg.TripleClick.BackgroundColor
		style.TextColor = cfg.TripleClick.TextColor
		style.MatchedTextColor = cfg.TripleClick.MatchedTextColor
		style.MatchedBackgroundColor = cfg.TripleClick.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.TripleClick.MatchedBorderColor
		style.BorderColor = cfg.TripleClick.BorderColor
	case "middle_click":
		style.BackgroundColor = cfg.MiddleClick.BackgroundColor
		style.TextColor = cfg.MiddleClick.TextColor
		style.MatchedTextColor = cfg.MiddleClick.MatchedTextColor
		style.MatchedBackgroundColor = cfg.MiddleClick.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.MiddleClick.MatchedBorderColor
		style.BorderColor = cfg.MiddleClick.BorderColor
	case "mouse_up":
		style.BackgroundColor = cfg.MouseUp.BackgroundColor
		style.TextColor = cfg.MouseUp.TextColor
		style.MatchedTextColor = cfg.MouseUp.MatchedTextColor
		style.MatchedBackgroundColor = cfg.MouseUp.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.MouseUp.MatchedBorderColor
		style.BorderColor = cfg.MouseUp.BorderColor
	case "mouse_down":
		style.BackgroundColor = cfg.MouseDown.BackgroundColor
		style.TextColor = cfg.MouseDown.TextColor
		style.MatchedTextColor = cfg.MouseDown.MatchedTextColor
		style.MatchedBackgroundColor = cfg.MouseDown.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.MouseDown.MatchedBorderColor
		style.BorderColor = cfg.MouseDown.BorderColor
	case "move_mouse":
		style.BackgroundColor = cfg.MoveMouse.BackgroundColor
		style.TextColor = cfg.MoveMouse.TextColor
		style.MatchedTextColor = cfg.MoveMouse.MatchedTextColor
		style.MatchedBackgroundColor = cfg.MoveMouse.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.MoveMouse.MatchedBorderColor
		style.BorderColor = cfg.MoveMouse.BorderColor
	case "scroll":
		style.BackgroundColor = cfg.Scroll.BackgroundColor
		style.TextColor = cfg.Scroll.TextColor
		style.MatchedTextColor = cfg.Scroll.MatchedTextColor
		style.MatchedBackgroundColor = cfg.Scroll.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.Scroll.MatchedBorderColor
		style.BorderColor = cfg.Scroll.BorderColor
	case "context_menu":
		style.BackgroundColor = cfg.ContextMenu.BackgroundColor
		style.TextColor = cfg.ContextMenu.TextColor
		style.MatchedTextColor = cfg.ContextMenu.MatchedTextColor
		style.MatchedBackgroundColor = cfg.ContextMenu.MatchedBackgroundColor
		style.MatchedBorderColor = cfg.ContextMenu.MatchedBorderColor
		style.BorderColor = cfg.ContextMenu.BorderColor
	}

	return style
}
