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
	cfg.General.Debug = true
	cfg.Hints.FontSize = 16

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if !loaded.General.Debug {
		t.Error("Expected Debug to be true")
	}
	if loaded.Hints.FontSize != 16 {
		t.Errorf("Expected FontSize to be 16, got %d", loaded.Hints.FontSize)
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
