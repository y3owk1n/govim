package accessibility

import "testing"

func TestIsLikelyElectronBundle(t *testing.T) {
	tests := []struct {
		name     string
		bundleID string
		want     bool
	}{
		// Exact matches
		{"Windsurf", "com.exafunction.windsurf", true},
		{"Teams", "com.microsoft.teams", true},
		{"Helium", "com.sindresorhus.helium", true},
		{"Slack", "com.tinyspeck.slackmacgap", true},

		// Prefix matches
		{"VS Code", "com.microsoft.VSCode", true},
		{"VS Code Insiders", "com.microsoft.VSCodeInsiders", true},
		{"Slack Desktop", "com.slack.Slack", true},
		{"GitHub Desktop", "com.github.GitHubClient", true},
		{"Zoom", "com.zoom.us.Zoom", true},
		{"Obsidian", "md.obsidian", true},
		{"ToDesktop App", "com.todesktop.myapp", true},

		// Non-Electron apps
		{"Safari", "com.apple.Safari", false},
		{"Finder", "com.apple.finder", false},
		{"Chrome", "com.google.Chrome", false},
		{"Firefox", "org.mozilla.firefox", false},

		// Edge cases
		{"Empty", "", false},
		{"Whitespace", "   ", false},
		{"Case insensitive", "COM.MICROSOFT.VSCODE", true},
		{"With spaces", " com.microsoft.VSCode ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLikelyElectronBundle(tt.bundleID)
			if got != tt.want {
				t.Errorf("IsLikelyElectronBundle(%q) = %v, want %v", tt.bundleID, got, tt.want)
			}
		})
	}
}

func TestShouldEnableElectronSupport(t *testing.T) {
	tests := []struct {
		name              string
		bundleID          string
		additionalBundles []string
		want              bool
	}{
		{
			name:              "Known Electron app",
			bundleID:          "com.exafunction.windsurf",
			additionalBundles: nil,
			want:              true,
		},
		{
			name:              "Chrome via additional bundles",
			bundleID:          "com.google.Chrome",
			additionalBundles: []string{"com.google.Chrome"},
			want:              true,
		},
		{
			name:              "Chrome with wildcard",
			bundleID:          "com.google.Chrome.canary",
			additionalBundles: []string{"com.google.Chrome*"},
			want:              true,
		},
		{
			name:              "Brave via additional bundles",
			bundleID:          "com.brave.Browser",
			additionalBundles: []string{"com.brave.Browser"},
			want:              true,
		},
		{
			name:              "Non-Electron without additional",
			bundleID:          "com.apple.Safari",
			additionalBundles: nil,
			want:              false,
		},
		{
			name:              "Empty bundle ID",
			bundleID:          "",
			additionalBundles: []string{"com.google.Chrome"},
			want:              false,
		},
		{
			name:              "Case insensitive additional",
			bundleID:          "COM.GOOGLE.CHROME",
			additionalBundles: []string{"com.google.chrome"},
			want:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldEnableElectronSupport(tt.bundleID, tt.additionalBundles)
			if got != tt.want {
				t.Errorf("ShouldEnableElectronSupport(%q, %v) = %v, want %v",
					tt.bundleID, tt.additionalBundles, got, tt.want)
			}
		})
	}
}

func TestMatchesAdditionalBundle(t *testing.T) {
	tests := []struct {
		name              string
		bundleID          string
		additionalBundles []string
		want              bool
	}{
		{
			name:              "Exact match",
			bundleID:          "com.google.Chrome",
			additionalBundles: []string{"com.google.Chrome"},
			want:              true,
		},
		{
			name:              "Wildcard match",
			bundleID:          "com.google.Chrome.canary",
			additionalBundles: []string{"com.google.Chrome*"},
			want:              true,
		},
		{
			name:              "No match",
			bundleID:          "com.apple.Safari",
			additionalBundles: []string{"com.google.Chrome"},
			want:              false,
		},
		{
			name:              "Empty additional bundles",
			bundleID:          "com.google.Chrome",
			additionalBundles: []string{},
			want:              false,
		},
		{
			name:              "Nil additional bundles",
			bundleID:          "com.google.Chrome",
			additionalBundles: nil,
			want:              false,
		},
		{
			name:              "Multiple bundles with match",
			bundleID:          "com.brave.Browser",
			additionalBundles: []string{"com.google.Chrome", "com.brave.Browser", "org.mozilla.firefox"},
			want:              true,
		},
		{
			name:              "Case insensitive",
			bundleID:          "COM.GOOGLE.CHROME",
			additionalBundles: []string{"com.google.chrome"},
			want:              true,
		},
		{
			name:              "Whitespace handling",
			bundleID:          " com.google.Chrome ",
			additionalBundles: []string{" com.google.Chrome "},
			want:              true,
		},
		{
			name:              "Empty string in bundles",
			bundleID:          "com.google.Chrome",
			additionalBundles: []string{"", "com.google.Chrome"},
			want:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAdditionalBundle(tt.bundleID, tt.additionalBundles)
			if got != tt.want {
				t.Errorf("matchesAdditionalBundle(%q, %v) = %v, want %v",
					tt.bundleID, tt.additionalBundles, got, tt.want)
			}
		})
	}
}
