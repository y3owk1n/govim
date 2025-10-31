package bridge

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Cocoa -framework Carbon -framework CoreGraphics
#include "accessibility.h"
#include "overlay.h"
#include "hotkeys.h"
#include "eventtap.h"
*/
import "C"

// This file ensures the bridge package is properly initialized
// and the Objective-C files are compiled with CGo
//
// The .m files are compiled separately via CGo's automatic source file detection
