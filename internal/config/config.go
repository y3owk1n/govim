package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the complete application configuration
type Config struct {
	General      GeneralConfig      `toml:"general"`
	Hotkeys      HotkeysConfig      `toml:"hotkeys"`
	Scroll       ScrollConfig       `toml:"scroll"`
	Hints        HintsConfig        `toml:"hints"`
	Appearance   AppearanceConfig   `toml:"appearance"`
	Performance  PerformanceConfig  `toml:"performance"`
	Logging      LoggingConfig      `toml:"logging"`
	Apps         AppsConfig         `toml:"apps"`
	Experimental ExperimentalConfig `toml:"experimental"`
}

type GeneralConfig struct {
	HintCharacters           string `toml:"hint_characters"`
	HintStyle                string `toml:"hint_style"`
	AnimationDurationMs      int    `toml:"animation_duration_ms"`
	AccessibilityCheckOnStart bool   `toml:"accessibility_check_on_start"`
	Debug                    bool   `toml:"debug"`
}

type HotkeysConfig struct {
	ActivateHintMode    string `toml:"activate_hint_mode"`
	ActivateScrollMode  string `toml:"activate_scroll_mode"`
	ExitMode            string `toml:"exit_mode"`
	ShowCommandPalette  string `toml:"show_command_palette"`
	ReloadConfig        string `toml:"reload_config"`
}

type ScrollConfig struct {
	ScrollSpeed            int     `toml:"scroll_speed"`
	HighlightScrollArea    bool    `toml:"highlight_scroll_area"`
	HighlightColor         string  `toml:"highlight_color"`
	HighlightWidth         int     `toml:"highlight_width"`
	PageHeight             int     `toml:"page_height"`
	HalfPageMultiplier     float64 `toml:"half_page_multiplier"`
	FullPageMultiplier     float64 `toml:"full_page_multiplier"`
	ScrollToEdgeIterations int     `toml:"scroll_to_edge_iterations"`
	ScrollToEdgeDelta      int     `toml:"scroll_to_edge_delta"`
}

type HintsConfig struct {
	FontSize        int     `toml:"font_size"`
	FontFamily      string  `toml:"font_family"`
	BackgroundColor string  `toml:"background_color"`
	TextColor       string  `toml:"text_color"`
	BorderRadius    int     `toml:"border_radius"`
	Padding         int     `toml:"padding"`
	BorderWidth     int     `toml:"border_width"`
	BorderColor     string  `toml:"border_color"`
	Opacity         float64 `toml:"opacity"`
}

type AppearanceConfig struct {
	OverlayOpacity   float64 `toml:"overlay_opacity"`
	DarkModeSupport  bool    `toml:"dark_mode_support"`
	UseSystemAccent  bool    `toml:"use_system_accent"`
}

type PerformanceConfig struct {
	MaxHintsDisplayed      int `toml:"max_hints_displayed"`
	DebounceMs             int `toml:"debounce_ms"`
	UseMetalAcceleration   bool `toml:"use_metal_acceleration"`
	CacheDurationMs        int `toml:"cache_duration_ms"`
	MaxConcurrentQueries   int `toml:"max_concurrent_queries"`
}

type LoggingConfig struct {
	LogLevel           string `toml:"log_level"`
	LogFile            string `toml:"log_file"`
	MaxLogSizeMB       int    `toml:"max_log_size_mb"`
	MaxLogBackups      int    `toml:"max_log_backups"`
	StructuredLogging  bool   `toml:"structured_logging"`
}

type AppsConfig struct {
	Exclude     []AppExclude     `toml:"exclude"`
	CustomHints []AppCustomHints `toml:"custom_hints"`
}

type AppExclude struct {
	BundleID string `toml:"bundle_id"`
	Reason   string `toml:"reason"`
}

type AppCustomHints struct {
	BundleID       string `toml:"bundle_id"`
	UseWebElements bool   `toml:"use_web_elements"`
	HintDelayMs    int    `toml:"hint_delay_ms"`
}

