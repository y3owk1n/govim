package electron

import (
	"strings"
	"sync"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/appwatcher"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

const (
	electronAttributeName = "AXManualAccessibility"
	enhancedAttributeName = "AXEnhancedUserInterface"
)

// ElectronManager handles Electron-specific functionality
type ElectronManager struct {
	watcher           *appwatcher.Watcher
	additionalBundles []string
}

var (
	electronPIDsMu      sync.Mutex
	electronEnabledPIDs = make(map[int]struct{})
)

// NewElectronManager creates a new ElectronManager
func NewElectronManager(additionalBundles ...string) *ElectronManager {
	em := &ElectronManager{
		watcher:           appwatcher.New(),
		additionalBundles: additionalBundles,
	}

	// Implement bridge.AppWatcher interface
	bridge.SetAppWatcher(bridge.AppWatcher(em))
	return em
}

// Start initializes the Electron manager and sets up watchers
func (em *ElectronManager) Start() {
	bridge.StartAppWatcher()
}

// Stop cleans up the Electron manager
func (em *ElectronManager) Stop() {
	bridge.StopAppWatcher()
}

// HandleLaunch implements bridge.AppWatcher
func (em *ElectronManager) HandleLaunch(appName, bundleID string) {
	logger.Debug("App launched", zap.String("bundle_id", bundleID))

	if !ShouldEnableElectronSupport(bundleID, em.additionalBundles) {
		logger.Debug("App does not require Electron support",
			zap.String("bundle_id", bundleID))
		return
	}

	app := accessibility.GetApplicationByBundleID(bundleID)

	info, err := app.GetInfo()
	if err != nil {
		logger.Debug("Failed to inspect app window", zap.Error(err))
		return
	}

	pid := info.PID

	if pid <= 0 {
		return
	}

	logger.Debug("App requires Electron support",
		zap.String("bundle_id", bundleID),
		zap.Int("pid", pid))

	ensureElectronAccessibility(pid, bundleID)
}

// HandleTerminate implements bridge.AppWatcher
func (em *ElectronManager) HandleTerminate(appName, bundleID string) {
	logger.Debug("App terminated", zap.String("bundle_id", bundleID))
}

// HandleActivate implements bridge.AppWatcher
func (em *ElectronManager) HandleActivate(appName, bundleID string) {
	logger.Debug("App activated", zap.String("bundle_id", bundleID))

	if !ShouldEnableElectronSupport(bundleID, em.additionalBundles) {
		logger.Debug("App does not require Electron support",
			zap.String("bundle_id", bundleID))
		return
	}

	app := accessibility.GetApplicationByBundleID(bundleID)

	info, err := app.GetInfo()
	if err != nil {
		logger.Debug("Failed to inspect app window", zap.Error(err))
		return
	}

	pid := info.PID

	if pid <= 0 {
		return
	}

	logger.Debug("App requires Electron support",
		zap.String("bundle_id", bundleID),
		zap.Int("pid", pid))

	ensureElectronAccessibility(pid, bundleID)
}

// HandleDeactivate implements bridge.AppWatcher
func (em *ElectronManager) HandleDeactivate(appName, bundleID string) {
	logger.Debug("App deactivated", zap.String("bundle_id", bundleID))
}

func ensureElectronAccessibility(pid int, bundleID string) bool {
	electronPIDsMu.Lock()
	_, already := electronEnabledPIDs[pid]
	electronPIDsMu.Unlock()

	if already {
		return true
	}

	successSetElectron := bridge.SetApplicationAttribute(pid, electronAttributeName, true)

	if successSetElectron {
		logger.Debug("Enabled AXManualAccessibility", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
	}

	successSetEnhanced := bridge.SetApplicationAttribute(pid, enhancedAttributeName, true)

	if successSetEnhanced {
		logger.Debug("Enabled AXEnhancedUserInterface", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
	}

	if !successSetEnhanced && !successSetElectron {
		logger.Warn("Failed to enable AXManualAccessibility or AXEnhancedUserInterface", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
		return false
	}

	electronPIDsMu.Lock()
	electronEnabledPIDs[pid] = struct{}{}
	electronPIDsMu.Unlock()
	return true
}

var KnownElectronBundles = []string{
	// browsers
	// NOTE: why is this here? It's not an electron app, but the same logic to enable `AXEnhancedUserInterface`. Put it here first until we have a better manager for this.
	// Chromium-based browsers
	"net.imput.helium",
	"com.google.Chrome",
	"com.brave.Browser",
	"company.thebrowser.Browser",
	// Firefox-based browsers
	"org.mozilla.firefox",
	"app.zen-browser.zen",
	// electrons
	"com.microsoft.VSCode",
	"com.exafunction.windsurf",
	"com.todesktop.230313mzl4w4u92", // cursor
	"com.tinyspeck.slackmacgap",
	"com.spotify.client",
	"md.obsidian",
}

// ShouldEnableElectronSupport determines if the provided bundle identifier
// should have Electron accessibility manually toggled based on defaults and
// user-specified overrides.
func ShouldEnableElectronSupport(bundleID string, additionalBundles []string) bool {
	if bundleID == "" {
		return false
	}

	if matchesAdditionalBundle(bundleID, additionalBundles) {
		return true
	}

	return IsLikelyElectronBundle(bundleID)
}

// IsLikelyElectronBundle returns true if the provided bundle identifier
// matches a known Electron signature.
func IsLikelyElectronBundle(bundleID string) bool {
	lower := strings.ToLower(strings.TrimSpace(bundleID))
	if lower == "" {
		return false
	}

	for _, exact := range KnownElectronBundles {
		if strings.EqualFold(strings.TrimSpace(exact), lower) {
			return true
		}
	}

	return false
}

func matchesAdditionalBundle(bundleID string, additionalBundles []string) bool {
	if len(additionalBundles) == 0 {
		return false
	}

	lower := strings.ToLower(strings.TrimSpace(bundleID))
	for _, candidate := range additionalBundles {
		trimmed := strings.ToLower(strings.TrimSpace(candidate))
		if trimmed == "" {
			continue
		}
		if strings.HasSuffix(trimmed, "*") {
			prefix := strings.TrimSuffix(trimmed, "*")
			if strings.HasPrefix(lower, prefix) {
				return true
			}
		} else if lower == trimmed {
			return true
		}
	}

	return false
}
