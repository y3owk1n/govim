package accessibility

import "strings"

var (
	// KnownElectronPrefixes lists bundle identifier prefixes commonly backed by Electron.
	KnownElectronPrefixes = []string{
		"com.microsoft.vscode",
		"com.microsoft.vscodeinsiders",
		"com.slack.",
		"com.github.",
		"com.todesktop.",
		"com.zoom.",
		"md.obsidian",
	}

	// KnownElectronExact matches bundle identifiers observed to require manual accessibility.
	KnownElectronExact = []string{
		"com.sindresorhus.helium",
		"com.sindresorhus.helium2",
		"com.tinyspeck.slackmacgap",
		"com.microsoft.teams",
		"com.exafunction.windsurf",
	}
)

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

	for _, exact := range KnownElectronExact {
		if strings.EqualFold(strings.TrimSpace(exact), lower) {
			return true
		}
	}

	for _, prefix := range KnownElectronPrefixes {
		p := strings.ToLower(strings.TrimSpace(prefix))
		if p == "" {
			continue
		}
		p = strings.TrimSuffix(p, "*")
		if strings.HasPrefix(lower, p) {
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
