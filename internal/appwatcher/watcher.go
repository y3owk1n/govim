package appwatcher

import (
	"sync"
)

// Callback function type for application events
type AppCallback func(appName string, bundleID string)

// Watcher represents an application watcher
type Watcher struct {
	mu sync.RWMutex
	// Callbacks for different events
	launchCallbacks     []AppCallback
	terminateCallbacks  []AppCallback
	activateCallbacks   []AppCallback
	deactivateCallbacks []AppCallback
}

// New creates a new application watcher
func New() *Watcher {
	w := &Watcher{}
	return w
}

// OnLaunch registers a callback for application launch events
func (w *Watcher) OnLaunch(callback AppCallback) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.launchCallbacks = append(w.launchCallbacks, callback)
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
