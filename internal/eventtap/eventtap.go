// Package eventtap provides low-level event tapping functionality for key capture.
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

// Package eventtap provides functionality for capturing and handling global keyboard events
// on macOS using the Carbon Event Manager API.

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
	eventTap := &EventTap{
		callback: callback,
		logger:   logger,
	}

	eventTap.handle = C.createEventTap(C.EventTapCallback(C.eventTapCallbackBridge), nil)
	if eventTap.handle == nil {
		logger.Error("Failed to create event tap - check Accessibility permissions")
		return nil
	}

	// Store in global variable for callbacks with mutex protection
	globalEventTapMu.Lock()
	globalEventTap = eventTap
	globalEventTapMu.Unlock()

	return eventTap
}

// Enable enables the event tap
func (eventTap *EventTap) Enable() {
	eventTap.logger.Debug("Enabling event tap")
	if eventTap.handle != nil {
		C.enableEventTap(eventTap.handle)
		eventTap.logger.Debug("Event tap enabled")
	} else {
		eventTap.logger.Warn("Cannot enable nil event tap")
	}
}

// SetHotkeys configures which hotkey combinations should pass through to the system
func (eventTap *EventTap) SetHotkeys(hotkeys []string) {
	eventTap.logger.Debug("Setting event tap hotkeys", zap.Int("count", len(hotkeys)))

	if eventTap.handle == nil {
		eventTap.logger.Warn("Cannot set hotkeys on nil event tap")
		return
	}

	// Convert Go string slice to C array
	cHotkeys := make([]*C.char, len(hotkeys))
	for index, hotkey := range hotkeys {
		if hotkey != "" {
			cHotkeys[index] = C.CString(hotkey)
			defer C.free(unsafe.Pointer(cHotkeys[index]))
			eventTap.logger.Debug("Adding hotkey", zap.String("hotkey", hotkey))
		} else {
			cHotkeys[index] = nil
		}
	}

	// Pass pointer to first element and length
	cHotkeysPtr := (**C.char)(nil)
	if len(cHotkeys) > 0 {
		cHotkeysPtr = &cHotkeys[0]
	}

	C.setEventTapHotkeys(eventTap.handle, cHotkeysPtr, C.int(len(cHotkeys)))
	eventTap.logger.Debug("Event tap hotkeys set")
}

// Disable disables the event tap
func (eventTap *EventTap) Disable() {
	eventTap.logger.Debug("Disabling event tap")
	if eventTap.handle != nil {
		C.disableEventTap(eventTap.handle)
		eventTap.logger.Debug("Event tap disabled")
	} else {
		eventTap.logger.Warn("Cannot disable nil event tap")
	}
}

// Destroy destroys the event tap
func (eventTap *EventTap) Destroy() {
	eventTap.logger.Debug("Destroying event tap")
	if eventTap.handle != nil {
		// Disable first to prevent any pending callbacks
		eventTap.Disable()

		// Destroy the tap
		C.destroyEventTap(eventTap.handle)
		eventTap.handle = nil

		// Clear callback to prevent any lingering references
		eventTap.callback = nil

		// Clear global reference if this is the global event tap
		globalEventTapMu.Lock()
		if globalEventTap == eventTap {
			globalEventTap = nil
		}
		globalEventTapMu.Unlock()

		eventTap.logger.Debug("Event tap destroyed")
	} else {
		eventTap.logger.Warn("Cannot destroy nil event tap")
	}
}

// handleCallback handles a key press callback from C
func (eventTap *EventTap) handleCallback(key string) {
	eventTap.logger.Debug("Key pressed", zap.String("key", key))

	if eventTap.callback != nil {
		eventTap.logger.Debug("Calling event tap callback")
		eventTap.callback(key)
	} else {
		eventTap.logger.Debug("No callback registered for event tap")
	}
}

// Global event tap instance for C callbacks with thread safety
var (
	globalEventTap   *EventTap
	globalEventTapMu sync.RWMutex
)

//export eventTapCallbackBridge
func eventTapCallbackBridge(key *C.char, _ unsafe.Pointer) {
	globalEventTapMu.RLock()
	eventTap := globalEventTap
	globalEventTapMu.RUnlock()

	if eventTap != nil && key != nil {
		goKey := C.GoString(key)
		eventTap.handleCallback(goKey)
	}
}
