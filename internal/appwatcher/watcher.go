package appwatcher

import (
	"sync"

	"github.com/y3owk1n/neru/internal/bridge"
)

// Callback function type for application events
type AppCallback func(appName string, bundleID string)

// Watcher represents an application watcher
type Watcher struct {
	mu sync.RWMutex
	// Callbacks for different events
	launchCallbacks       []AppCallback
	terminateCallbacks    []AppCallback
	activateCallbacks     []AppCallback
	deactivateCallbacks   []AppCallback
	screenChangeCallbacks []func()
}

func NewWatcher() *Watcher {
	w := &Watcher{}

	bridge.SetAppWatcher(bridge.AppWatcher(w))

	return w
}

func (w *Watcher) Start() {
	bridge.StartAppWatcher()
}

func (w *Watcher) Stop() {
	bridge.StopAppWatcher()
}

// OnScreenParametersChanged registers a callback for screen parameter change events
func (w *Watcher) OnScreenParametersChanged(callback func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.screenChangeCallbacks = append(w.screenChangeCallbacks, callback)
}

// OnTerminate registers a callback for application termination events
func (w *Watcher) OnTerminate(callback AppCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.terminateCallbacks = append(w.terminateCallbacks, callback)
}

// OnActivate registers a callback for application activation events
func (w *Watcher) OnActivate(callback AppCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.activateCallbacks = append(w.activateCallbacks, callback)
}

// OnDeactivate registers a callback for application deactivation events
func (w *Watcher) OnDeactivate(callback AppCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.deactivateCallbacks = append(w.deactivateCallbacks, callback)
}

// handleLaunch is called from the bridge when an application launches
func (w *Watcher) HandleLaunch(appName, bundleID string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, callback := range w.launchCallbacks {
		callback(appName, bundleID)
	}
}

// handleTerminate is called from the bridge when an application terminates
func (w *Watcher) HandleTerminate(appName, bundleID string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, callback := range w.terminateCallbacks {
		callback(appName, bundleID)
	}
}

// handleActivate is called from the bridge when an application is activated
func (w *Watcher) HandleActivate(appName, bundleID string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, callback := range w.activateCallbacks {
		callback(appName, bundleID)
	}
}

// handleDeactivate is called from the bridge when an application is deactivated
func (w *Watcher) HandleDeactivate(appName, bundleID string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, callback := range w.deactivateCallbacks {
		callback(appName, bundleID)
	}
}

// HandleScreenParametersChanged is called from the bridge when display parameters change
func (w *Watcher) HandleScreenParametersChanged() {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, callback := range w.screenChangeCallbacks {
		callback()
	}
}
