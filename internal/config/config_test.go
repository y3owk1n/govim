package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Hints.HintCharacters != "asdfghjkl" {
		t.Errorf("Expected hint_characters to be 'asdfghjkl', got '%s'", cfg.Hints.HintCharacters)
	}

	if cfg.Logging.LogLevel != "info" {
		t.Errorf("Expected log_level to be 'info', got '%s'", cfg.Logging.LogLevel)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "invalid log level",
			modify: func(c *Config) {
				c.Logging.LogLevel = "invalid"
			},
			wantErr: true,
		},
		{
			name: "invalid opacity",
			modify: func(c *Config) {
				c.Hints.Opacity = 1.5
			},
			wantErr: true,
		},
		{
			name: "too few hint characters",
			modify: func(c *Config) {
				c.Hints.HintCharacters = "a"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create and save config
	cfg := DefaultConfig()
	cfg.Hints.FontSize = 16
	cfg.Scroll.ScrollSpeed = 100

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loaded.Hints.FontSize != 16 {
		t.Errorf("Expected FontSize to be 16, got %d", loaded.Hints.FontSize)
	}
	if loaded.Scroll.ScrollSpeed != 100 {
		t.Errorf("Expected ScrollSpeed to be 100, got %d", loaded.Scroll.ScrollSpeed)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	// Loading non-existent file should return default config
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got: %v", err)
	}

	if cfg.Hints.HintCharacters != "asdfghjkl" {
		t.Error("Expected default config when file doesn't exist")
	}
}

func TestConfigWithAppSpecificSettings(t *testing.T) {
	// Create a config with app-specific settings
	cfg := DefaultConfig()

	// Add app-specific configurations
	cfg.Accessibility.AppConfigs = []AppConfig{
		{
			BundleID:             "com.example.app1",
			AdditionalClickable:  []string{"CustomButton1", "CustomLink1"},
			AdditionalScrollable: []string{"CustomScroll1"},
		},
		{
			BundleID:             "com.example.app2",
			AdditionalClickable:  []string{"CustomButton2"},
			AdditionalScrollable: []string{"CustomScroll2", "CustomPanel2"},
		},
	}

	// Test getting clickable roles for specific app
	clickableRoles := cfg.GetClickableRolesForApp("com.example.app1")

	// Should include both default and app-specific roles
	defaultCount := len(cfg.Accessibility.ClickableRoles)
	expectedCount := defaultCount + 2 // Default + 2 custom roles

	if len(clickableRoles) != expectedCount {
		t.Errorf("Expected %d clickable roles, got %d", expectedCount, len(clickableRoles))
	}

	// Check if app-specific roles are included
	found := slices.Contains(clickableRoles, "CustomButton1")

	if !found {
		t.Errorf("App-specific role 'CustomButton1' not found in clickable roles")
	}

	// Test getting scrollable roles for specific app
	scrollableRoles := cfg.GetScrollableRolesForApp("com.example.app2")

	// Should include both default and app-specific roles
	defaultScrollCount := len(cfg.Accessibility.ScrollableRoles)
	expectedScrollCount := defaultScrollCount + 2 // Default + 2 custom roles

	if len(scrollableRoles) != expectedScrollCount {
		t.Errorf("Expected %d scrollable roles, got %d", expectedScrollCount, len(scrollableRoles))
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("Expected non-empty config path")
	}

	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, "Library", "Application Support", "neru", "config.toml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestIsAppExcluded(t *testing.T) {
	tests := []struct {
		name         string
		excludedApps []string
		bundleID     string
		expected     bool
	}{
		{
			name:         "empty exclusion list",
			excludedApps: []string{},
			bundleID:     "com.example.app",
			expected:     false,
		},
		{
			name:         "app not in exclusion list",
			excludedApps: []string{"com.apple.Terminal", "com.googlecode.iterm2"},
			bundleID:     "com.example.app",
			expected:     false,
		},
		{
			name:         "app in exclusion list - exact match",
			excludedApps: []string{"com.apple.Terminal", "com.example.app", "com.googlecode.iterm2"},
			bundleID:     "com.example.app",
			expected:     true,
		},
		{
			name:         "app in exclusion list - case insensitive",
			excludedApps: []string{"com.apple.Terminal", "COM.EXAMPLE.APP", "com.googlecode.iterm2"},
			bundleID:     "com.example.app",
			expected:     true,
		},
		{
			name:         "app in exclusion list - with whitespace",
			excludedApps: []string{"com.apple.Terminal", " com.example.app ", "com.googlecode.iterm2"},
			bundleID:     "com.example.app",
			expected:     true,
		},
		{
			name:         "bundle ID with whitespace",
			excludedApps: []string{"com.apple.Terminal", "com.example.app", "com.googlecode.iterm2"},
			bundleID:     " com.example.app ",
			expected:     true,
		},
		{
			name:         "both with whitespace and different case",
			excludedApps: []string{"com.apple.Terminal", " COM.EXAMPLE.APP ", "com.googlecode.iterm2"},
			bundleID:     " com.example.app ",
			expected:     true,
		},
		{
			name:         "empty bundle ID",
			excludedApps: []string{"com.apple.Terminal", "com.example.app"},
			bundleID:     "",
			expected:     false,
		},
		{
			name:         "nil exclusion list",
			excludedApps: nil,
			bundleID:     "com.example.app",
			expected:     false,
		},
		{
			name:         "partial match should not exclude",
			excludedApps: []string{"com.example"},
			bundleID:     "com.example.app",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.General.ExcludedApps = tt.excludedApps

			result := cfg.IsAppExcluded(tt.bundleID)
			if result != tt.expected {
				t.Errorf("IsAppExcluded() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultConfigExcludedApps(t *testing.T) {
	cfg := DefaultConfig()

	// Default config should have empty excluded apps list
	if cfg.General.ExcludedApps == nil {
		t.Error("Expected ExcludedApps to be initialized, got nil")
	}

	if len(cfg.General.ExcludedApps) != 0 {
		t.Errorf("Expected empty ExcludedApps list by default, got %d items", len(cfg.General.ExcludedApps))
	}
}

func TestLoadAndSaveWithExcludedApps(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create config with excluded apps
	cfg := DefaultConfig()
	cfg.General.ExcludedApps = []string{
		"com.apple.Terminal",
		"com.googlecode.iterm2",
		"com.microsoft.rdc.macos",
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify excluded apps were saved and loaded correctly
	if len(loaded.General.ExcludedApps) != 3 {
		t.Errorf("Expected 3 excluded apps, got %d", len(loaded.General.ExcludedApps))
	}

	expectedApps := []string{
		"com.apple.Terminal",
		"com.googlecode.iterm2",
		"com.microsoft.rdc.macos",
	}

	for i, expected := range expectedApps {
		if i >= len(loaded.General.ExcludedApps) || loaded.General.ExcludedApps[i] != expected {
			t.Errorf("Expected excluded app at index %d to be '%s', got '%s'",
				i, expected, loaded.General.ExcludedApps[i])
		}
	}

	// Test that exclusion works correctly
	if !loaded.IsAppExcluded("com.apple.Terminal") {
		t.Error("Expected com.apple.Terminal to be excluded")
	}

	if loaded.IsAppExcluded("com.example.notexcluded") {
		t.Error("Expected com.example.notexcluded to not be excluded")
	}
}
