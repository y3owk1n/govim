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
	"image"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

var (
	gridCallbackID        uint64
	gridCallbackMap       = make(map[uint64]chan struct{})
	gridCallbackLock      sync.Mutex
	gridCellSlicePool     sync.Pool
	gridLabelSlicePool    sync.Pool
	subgridCellSlicePool  sync.Pool
	subgridLabelSlicePool sync.Pool
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
	gridCellSlicePool = sync.Pool{New: func() any { s := make([]C.GridCell, 0); return &s }}
	gridLabelSlicePool = sync.Pool{New: func() any { s := make([]*C.char, 0); return &s }}
	subgridCellSlicePool = sync.Pool{New: func() any { s := make([]C.GridCell, 0); return &s }}
	subgridLabelSlicePool = sync.Pool{New: func() any { s := make([]*C.char, 0); return &s }}
	chars := cfg.Characters
	if strings.TrimSpace(chars) == "" {
		chars = cfg.Characters
	}
	go Prewarm(chars, []image.Rectangle{
		image.Rect(0, 0, 1280, 800),
		image.Rect(0, 0, 1366, 768),
		image.Rect(0, 0, 1440, 900),
		image.Rect(0, 0, 1920, 1080),
		image.Rect(0, 0, 2560, 1440),
		image.Rect(0, 0, 3440, 1440),
		image.Rect(0, 0, 3840, 2160),
	})
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

	if o.logger != nil {
		o.logger.Debug("Grid overlay resize started", zap.Uint64("callback_id", id))
	}

	// Pass ID as context (safe - no Go pointers)
	// Note: uintptr conversion must happen in same expression to satisfy go vet
	C.resizeOverlayToActiveScreenWithCallback(
		o.window,
		(C.ResizeCompletionCallback)(unsafe.Pointer(C.gridResizeCompletionCallback)),
		*(*unsafe.Pointer)(unsafe.Pointer(&id)),
	)

	// Don't wait for callback - continue immediately for better UX
	// The resize operation is typically fast and visually complete before callback
	// Start a goroutine to handle cleanup when callback eventually arrives
	go func() {
		if o.logger != nil {
			o.logger.Debug("Grid overlay resize background cleanup started", zap.Uint64("callback_id", id))
		}

		select {
		case <-done:
			// Callback received, normal cleanup already handled in callback
			if o.logger != nil {
				o.logger.Debug("Grid overlay resize callback received", zap.Uint64("callback_id", id))
			}
		case <-time.After(2 * time.Second):
			// Long timeout for cleanup only - callback likely failed
			gridCallbackLock.Lock()
			delete(gridCallbackMap, id)
			gridCallbackLock.Unlock()

			if o.logger != nil {
				o.logger.Debug("Grid overlay resize cleanup timeout - removed callback from map",
					zap.Uint64("callback_id", id))
			}
		}
	}()
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

	start := time.Now()
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)

	o.drawGridCells(cells, currentInput, style)

	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	o.logger.Info("Grid draw perf",
		zap.Int("cell_count", len(cells)),
		zap.Duration("duration", time.Since(start)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))

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

	cellsPtr := subgridCellSlicePool.Get().(*[]C.GridCell)
	if cap(*cellsPtr) < count {
		s := make([]C.GridCell, count)
		cellsPtr = &s
	} else {
		*cellsPtr = (*cellsPtr)[:count]
	}
	cells := *cellsPtr
	labelsPtr := subgridLabelSlicePool.Get().(*[]*C.char)
	if cap(*labelsPtr) < count {
		s := make([]*C.char, count)
		labelsPtr = &s
	} else {
		*labelsPtr = (*labelsPtr)[:count]
	}
	labels := *labelsPtr

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

		var cell C.GridCell
		cell.label = labels[i]
		cell.bounds.origin.x = C.double(left)
		cell.bounds.origin.y = C.double(top)
		cell.bounds.size.width = C.double(right - left)
		cell.bounds.size.height = C.double(bottom - top)
		cell.isMatched = C.int(0)
		cell.isSubgrid = C.int(1)           // Mark as subgrid cell
		cell.matchedPrefixLength = C.int(0) // Subgrid cells don't have matched prefixes

		cells[i] = cell
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
	*cellsPtr = (*cellsPtr)[:0]
	*labelsPtr = (*labelsPtr)[:0]
	subgridCellSlicePool.Put(cellsPtr)
	subgridLabelSlicePool.Put(labelsPtr)
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

	cGridCellsPtr := gridCellSlicePool.Get().(*[]C.GridCell)
	if cap(*cGridCellsPtr) < len(cellsGo) {
		s := make([]C.GridCell, len(cellsGo))
		cGridCellsPtr = &s
	} else {
		*cGridCellsPtr = (*cGridCellsPtr)[:len(cellsGo)]
	}
	cGridCells := *cGridCellsPtr
	cLabelsPtr := gridLabelSlicePool.Get().(*[]*C.char)
	if cap(*cLabelsPtr) < len(cellsGo) {
		s := make([]*C.char, len(cellsGo))
		cLabelsPtr = &s
	} else {
		*cLabelsPtr = (*cLabelsPtr)[:len(cellsGo)]
	}
	cLabels := *cLabelsPtr

	matchedCount := 0
	for i, cell := range cellsGo {
		cLabels[i] = C.CString(cell.Coordinate)

		isMatched := 0
		matchedPrefixLength := 0
		if currentInput != "" && strings.HasPrefix(cell.Coordinate, currentInput) {
			isMatched = 1
			matchedCount++
			matchedPrefixLength = len(currentInput)
		}

		var cGridCell C.GridCell
		cGridCell.label = cLabels[i]
		cGridCell.bounds.origin.x = C.double(cell.Bounds.Min.X)
		cGridCell.bounds.origin.y = C.double(cell.Bounds.Min.Y)
		cGridCell.bounds.size.width = C.double(cell.Bounds.Dx())
		cGridCell.bounds.size.height = C.double(cell.Bounds.Dy())
		cGridCell.isMatched = C.int(isMatched)
		cGridCell.isSubgrid = C.int(0) // Mark as regular grid cell
		cGridCell.matchedPrefixLength = C.int(matchedPrefixLength)

		cGridCells[i] = cGridCell
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
	*cGridCellsPtr = (*cGridCellsPtr)[:0]
	*cLabelsPtr = (*cLabelsPtr)[:0]
	gridCellSlicePool.Put(cGridCellsPtr)
	gridLabelSlicePool.Put(cLabelsPtr)
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

// BuildStyle returns GridStyle based on action name using the provided config
func BuildStyle(cfg config.GridConfig) GridStyle {
	style := GridStyle{
		FontSize:               cfg.FontSize,
		FontFamily:             cfg.FontFamily,
		Opacity:                cfg.Opacity,
		BorderWidth:            cfg.BorderWidth,
		BackgroundColor:        cfg.BackgroundColor,
		TextColor:              cfg.TextColor,
		MatchedTextColor:       cfg.MatchedTextColor,
		MatchedBackgroundColor: cfg.MatchedBackgroundColor,
		MatchedBorderColor:     cfg.MatchedBorderColor,
		BorderColor:            cfg.BorderColor,
	}

	return style
}
