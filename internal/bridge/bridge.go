package bridge

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Cocoa -framework Carbon -framework CoreGraphics
#include "accessibility.h"
#include "overlay.h"
#include "hotkeys.h"
#include "eventtap.h"
#include "appwatcher.h"
#include <stdlib.h>

CGRect getActiveScreenBounds();
*/
import "C"

import (
	"image"
	"sync"
	"unsafe"
)

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

var (
	appWatcher     AppWatcher
	appWatcherOnce sync.Once
)

// AppWatcher interface defines the methods for application watching
type AppWatcher interface {
	HandleLaunch(appName, bundleID string)
	HandleTerminate(appName, bundleID string)
	HandleActivate(appName, bundleID string)
	HandleDeactivate(appName, bundleID string)
}

// SetAppWatcher sets the application watcher implementation
func SetAppWatcher(w AppWatcher) {
	appWatcher = w
}

// StartAppWatcher starts watching for application events
func StartAppWatcher() {
	C.startAppWatcher()
}

// StopAppWatcher stops watching for application events
func StopAppWatcher() {
	C.stopAppWatcher()
}

//export handleAppLaunch
func handleAppLaunch(cAppName *C.char, cBundleID *C.char) {
	if appWatcher != nil {
		appName := C.GoString(cAppName)
		bundleID := C.GoString(cBundleID)
		appWatcher.HandleLaunch(appName, bundleID)
	}
}

//export handleAppTerminate
func handleAppTerminate(cAppName *C.char, cBundleID *C.char) {
	if appWatcher != nil {
		appName := C.GoString(cAppName)
		bundleID := C.GoString(cBundleID)
		appWatcher.HandleTerminate(appName, bundleID)
	}
}

//export handleAppActivate
func handleAppActivate(cAppName *C.char, cBundleID *C.char) {
	if appWatcher != nil {
		appName := C.GoString(cAppName)
		bundleID := C.GoString(cBundleID)
		appWatcher.HandleActivate(appName, bundleID)
	}
}

//export handleAppDeactivate
func handleAppDeactivate(cAppName *C.char, cBundleID *C.char) {
	if appWatcher != nil {
		appName := C.GoString(cAppName)
		bundleID := C.GoString(cBundleID)
		appWatcher.HandleDeactivate(appName, bundleID)
	}
}

// GetMainScreenBounds returns the bounds of the main screen
func GetMainScreenBounds() image.Rectangle {
	rect := C.getMainScreenBounds()
	return image.Rect(
		int(rect.origin.x),
		int(rect.origin.y),
		int(rect.origin.x+rect.size.width),
		int(rect.origin.y+rect.size.height),
	)
}

// GetActiveScreenBounds returns the bounds of the screen containing the mouse cursor
func GetActiveScreenBounds() image.Rectangle {
	rect := C.getActiveScreenBounds()
	return image.Rect(
		int(rect.origin.x),
		int(rect.origin.y),
		int(rect.origin.x+rect.size.width),
		int(rect.origin.y+rect.size.height),
	)
}