type ExperimentalConfig struct {
	EnableExperimental      bool `toml:"enable_experimental"`
	AlternativeHintAlgorithm bool `toml:"alternative_hint_algorithm"`
	GestureSupport          bool `toml:"gesture_support"`
	PluginSystem            bool `toml:"plugin_system"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			HintCharacters:           "asdfghjkl",
			HintStyle:                "alphabet",
			AnimationDurationMs:      150,
			AccessibilityCheckOnStart: true,
			Debug:                    false,
		},
		Hotkeys: HotkeysConfig{
			ActivateHintMode:   "Cmd+Shift+Space",
			ActivateScrollMode: "Cmd+Shift+J",
			ExitMode:           "Escape",
			ShowCommandPalette: "Cmd+Shift+P",
			ReloadConfig:       "Cmd+Shift+R",
		},
		Scroll: ScrollConfig{
			ScrollSpeed:            50,
			HighlightScrollArea:    true,
			HighlightColor:         "#FF0000",
			HighlightWidth:         2,
			PageHeight:             1200,
			HalfPageMultiplier:     0.5,
			FullPageMultiplier:     0.9,
			ScrollToEdgeIterations: 20,
			ScrollToEdgeDelta:      5000,
		},
		Hints: HintsConfig{
			FontSize:        14,
			FontFamily:      "SF Mono",
			BackgroundColor: "#FFD700",
			TextColor:       "#000000",
			BorderRadius:    4,
			Padding:         4,
			BorderWidth:     1,
			BorderColor:     "#000000",
			Opacity:         0.95,
		},
		Appearance: AppearanceConfig{
			OverlayOpacity:  0.95,
			DarkModeSupport: true,
			UseSystemAccent: false,
		},
		Performance: PerformanceConfig{
			MaxHintsDisplayed:    200,
			DebounceMs:           50,
			UseMetalAcceleration: true,
			CacheDurationMs:      100,
			MaxConcurrentQueries: 10,
		},
		Logging: LoggingConfig{
			LogLevel:          "info",
			LogFile:           "",
			MaxLogSizeMB:      10,
			MaxLogBackups:     3,
			StructuredLogging: true,
		},
		Apps: AppsConfig{
			Exclude:     []AppExclude{},
			CustomHints: []AppCustomHints{},
		},
		Experimental: ExperimentalConfig{
			EnableExperimental:      false,
			AlternativeHintAlgorithm: false,
			GestureSupport:          false,
			PluginSystem:            false,
		},
	}
}

// Load loads configuration from the specified path
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If path is empty, try default locations
	if path == "" {
		path = findConfigFile()
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Parse TOML file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// findConfigFile searches for config file in default locations
func findConfigFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Try ~/.config/govim/config.toml
	configPath := filepath.Join(homeDir, ".config", "govim", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Try ~/Library/Application Support/govim/config.toml
	configPath = filepath.Join(homeDir, "Library", "Application Support", "govim", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	return ""
}

// GetConfigPath returns the expected config file path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Prefer macOS standard location
	return filepath.Join(homeDir, "Library", "Application Support", "govim", "config.toml")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate hint style
	if c.General.HintStyle != "alphabet" && c.General.HintStyle != "numeric" {
		return fmt.Errorf("hint_style must be 'alphabet' or 'numeric'")
	}

	// Validate hint characters
	if c.General.HintStyle == "alphabet" && len(c.General.HintCharacters) < 2 {
		return fmt.Errorf("hint_characters must contain at least 2 characters")
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.LogLevel] {
		return fmt.Errorf("log_level must be one of: debug, info, warn, error")
	}

	// Validate opacity values
	if c.Hints.Opacity < 0 || c.Hints.Opacity > 1 {
		return fmt.Errorf("hints.opacity must be between 0 and 1")
	}
	if c.Appearance.OverlayOpacity < 0 || c.Appearance.OverlayOpacity > 1 {
		return fmt.Errorf("appearance.overlay_opacity must be between 0 and 1")
	}

	// Validate performance settings
	if c.Performance.MaxHintsDisplayed < 1 {
		return fmt.Errorf("performance.max_hints_displayed must be at least 1")
	}
	if c.Performance.MaxConcurrentQueries < 1 {
		return fmt.Errorf("performance.max_concurrent_queries must be at least 1")
	}

	return nil
}

// Save saves the configuration to the specified path
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// Encode to TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
