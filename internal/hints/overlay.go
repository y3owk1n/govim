package hints

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/y3owk1n/govim/internal/config"
)

// Overlay manages the hint overlay window
type Overlay struct {
	window C.OverlayWindow
	config config.HintsConfig
}

// NewOverlay creates a new overlay
func NewOverlay(cfg config.HintsConfig) (*Overlay, error) {
	window := C.createOverlayWindow()
	if window == nil {
		return nil, fmt.Errorf("failed to create overlay window")
	}

	return &Overlay{
		window: window,
		config: cfg,
	}, nil
}

// Show shows the overlay
func (o *Overlay) Show() {
	C.showOverlayWindow(o.window)
}

// Hide hides the overlay
func (o *Overlay) Hide() {
	C.hideOverlayWindow(o.window)
}

// Clear clears all hints from the overlay
func (o *Overlay) Clear() {
	C.clearOverlay(o.window)
}

// DrawHints draws hints on the overlay
func (o *Overlay) DrawHints(hints []*Hint) error {
	if len(hints) == 0 {
		o.Clear()
		return nil
	}

	// Convert hints to C array - collect all C strings to free later
	cHints := make([]C.HintData, len(hints))
	cLabels := make([]*C.char, len(hints))
	
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
		}
	}

	// Create style
	cFontFamily := C.CString(o.config.FontFamily)
	cBgColor := C.CString(o.config.BackgroundColor)
	cTextColor := C.CString(o.config.TextColor)
	cBorderColor := C.CString(o.config.BorderColor)

	style := C.HintStyle{
		fontSize:        C.int(o.config.FontSize),
		fontFamily:      cFontFamily,
		backgroundColor: cBgColor,
		textColor:       cTextColor,
		borderColor:     cBorderColor,
		borderRadius:    C.int(o.config.BorderRadius),
		borderWidth:     C.int(o.config.BorderWidth),
		padding:         C.int(o.config.Padding),
		opacity:         C.double(o.config.Opacity),
	}

	// Draw hints
	C.drawHints(o.window, &cHints[0], C.int(len(cHints)), style)

	// Free all C strings
	for _, cLabel := range cLabels {
		C.free(unsafe.Pointer(cLabel))
	}
	C.free(unsafe.Pointer(cFontFamily))
	C.free(unsafe.Pointer(cBgColor))
	C.free(unsafe.Pointer(cTextColor))
	C.free(unsafe.Pointer(cBorderColor))

	return nil
}

// DrawScrollHighlight draws a highlight around a scroll area
func (o *Overlay) DrawScrollHighlight(x, y, width, height int, color string, borderWidth int) {
	bounds := C.CGRect{
		origin: C.CGPoint{
			x: C.double(x),
			y: C.double(y),
		},
		size: C.CGSize{
			width:  C.double(width),
			height: C.double(height),
		},
	}

	cColor := C.CString(color)
	defer C.free(unsafe.Pointer(cColor))

	C.drawScrollHighlight(o.window, bounds, cColor, C.int(borderWidth))
}

// SetLevel sets the window level
func (o *Overlay) SetLevel(level int) {
	C.setOverlayLevel(o.window, C.int(level))
}

// Destroy destroys the overlay
func (o *Overlay) Destroy() {
	if o.window != nil {
		C.destroyOverlayWindow(o.window)
		o.window = nil
	}
}
