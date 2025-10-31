package eventtap

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/eventtap.h"
#include <stdlib.h>

extern void eventTapCallbackBridge(char* key, void* userData);
*/
import "C"
import (
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

	// Store in global map for callbacks
	globalEventTap = et

	return et
}

// Enable enables the event tap
func (et *EventTap) Enable() {
	if et.handle != nil {
		C.enableEventTap(et.handle)
		et.logger.Debug("Event tap enabled")
	}
}

// Disable disables the event tap
func (et *EventTap) Disable() {
	if et.handle != nil {
		C.disableEventTap(et.handle)
		et.logger.Debug("Event tap disabled")
	}
}

// Destroy destroys the event tap
func (et *EventTap) Destroy() {
	if et.handle != nil {
		C.destroyEventTap(et.handle)
		et.handle = nil
	}
}

// handleCallback handles a key press callback from C
func (et *EventTap) handleCallback(key string) {
	et.logger.Debug("Key pressed", zap.String("key", key))
	
	if et.callback != nil {
		et.callback(key)
	}
}

// Global event tap instance for C callbacks
var globalEventTap *EventTap

//export eventTapCallbackBridge
func eventTapCallbackBridge(key *C.char, userData unsafe.Pointer) {
	if globalEventTap != nil && key != nil {
		goKey := C.GoString(key)
		globalEventTap.handleCallback(goKey)
	}
}
