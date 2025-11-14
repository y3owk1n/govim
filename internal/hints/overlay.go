package hints

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>

// Callback function that Go can reference
extern void resizeHintCompletionCallback(void* context);
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

var (
	hintCallbackID   uint64
	hintCallbackMap  = make(map[uint64]chan struct{})
	hintCallbackLock sync.Mutex
	hintDataPool     sync.Pool
	cLabelSlicePool  sync.Pool
)

//export resizeHintCompletionCallback
func resizeHintCompletionCallback(context unsafe.Pointer) {
	// Convert context to callback ID
	id := uint64(uintptr(context))

	hintCallbackLock.Lock()
	if done, ok := hintCallbackMap[id]; ok {
		close(done)
		delete(hintCallbackMap, id)
	}
	hintCallbackLock.Unlock()
}

// Overlay manages the hint overlay window
type Overlay struct {
	window C.OverlayWindow
	config config.HintsConfig
	logger *zap.Logger
}

type StyleMode struct {
	FontSize         int
	FontFamily       string
	BorderRadius     int
	Padding          int
	BorderWidth      int
	Opacity          float64
	BackgroundColor  string
	TextColor        string
	MatchedTextColor string
	BorderColor      string
}

// NewOverlay creates a new overlay
func NewOverlay(cfg config.HintsConfig, logger *zap.Logger) (*Overlay, error) {
	window := C.createOverlayWindow()
	if window == nil {
		return nil, fmt.Errorf("failed to create overlay window")
	}
	hintDataPool = sync.Pool{New: func() any { s := make([]C.HintData, 0); return &s }}
	cLabelSlicePool = sync.Pool{New: func() any { s := make([]*C.char, 0); return &s }}

	return &Overlay{
		window: window,
		config: cfg,
		logger: logger,
	}, nil
}

// Show shows the overlay
func (o *Overlay) Show() {
	o.logger.Debug("Showing hint overlay")
	C.showOverlayWindow(o.window)
	o.logger.Debug("Hint overlay shown successfully")
}

// Hide hides the overlay
func (o *Overlay) Hide() {
	o.logger.Debug("Hiding hint overlay")
	C.hideOverlayWindow(o.window)
	o.logger.Debug("Hint overlay hidden successfully")
}

// Clear clears all hints from the overlay
func (o *Overlay) Clear() {
	o.logger.Debug("Clearing hint overlay")
	C.clearOverlay(o.window)
	o.logger.Debug("Hint overlay cleared successfully")
}

// ResizeToActiveScreen resizes the overlay window to the screen containing the mouse cursor
func (o *Overlay) ResizeToActiveScreen() {
	C.resizeOverlayToActiveScreen(o.window)
}

// ResizeToActiveScreenSync resizes the overlay window synchronously with callback notification
func (o *Overlay) ResizeToActiveScreenSync() {
	done := make(chan struct{})

	// Generate unique ID for this callback
	id := atomic.AddUint64(&hintCallbackID, 1)

	// Store channel in map
	hintCallbackLock.Lock()
	hintCallbackMap[id] = done
	hintCallbackLock.Unlock()

	if o.logger != nil {
		o.logger.Debug("Hint overlay resize started", zap.Uint64("callback_id", id))
	}

	// Pass ID as context (safe - no Go pointers)
	// Note: uintptr conversion must happen in same expression to satisfy go vet
	C.resizeOverlayToActiveScreenWithCallback(
		o.window,
		(C.ResizeCompletionCallback)(unsafe.Pointer(C.resizeHintCompletionCallback)),
		*(*unsafe.Pointer)(unsafe.Pointer(&id)),
	)

	// Don't wait for callback - continue immediately for better UX
	// The resize operation is typically fast and visually complete before callback
	// Start a goroutine to handle cleanup when callback eventually arrives
	go func() {
		if o.logger != nil {
			o.logger.Debug("Hint overlay resize background cleanup started", zap.Uint64("callback_id", id))
		}

		select {
		case <-done:
			// Callback received, normal cleanup already handled in callback
			if o.logger != nil {
				o.logger.Debug("Hint overlay resize callback received", zap.Uint64("callback_id", id))
			}
		case <-time.After(2 * time.Second):
			// Long timeout for cleanup only - callback likely failed
			hintCallbackLock.Lock()
			delete(hintCallbackMap, id)
			hintCallbackLock.Unlock()

			if o.logger != nil {
				o.logger.Debug("Hint overlay resize cleanup timeout - removed callback from map",
					zap.Uint64("callback_id", id))
			}
		}
	}()
}

// DrawHintsWithoutArrow draws hints without arrows
func (o *Overlay) DrawHintsWithoutArrow(hints []*Hint, style StyleMode) error {
	return o.drawHintsInternal(hints, style, false)
}

// DrawHintsWithStyle draws hints on the overlay with custom style
func (o *Overlay) DrawHintsWithStyle(hints []*Hint, style StyleMode) error {
	return o.drawHintsInternal(hints, style, true)
}

