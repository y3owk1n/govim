// Package hotkeys provides hotkey management functionality.
package hotkeys

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/hotkeys.h"
#include <stdlib.h>

extern void hotkeyCallbackBridge(int hotkeyId, void* userData);
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

// Package hotkeys provides functionality for registering and handling global hotkeys
// in the Neru application using the Carbon Event Manager API.

// HotkeyID represents a unique hotkey identifier.
type HotkeyID int

// Callback is called when a hotkey is pressed.
type Callback func()

// Manager manages global hotkeys.
type Manager struct {
	callbacks map[HotkeyID]Callback
	mu        sync.RWMutex
	logger    *zap.Logger
	nextID    HotkeyID
}

// NewManager creates a new hotkey manager.
func NewManager(logger *zap.Logger) *Manager {
	manager := &Manager{
		callbacks: make(map[HotkeyID]Callback),
		logger:    logger,
		nextID:    1,
	}

	return manager
}

// Register registers a global hotkey.
func (m *Manager) Register(keyString string, callback Callback) (HotkeyID, error) {
	m.logger.Debug("Registering hotkey", zap.String("key", keyString))

	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse key string
	var keyCode, modifiers C.int
	cKeyString := C.CString(keyString)
	defer C.free(unsafe.Pointer(cKeyString))

	result := C.parseKeyString(cKeyString, &keyCode, &modifiers)
	if result == 0 {
		m.logger.Error("Failed to parse key string", zap.String("key", keyString))
		return 0, fmt.Errorf("failed to parse key string: %s", keyString)
	}

	// Generate hotkey ID
	hotkeyID := m.nextID
	m.nextID++

	m.logger.Debug("Parsed key string",
		zap.String("key", keyString),
		zap.Int("key_code", int(keyCode)),
		zap.Int("modifiers", int(modifiers)),
		zap.Int("id", int(hotkeyID)))

	// Register hotkey
	success := C.registerHotkey(keyCode, modifiers, C.int(hotkeyID),
		C.HotkeyCallback(C.hotkeyCallbackBridge), nil)

	if success == 0 {
		m.logger.Error("Failed to register hotkey", zap.String("key", keyString))
		return 0, fmt.Errorf("failed to register hotkey: %s", keyString)
	}

	// Store callback
	m.callbacks[hotkeyID] = callback

	m.logger.Info("Registered hotkey",
		zap.String("key", keyString),
		zap.Int("id", int(hotkeyID)))

	return hotkeyID, nil
}

// Unregister unregisters a hotkey.
func (m *Manager) Unregister(hotkeyID HotkeyID) {
	m.logger.Debug("Unregistering hotkey", zap.Int("id", int(hotkeyID)))

	m.mu.Lock()
	defer m.mu.Unlock()

	C.unregisterHotkey(C.int(hotkeyID))
	delete(m.callbacks, hotkeyID)

	m.logger.Info("Unregistered hotkey", zap.Int("id", int(hotkeyID)))
}

// UnregisterAll unregisters all hotkeys.
func (m *Manager) UnregisterAll() {
	m.logger.Debug("Unregistering all hotkeys")

	m.mu.Lock()
	defer m.mu.Unlock()

	C.unregisterAllHotkeys()
	m.callbacks = make(map[HotkeyID]Callback)

	m.logger.Info("Unregistered all hotkeys")
}

// handleCallback handles a hotkey callback from C.
func (m *Manager) handleCallback(hotkeyID HotkeyID) {
	m.logger.Debug("Handling hotkey callback", zap.Int("id", int(hotkeyID)))

	m.mu.RLock()
	callback, ok := m.callbacks[hotkeyID]
	m.mu.RUnlock()

	if ok && callback != nil {
		m.logger.Debug("Hotkey pressed", zap.Int("id", int(hotkeyID)))
		callback()
	} else {
		m.logger.Debug("No callback registered for hotkey", zap.Int("id", int(hotkeyID)))
	}
}

// Global manager instance for C callbacks.
var globalManager *Manager

// SetGlobalManager sets the global manager for C callbacks.
func SetGlobalManager(manager *Manager) {
	if manager != nil {
		manager.logger.Debug("Setting global hotkey manager")
	} else {
		// This would be unusual but let's log it
		logger.Get().Info("Setting global hotkey manager to nil")
	}
	globalManager = manager
}

//export hotkeyCallbackBridge
func hotkeyCallbackBridge(hotkeyID C.int, _ unsafe.Pointer) {
	if globalManager != nil {
		globalManager.logger.Debug("Hotkey callback bridge called", zap.Int("id", int(hotkeyID)))
		go globalManager.handleCallback(HotkeyID(hotkeyID))
	}
}
