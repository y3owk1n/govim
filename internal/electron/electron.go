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
	watcher                   *appwatcher.Watcher
	additionalElectronBundles []string
	additionalChromiumBundles []string
	additionalFirefoxBundles  []string
}

var (
	electronPIDsMu      sync.Mutex
	electronEnabledPIDs = make(map[int]struct{})
	chromiumPIDsMu      sync.Mutex
	chromiumEnabledPIDs = make(map[int]struct{})
	firefoxPIDsMu       sync.Mutex
	firefoxEnabledPIDs  = make(map[int]struct{})
)

// NewElectronManager creates a new ElectronManager
func NewElectronManager(additionalElectronBundles []string, additionalChromiumBundles []string, additionalFirefoxBundles []string) *ElectronManager {
	em := &ElectronManager{
		watcher:                   appwatcher.New(),
		additionalElectronBundles: additionalElectronBundles,
		additionalChromiumBundles: additionalChromiumBundles,
		additionalFirefoxBundles:  additionalFirefoxBundles,
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
}

// HandleTerminate implements bridge.AppWatcher
func (em *ElectronManager) HandleTerminate(appName, bundleID string) {
	logger.Debug("App terminated", zap.String("bundle_id", bundleID))
}

// HandleActivate implements bridge.AppWatcher
func (em *ElectronManager) HandleActivate(appName, bundleID string) {
	logger.Debug("App activated", zap.String("bundle_id", bundleID))

	if ShouldEnableElectronSupport(bundleID, em.additionalElectronBundles) {
		ensureElectronAccessibility(bundleID)
		return
	}

	if ShouldEnableChromiumSupport(bundleID, em.additionalChromiumBundles) {
		ensureChromiumAccessibility(bundleID)
		return
	}

	if ShouldEnableFirefoxSupport(bundleID, em.additionalFirefoxBundles) {
		ensureFirefoxAccessibility(bundleID)
		return
	}
}

// HandleDeactivate implements bridge.AppWatcher
func (em *ElectronManager) HandleDeactivate(appName, bundleID string) {
	logger.Debug("App deactivated", zap.String("bundle_id", bundleID))
}

func ensureElectronAccessibility(bundleID string) bool {
	app := accessibility.GetApplicationByBundleID(bundleID)

	info, err := app.GetInfo()
	if err != nil {
		logger.Debug("Failed to inspect app window", zap.Error(err))
		return false
	}

	pid := info.PID

	if pid <= 0 {
		logger.Debug("No PID found for app", zap.String("bundle_id", bundleID))
		return false
	}

	logger.Debug("App requires Electron support",
		zap.String("bundle_id", bundleID),
		zap.Int("pid", pid))

	electronPIDsMu.Lock()
	_, already := electronEnabledPIDs[pid]
	electronPIDsMu.Unlock()

	if already {
		logger.Debug("Already enabled Electron support", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
		return true
	}

	successSetElectron := bridge.SetApplicationAttribute(pid, electronAttributeName, true)

	if successSetElectron {
		logger.Debug("Enabled AXManualAccessibility", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
	}

	if !successSetElectron {
		logger.Warn("Failed to enable AXManualAccessibility", zap.Int("pid", pid), zap.String("bundle_id", bundleID))
		return false
	}

	electronPIDsMu.Lock()
	electronEnabledPIDs[pid] = struct{}{}
	electronPIDsMu.Unlock()
	return true
}

func ensureChromiumAccessibility(bundleID string) bool {
	app := accessibility.GetApplicationByBundleID(bundleID)

	info, err := app.GetInfo()
	if err != nil {
		logger.Debug("Failed to inspect app window", zap.Error(err))
		return false
	}

	pid := info.PID

	if pid <= 0 {
		logger.Debug("No PID found for app", zap.String("bundle_id", bundleID))
		return false
	}

	logger.Debug("Chromium requires AXEnhancedUserInterface",
		zap.String("bundle_id", bundleID),
		zap.Int("pid", pid))

	chromiumPIDsMu.Lock()
	_, already := chromiumEnabledPIDs[pid]
	chromiumPIDsMu.Unlock()

	if already {
		logger.Debug("Already enabled Chromium support", zap.String("bundle_id", bundleID))
		return true
	}

	successSetChromium := bridge.SetApplicationAttribute(pid, enhancedAttributeName, true)

	if successSetChromium {
		logger.Debug("Enabled AXEnhancedUserInterface", zap.String("bundle_id", bundleID))
	}

	if !successSetChromium {
		logger.Warn("Failed to enable AXEnhancedUserInterface", zap.String("bundle_id", bundleID))
		return false
	}

	chromiumPIDsMu.Lock()
	chromiumEnabledPIDs[pid] = struct{}{}
	chromiumPIDsMu.Unlock()
	return true
}

func ensureFirefoxAccessibility(bundleID string) bool {
	app := accessibility.GetApplicationByBundleID(bundleID)

	info, err := app.GetInfo()
	if err != nil {
		logger.Debug("Failed to inspect app window", zap.Error(err))
		return false
	}

	pid := info.PID

	if pid <= 0 {
		logger.Debug("No PID found for app", zap.String("bundle_id", bundleID))
		return false
	}

	logger.Debug("Firefox requires AXEnhancedUserInterface support",
		zap.String("bundle_id", bundleID),
		zap.Int("pid", pid))

	firefoxPIDsMu.Lock()
	_, already := firefoxEnabledPIDs[pid]
	firefoxPIDsMu.Unlock()

	if already {
		logger.Debug("Already enabled Firefox support", zap.String("bundle_id", bundleID))
		return true
	}

	successSetFirefox := bridge.SetApplicationAttribute(pid, enhancedAttributeName, true)

	if successSetFirefox {
		logger.Debug("Enabled AXEnhancedUserInterface", zap.String("bundle_id", bundleID))
	}

	if !successSetFirefox {
		logger.Warn("Failed to enable AXEnhancedUserInterface", zap.String("bundle_id", bundleID))
		return false
	}

	firefoxPIDsMu.Lock()
	firefoxEnabledPIDs[pid] = struct{}{}
	firefoxPIDsMu.Unlock()
	return true
}

var KnownChromiumBundles = []string{
	"net.imput.helium",
	"com.google.Chrome",
	"com.brave.Browser",
	"company.thebrowser.Browser",
}

var KnownFirefoxBundles = []string{
	"org.mozilla.firefox",
	"app.zen-browser.zen",
}

var KnownElectronBundles = []string{
	// electrons
	"com.microsoft.VSCode",
	"com.exafunction.windsurf",
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

func ShouldEnableChromiumSupport(bundleID string, additionalBundles []string) bool {
	if bundleID == "" {
		return false
	}

	if matchesAdditionalBundle(bundleID, additionalBundles) {
		return true
	}

	return IsLikelyChromiumBundle(bundleID)
}

func ShouldEnableFirefoxSupport(bundleID string, additionalBundles []string) bool {
	if bundleID == "" {
		return false
	}

	if matchesAdditionalBundle(bundleID, additionalBundles) {
		return true
	}

	return IsLikelyFirefoxBundle(bundleID)
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

func IsLikelyChromiumBundle(bundleID string) bool {
	lower := strings.ToLower(strings.TrimSpace(bundleID))
	if lower == "" {
		return false
	}

	for _, exact := range KnownChromiumBundles {
		if strings.EqualFold(strings.TrimSpace(exact), lower) {
			return true
		}
	}

	return false
}

func IsLikelyFirefoxBundle(bundleID string) bool {
	lower := strings.ToLower(strings.TrimSpace(bundleID))
	if lower == "" {
		return false
	}

	for _, exact := range KnownFirefoxBundles {
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
