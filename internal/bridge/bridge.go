package bridge

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Cocoa -framework Carbon -framework CoreGraphics
#include "accessibility.h"
#include "overlay.h"
#include "hotkeys.h"
#include "eventtap.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// This file ensures the bridge package is properly initialized
// and the Objective-C files are compiled with CGo
//
// The .m files are compiled separately via CGo's automatic source file detection

// SetApplicationAttribute toggles an accessibility attribute on the application with the provided PID.
func SetApplicationAttribute(pid int, attribute string, value bool) bool {
	cAttr := C.CString(attribute)
	defer C.free(unsafe.Pointer(cAttr))

	var cValue C.int
	if value {
		cValue = 1
	}

	result := C.setApplicationAttribute(C.int(pid), cAttr, cValue)
	return result == 1
}

// HasClickAction checks if an element has the AXPress action available.
func HasClickAction(element unsafe.Pointer) bool {
	if element == nil {
		return false
	}
	result := C.hasClickAction(element)
	return result == 1
}
