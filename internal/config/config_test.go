package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.General.HintCharacters != "asdfghjkl" {
		t.Errorf("Expected hint_characters to be 'asdfghjkl', got '%s'", cfg.General.HintCharacters)
	}

	if cfg.General.HintStyle != "alphabet" {
		t.Errorf("Expected hint_style to be 'alphabet', got '%s'", cfg.General.HintStyle)
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
			name: "invalid hint style",
			modify: func(c *Config) {
				c.General.HintStyle = "invalid"
			},
			wantErr: true,
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
				c.General.HintCharacters = "a"
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

	if cfg.General.HintCharacters != "asdfghjkl" {
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
	found := false
	for _, role := range clickableRoles {
		if role == "CustomButton1" {
			found = true
			break
		}
	}

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
	expectedPath := filepath.Join(homeDir, "Library", "Application Support", "govim", "config.toml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}
