package scroll

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>

// Callback function that Go can reference
extern void resizeScrollCompletionCallback(void* context);
*/
import "C"

import (
	"fmt"
	"image"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
)

var (
	scrollCallbackID   uint64
	scrollCallbackMap  = make(map[uint64]chan struct{})
	scrollCallbackLock sync.Mutex
)

//export resizeScrollCompletionCallback
func resizeScrollCompletionCallback(context unsafe.Pointer) {
	// Convert context to callback ID
	id := uint64(uintptr(context))

	scrollCallbackLock.Lock()
	if done, ok := scrollCallbackMap[id]; ok {
		close(done)
		delete(scrollCallbackMap, id)
	}
	scrollCallbackLock.Unlock()
}

type Overlay struct {
	window C.OverlayWindow
	config config.ScrollConfig
	logger *zap.Logger
}

// NewOverlay creates a new overlay
func NewOverlay(cfg config.ScrollConfig, logger *zap.Logger) (*Overlay, error) {
	window := C.createOverlayWindow()
	if window == nil {
		return nil, fmt.Errorf("failed to create overlay window")
	}
	return &Overlay{
		window: window,
		config: cfg,
		logger: logger,
	}, nil
}

// Show shows the overlay
func (o *Overlay) Show() {
	o.logger.Debug("Showing scroll overlay")
	C.showOverlayWindow(o.window)
	o.logger.Debug("Scroll overlay shown successfully")
}

// Hide hides the overlay
func (o *Overlay) Hide() {
	o.logger.Debug("Hiding scroll overlay")
	C.hideOverlayWindow(o.window)
	o.logger.Debug("Scroll overlay hidden successfully")
}

// Clear clears all scroll from the overlay
func (o *Overlay) Clear() {
	o.logger.Debug("Clearing scroll overlay")
	C.clearOverlay(o.window)
	o.logger.Debug("Scroll overlay cleared successfully")
}

// ResizeToActiveScreen resizes the overlay window to the screen containing the mouse cursor
func (o *Overlay) ResizeToActiveScreen() {
	C.resizeOverlayToActiveScreen(o.window)
}

// ResizeToActiveScreenSync resizes the overlay window synchronously with callback notification
func (o *Overlay) ResizeToActiveScreenSync() {
	done := make(chan struct{})

	// Generate unique ID for this callback
	id := atomic.AddUint64(&scrollCallbackID, 1)

	// Store channel in map
	scrollCallbackLock.Lock()
	scrollCallbackMap[id] = done
	scrollCallbackLock.Unlock()

	if o.logger != nil {
		o.logger.Debug("Scroll overlay resize started", zap.Uint64("callback_id", id))
	}

	// Pass ID as context (safe - no Go pointers)
	// Note: uintptr conversion must happen in same expression to satisfy go vet
	C.resizeOverlayToActiveScreenWithCallback(
		o.window,
		(C.ResizeCompletionCallback)(unsafe.Pointer(C.resizeScrollCompletionCallback)),
		*(*unsafe.Pointer)(unsafe.Pointer(&id)),
	)

	// Don't wait for callback - continue immediately for better UX
	// The resize operation is typically fast and visually complete before callback
	// Start a goroutine to handle cleanup when callback eventually arrives
	go func() {
		if o.logger != nil {
			o.logger.Debug("Scroll overlay resize background cleanup started", zap.Uint64("callback_id", id))
		}

		select {
		case <-done:
			// Callback received, normal cleanup already handled in callback
			if o.logger != nil {
				o.logger.Debug("Scroll overlay resize callback received", zap.Uint64("callback_id", id))
			}
		case <-time.After(2 * time.Second):
			// Long timeout for cleanup only - callback likely failed
			scrollCallbackLock.Lock()
			delete(scrollCallbackMap, id)
			scrollCallbackLock.Unlock()

			if o.logger != nil {
				o.logger.Debug("Scroll overlay resize cleanup timeout - removed callback from map",
					zap.Uint64("callback_id", id))
			}
		}
	}()
}

// Destroy destroys the overlay
func (o *Overlay) Destroy() {
	if o.window != nil {
		C.destroyOverlayWindow(o.window)
		o.window = nil
	}
}

// DrawScrollHighlight draws a highlight around a scroll area
func (o *Overlay) DrawScrollHighlight() {
	o.logger.Debug("DrawScrollHighlight called")
	window := accessibility.GetFrontmostWindow()
	if window == nil {
		o.logger.Debug("No frontmost window")
		return
	}
	defer window.Release()

	// Get scroll bounds in absolute screen coordinates
	bounds := window.GetScrollBounds()

	// Get active screen bounds to normalize coordinates
	screenBounds := bridge.GetActiveScreenBounds()

	// Convert absolute bounds to window-local coordinates
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	localBounds := image.Rect(
		bounds.Min.X-screenBounds.Min.X,
		bounds.Min.Y-screenBounds.Min.Y,
		bounds.Max.X-screenBounds.Min.X,
		bounds.Max.Y-screenBounds.Min.Y,
	)

	o.logger.Debug("Drawing scroll highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))

	color := o.config.HighlightColor
	borderWidth := o.config.HighlightWidth

	renderBounds := C.CGRect{
		origin: C.CGPoint{
			x: C.double(localBounds.Min.X),
			y: C.double(localBounds.Min.Y),
		},
		size: C.CGSize{
			width:  C.double(localBounds.Dx()),
			height: C.double(localBounds.Dy()),
		},
	}

	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))

	C.drawScrollHighlight(o.window, renderBounds, cColor, C.int(borderWidth))
}
