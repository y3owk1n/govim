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

	"github.com/y3owk1n/neru/internal/appwatcher"
	"github.com/y3owk1n/neru/internal/bridge"
	"go.uber.org/zap"
)

// HotkeyID represents a unique hotkey identifier
type HotkeyID int

// Callback is called when a hotkey is pressed
type Callback func()

// Manager manages global hotkeys
type Manager struct {
	watcher                  *appwatcher.Watcher
	excludedApps             []string
	callbacks                map[HotkeyID]Callback
	mu                       sync.RWMutex
	logger                   *zap.Logger
	nextID                   HotkeyID
	handleLaunchCallback     func(appName, bundleID string)
	handleTerminateCallback  func(appName, bundleID string)
	handleActivateCallback   func(appName, bundleID string)
	handleDeactivateCallback func(appName, bundleID string)
}

// NewManager creates a new hotkey manager
func NewManager(logger *zap.Logger, excludedApps []string) *Manager {
	manager := &Manager{
		watcher:                  appwatcher.New(),
		excludedApps:             excludedApps,
		callbacks:                make(map[HotkeyID]Callback),
		logger:                   logger,
		nextID:                   1,
		handleLaunchCallback:     func(appName, bundleID string) {},
		handleTerminateCallback:  func(appName, bundleID string) {},
		handleActivateCallback:   func(appName, bundleID string) {},
		handleDeactivateCallback: func(appName, bundleID string) {},
	}

	bridge.SetAppWatcher(bridge.AppWatcher(manager))

	return manager
}

// Start initialize
func (m *Manager) Start() {
	bridge.StartAppWatcher()
}

func (m *Manager) Stop() {
	bridge.StopAppWatcher()
}

func (m *Manager) SetLaunchCallback(callback func(appName, bundleID string)) {
	m.handleLaunchCallback = callback
}

func (m *Manager) SetTerminateCallback(callback func(appName, bundleID string)) {
	m.handleTerminateCallback = callback
}

func (m *Manager) SetActivateCallback(callback func(appName, bundleID string)) {
	m.handleActivateCallback = callback
}

func (m *Manager) SetDeactivateCallback(callback func(appName, bundleID string)) {
	m.handleDeactivateCallback = callback
}

func (m *Manager) HandleLaunch(appName, bundleID string) {
	m.logger.Debug("[Hotkey] App launched", zap.String("bundle_id", bundleID))

	if m.handleLaunchCallback != nil {
		m.handleLaunchCallback(appName, bundleID)
	}
}

func (m *Manager) HandleTerminate(appName, bundleID string) {
	m.logger.Debug("[Hotkey] App terminated", zap.String("bundle_id", bundleID))

	if m.handleTerminateCallback != nil {
		m.handleTerminateCallback(appName, bundleID)
	}
}

func (m *Manager) HandleActivate(appName, bundleID string) {
	m.logger.Debug("[Hotkey] App activated", zap.String("bundle_id", bundleID))

	if m.handleActivateCallback != nil {
		m.handleActivateCallback(appName, bundleID)
	}
}

func (m *Manager) HandleDeactivate(appName, bundleID string) {
	m.logger.Debug("[Hotkey] App deactivated", zap.String("bundle_id", bundleID))

	if m.handleDeactivateCallback != nil {
		m.handleDeactivateCallback(appName, bundleID)
	}
}

// Register registers a global hotkey
func (m *Manager) Register(keyString string, callback Callback) (HotkeyID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse key string
	var keyCode, modifiers C.int
	cKeyString := C.CString(keyString)
	defer C.free(unsafe.Pointer(cKeyString))

	result := C.parseKeyString(cKeyString, &keyCode, &modifiers)
	if result == 0 {
		return 0, fmt.Errorf("failed to parse key string: %s", keyString)
	}

	// Generate hotkey ID
	hotkeyID := m.nextID
	m.nextID++

	// Register hotkey
	success := C.registerHotkey(keyCode, modifiers, C.int(hotkeyID),
		C.HotkeyCallback(C.hotkeyCallbackBridge), nil)

	if success == 0 {
		return 0, fmt.Errorf("failed to register hotkey: %s", keyString)
	}

	// Store callback
	m.callbacks[hotkeyID] = callback

	m.logger.Info("Registered hotkey",
		zap.String("key", keyString),
		zap.Int("id", int(hotkeyID)))

	return hotkeyID, nil
}

// Unregister unregisters a hotkey
func (m *Manager) Unregister(id HotkeyID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	C.unregisterHotkey(C.int(id))
	delete(m.callbacks, id)

	m.logger.Info("Unregistered hotkey", zap.Int("id", int(id)))
}

// UnregisterAll unregisters all hotkeys
func (m *Manager) UnregisterAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	C.unregisterAllHotkeys()
	m.callbacks = make(map[HotkeyID]Callback)

	m.logger.Info("Unregistered all hotkeys")
}

// handleCallback handles a hotkey callback from C
func (m *Manager) handleCallback(hotkeyID HotkeyID) {
	m.mu.RLock()
	callback, ok := m.callbacks[hotkeyID]
	m.mu.RUnlock()

	if ok && callback != nil {
		m.logger.Debug("Hotkey pressed", zap.Int("id", int(hotkeyID)))
		callback()
	}
}

// Global manager instance for C callbacks
var globalManager *Manager

// SetGlobalManager sets the global manager for C callbacks
func SetGlobalManager(m *Manager) {
	globalManager = m
}

//export hotkeyCallbackBridge
func hotkeyCallbackBridge(hotkeyID C.int, userData unsafe.Pointer) {
	if globalManager != nil {
		globalManager.handleCallback(HotkeyID(hotkeyID))
	}
}