// drawHintsInternal is the internal implementation for drawing hints
func (o *Overlay) drawHintsInternal(hints []*Hint, style StyleMode, showArrow bool) error {
	o.logger.Debug("Drawing hints internally",
		zap.Int("hint_count", len(hints)),
		zap.Bool("show_arrow", showArrow))

	if len(hints) == 0 {
		o.Clear()
		o.logger.Debug("No hints to draw, cleared overlay")
		return nil
	}

	start := time.Now()
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)
	cHintsPtr := hintDataPool.Get().(*[]C.HintData)
	if cap(*cHintsPtr) < len(hints) {
		s := make([]C.HintData, len(hints))
		cHintsPtr = &s
	} else {
		*cHintsPtr = (*cHintsPtr)[:len(hints)]
	}
	cHints := *cHintsPtr
	cLabelsPtr := cLabelSlicePool.Get().(*[]*C.char)
	if cap(*cLabelsPtr) < len(hints) {
		s := make([]*C.char, len(hints))
		cLabelsPtr = &s
	} else {
		*cLabelsPtr = (*cLabelsPtr)[:len(hints)]
	}
	cLabels := *cLabelsPtr

	matchedCount := 0
	for i, hint := range hints {
		cLabels[i] = C.CString(hint.Label)
		cHints[i] = C.HintData{
			label: cLabels[i],
			position: C.CGPoint{
				x: C.double(hint.Position.X),
				y: C.double(hint.Position.Y),
			},
			size: C.CGSize{
				width:  C.double(hint.Size.X),
				height: C.double(hint.Size.Y),
			},
			matchedPrefixLength: C.int(len(hint.MatchedPrefix)),
		}

		if len(hint.MatchedPrefix) > 0 {
			matchedCount++
		}
	}

	o.logger.Debug("Hint match statistics",
		zap.Int("total_hints", len(hints)),
		zap.Int("matched_hints", matchedCount))

	// Create style
	cFontFamily := C.CString(style.FontFamily)
	cBgColor := C.CString(style.BackgroundColor)
	cTextColor := C.CString(style.TextColor)
	cMatchedTextColor := C.CString(style.MatchedTextColor)
	cBorderColor := C.CString(style.BorderColor)

	arrowFlag := 0
	if showArrow {
		arrowFlag = 1
	}

	finalStyle := C.HintStyle{
		fontSize:         C.int(style.FontSize),
		fontFamily:       cFontFamily,
		backgroundColor:  cBgColor,
		textColor:        cTextColor,
		matchedTextColor: cMatchedTextColor,
		borderColor:      cBorderColor,
		borderRadius:     C.int(style.BorderRadius),
		borderWidth:      C.int(style.BorderWidth),
		padding:          C.int(style.Padding),
		opacity:          C.double(style.Opacity),
		showArrow:        C.int(arrowFlag),
	}

	// Draw hints
	C.drawHints(o.window, &cHints[0], C.int(len(cHints)), finalStyle)

	// Free all C strings
	for _, cLabel := range cLabels {
		C.free(unsafe.Pointer(cLabel))
	}
	*cHintsPtr = (*cHintsPtr)[:0]
	*cLabelsPtr = (*cLabelsPtr)[:0]
	hintDataPool.Put(cHintsPtr)
	cLabelSlicePool.Put(cLabelsPtr)
	C.free(unsafe.Pointer(cFontFamily))
	C.free(unsafe.Pointer(cBgColor))
	C.free(unsafe.Pointer(cTextColor))
	C.free(unsafe.Pointer(cMatchedTextColor))
	C.free(unsafe.Pointer(cBorderColor))

	o.logger.Debug("Hints drawn successfully")
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	o.logger.Info("Hints draw perf",
		zap.Int("hint_count", len(hints)),
		zap.Duration("duration", time.Since(start)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))
	return nil
}

// GetWindow returns the underlying C overlay window
func (o *Overlay) GetWindow() C.OverlayWindow {
	return o.window
}

// DrawTargetDot draws a small circular dot at the target position
func (o *Overlay) DrawTargetDot(x, y int, radius float64, color, borderColor string, borderWidth float64) error {
	center := C.CGPoint{
		x: C.double(x),
		y: C.double(y),
	}

	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))

	var cBorderColor *C.char
	if borderColor != "" {
		cBorderColor = C.CString(borderColor)
		defer C.free(unsafe.Pointer(cBorderColor))
	}

	C.drawTargetDot(o.window, center, C.double(radius), cColor, cBorderColor, C.double(borderWidth))

	return nil
}

// BuildStyle returns StyleMode based on action name using the provided config
func BuildStyle(cfg config.HintsConfig) StyleMode {
	style := StyleMode{
		FontSize:         cfg.FontSize,
		FontFamily:       cfg.FontFamily,
		BorderRadius:     cfg.BorderRadius,
		Padding:          cfg.Padding,
		BorderWidth:      cfg.BorderWidth,
		Opacity:          cfg.Opacity,
		BackgroundColor:  cfg.BackgroundColor,
		TextColor:        cfg.TextColor,
		MatchedTextColor: cfg.MatchedTextColor,
		BorderColor:      cfg.BorderColor,
	}

	return style
}

// Destroy destroys the overlay
func (o *Overlay) Destroy() {
	if o.window != nil {
		C.destroyOverlayWindow(o.window)
		o.window = nil
	}
}
