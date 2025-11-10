package hints

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/y3owk1n/neru/internal/config"
)

// Overlay manages the hint overlay window
type Overlay struct {
	window C.OverlayWindow
	config config.HintsConfig
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
			matchedPrefixLength: C.int(len(hint.MatchedPrefix)),
		}
	}

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
	C.free(unsafe.Pointer(cFontFamily))
	C.free(unsafe.Pointer(cBgColor))
	C.free(unsafe.Pointer(cTextColor))
	C.free(unsafe.Pointer(cMatchedTextColor))
	C.free(unsafe.Pointer(cBorderColor))

	return nil
}

// GetWindow returns the underlying C overlay window
func (o *Overlay) GetWindow() C.OverlayWindow {
	return o.window
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

// BuildStyleForAction returns StyleMode based on action name using the provided config
func BuildStyleForAction(cfg config.HintsConfig, action string) StyleMode {
	style := StyleMode{
		FontSize:     cfg.FontSize,
		FontFamily:   cfg.FontFamily,
		BorderRadius: cfg.BorderRadius,
		Padding:      cfg.Padding,
		BorderWidth:  cfg.BorderWidth,
		Opacity:      cfg.Opacity,
	}

	switch action {
	case "left_click":
		style.BackgroundColor = cfg.LeftClickHints.BackgroundColor
		style.TextColor = cfg.LeftClickHints.TextColor
		style.MatchedTextColor = cfg.LeftClickHints.MatchedTextColor
		style.BorderColor = cfg.LeftClickHints.BorderColor
	case "right_click":
		style.BackgroundColor = cfg.RightClickHints.BackgroundColor
		style.TextColor = cfg.RightClickHints.TextColor
		style.MatchedTextColor = cfg.RightClickHints.MatchedTextColor
		style.BorderColor = cfg.RightClickHints.BorderColor
	case "double_click":
		style.BackgroundColor = cfg.DoubleClickHints.BackgroundColor
		style.TextColor = cfg.DoubleClickHints.TextColor
		style.MatchedTextColor = cfg.DoubleClickHints.MatchedTextColor
		style.BorderColor = cfg.DoubleClickHints.BorderColor
	case "triple_click":
		style.BackgroundColor = cfg.TripleClickHints.BackgroundColor
		style.TextColor = cfg.TripleClickHints.TextColor
		style.MatchedTextColor = cfg.TripleClickHints.MatchedTextColor
		style.BorderColor = cfg.TripleClickHints.BorderColor
	case "mouse_up":
		style.BackgroundColor = cfg.MouseUpHints.BackgroundColor
		style.TextColor = cfg.MouseUpHints.TextColor
		style.MatchedTextColor = cfg.MouseUpHints.MatchedTextColor
		style.BorderColor = cfg.MouseUpHints.BorderColor
	case "mouse_down":
		style.BackgroundColor = cfg.MouseDownHints.BackgroundColor
		style.TextColor = cfg.MouseDownHints.TextColor
		style.MatchedTextColor = cfg.MouseDownHints.MatchedTextColor
		style.BorderColor = cfg.MouseDownHints.BorderColor
	case "move_mouse":
		style.BackgroundColor = cfg.MoveMouseHints.BackgroundColor
		style.TextColor = cfg.MoveMouseHints.TextColor
		style.MatchedTextColor = cfg.MoveMouseHints.MatchedTextColor
		style.BorderColor = cfg.MoveMouseHints.BorderColor
	case "middle_click":
		style.BackgroundColor = cfg.MiddleClickHints.BackgroundColor
		style.TextColor = cfg.MiddleClickHints.TextColor
		style.MatchedTextColor = cfg.MiddleClickHints.MatchedTextColor
		style.BorderColor = cfg.MiddleClickHints.BorderColor
	case "scroll":
		style.BackgroundColor = cfg.ScrollHints.BackgroundColor
		style.TextColor = cfg.ScrollHints.TextColor
		style.MatchedTextColor = cfg.ScrollHints.MatchedTextColor
		style.BorderColor = cfg.ScrollHints.BorderColor
	case "context_menu":
		style.BackgroundColor = cfg.ContextMenuHints.BackgroundColor
		style.TextColor = cfg.ContextMenuHints.TextColor
		style.MatchedTextColor = cfg.ContextMenuHints.MatchedTextColor
		style.BorderColor = cfg.ContextMenuHints.BorderColor
	}

	return style
}

// BuildContextMenuLabel returns the context menu label block used for hints action menu
func BuildContextMenuLabel() string {
	items := ContextMenuItems()
	var formatted []string
	for _, it := range items {
		formatted = append(formatted, fmt.Sprintf("[%s]%s", it.Key, it.Label))
	}
	return strings.Join(formatted, "\n")
}

// Destroy destroys the overlay
func (o *Overlay) Destroy() {
	if o.window != nil {
		C.destroyOverlayWindow(o.window)
		o.window = nil
	}
}
