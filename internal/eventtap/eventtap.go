package eventtap

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/eventtap.h"
#include <stdlib.h>

extern void eventTapCallbackBridge(char* key, void* userData);
*/
import "C"

import (
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

// Callback is called when a key is pressed
type Callback func(key string)

// EventTap manages keyboard event interception
type EventTap struct {
	handle   C.EventTap
	callback Callback
	logger   *zap.Logger
}

// NewEventTap creates a new event tap
func NewEventTap(callback Callback, logger *zap.Logger) *EventTap {
	et := &EventTap{
		callback: callback,
		logger:   logger,
	}

	et.handle = C.createEventTap(C.EventTapCallback(C.eventTapCallbackBridge), nil)
	if et.handle == nil {
		logger.Error("Failed to create event tap - check Accessibility permissions")
		return nil
	}

	// Store in global variable for callbacks with mutex protection
	globalEventTapMu.Lock()
	globalEventTap = et
	globalEventTapMu.Unlock()

	return et
}

// Enable enables the event tap
func (et *EventTap) Enable() {
	et.logger.Debug("Enabling event tap")
	if et.handle != nil {
		C.enableEventTap(et.handle)
		et.logger.Debug("Event tap enabled")
	} else {
		et.logger.Warn("Cannot enable nil event tap")
	}
}

// SetHotkeys configures which hotkey combinations should pass through to the system
func (et *EventTap) SetHotkeys(hotkeys []string) {
	et.logger.Debug("Setting event tap hotkeys", zap.Int("count", len(hotkeys)))

	if et.handle == nil {
		et.logger.Warn("Cannot set hotkeys on nil event tap")
		return
	}

	// Convert Go string slice to C array
	cHotkeys := make([]*C.char, len(hotkeys))
	for i, hotkey := range hotkeys {
		if hotkey != "" {
			cHotkeys[i] = C.CString(hotkey)
			defer C.free(unsafe.Pointer(cHotkeys[i]))
			et.logger.Debug("Adding hotkey", zap.String("hotkey", hotkey))
		} else {
			cHotkeys[i] = nil
		}
	}

	// Pass pointer to first element and length
	var cHotkeysPtr **C.char
	if len(cHotkeys) > 0 {
		cHotkeysPtr = &cHotkeys[0]
	}

	C.setEventTapHotkeys(et.handle, cHotkeysPtr, C.int(len(cHotkeys)))
	et.logger.Debug("Event tap hotkeys set")
}

// Disable disables the event tap
func (et *EventTap) Disable() {
	et.logger.Debug("Disabling event tap")
	if et.handle != nil {
		C.disableEventTap(et.handle)
		et.logger.Debug("Event tap disabled")
	} else {
		et.logger.Warn("Cannot disable nil event tap")
	}
}

// Destroy destroys the event tap
func (et *EventTap) Destroy() {
	et.logger.Debug("Destroying event tap")
	if et.handle != nil {
		// Disable first to prevent any pending callbacks
		et.Disable()

		// Destroy the tap
		C.destroyEventTap(et.handle)
		et.handle = nil

		// Clear callback to prevent any lingering references
		et.callback = nil

		// Clear global reference if this is the global event tap
		globalEventTapMu.Lock()
		if globalEventTap == et {
			globalEventTap = nil
		}
		globalEventTapMu.Unlock()

		et.logger.Debug("Event tap destroyed")
	} else {
		et.logger.Warn("Cannot destroy nil event tap")
	}
}

// handleCallback handles a key press callback from C
func (et *EventTap) handleCallback(key string) {
	et.logger.Debug("Key pressed", zap.String("key", key))

	if et.callback != nil {
		et.logger.Debug("Calling event tap callback")
		et.callback(key)
	} else {
		et.logger.Debug("No callback registered for event tap")
	}
}

// Global event tap instance for C callbacks with thread safety
var (
	globalEventTap   *EventTap
	globalEventTapMu sync.RWMutex
)

//export eventTapCallbackBridge
func eventTapCallbackBridge(key *C.char, userData unsafe.Pointer) {
	globalEventTapMu.RLock()
	et := globalEventTap
	globalEventTapMu.RUnlock()

	if et != nil && key != nil {
		goKey := C.GoString(key)
		et.handleCallback(goKey)
	}
}
