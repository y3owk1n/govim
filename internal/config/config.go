package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the complete application configuration
type Config struct {
	General       GeneralConfig       `toml:"general"`
	Accessibility AccessibilityConfig `toml:"accessibility"`
	Hotkeys       HotkeysConfig       `toml:"hotkeys"`
	Scroll        ScrollConfig        `toml:"scroll"`
	Hints         HintsConfig         `toml:"hints"`
	Performance   PerformanceConfig   `toml:"performance"`
	Logging       LoggingConfig       `toml:"logging"`
}

type GeneralConfig struct {
	HintCharacters            string `toml:"hint_characters"`
	HintStyle                 string `toml:"hint_style"`
	AccessibilityCheckOnStart bool   `toml:"accessibility_check_on_start"`
}

type AccessibilityConfig struct {
	ClickableRoles  []string              `toml:"clickable_roles"`
	ScrollableRoles []string              `toml:"scrollable_roles"`
	ElectronSupport ElectronSupportConfig `toml:"electron_support"`
}

type HotkeysConfig struct {
	ActivateHintMode            string `toml:"activate_hint_mode"`
	ActivateHintModeWithActions string `toml:"activate_hint_mode_with_actions"`
	ActivateScrollMode          string `toml:"activate_scroll_mode"`
	ReloadConfig                string `toml:"reload_config"`
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
	FontSize         int     `toml:"font_size"`
	FontFamily       string  `toml:"font_family"`
	BackgroundColor  string  `toml:"background_color"`
	TextColor        string  `toml:"text_color"`
	MatchedTextColor string  `toml:"matched_text_color"`
	BorderRadius     int     `toml:"border_radius"`
	Padding          int     `toml:"padding"`
	BorderWidth      int     `toml:"border_width"`
	BorderColor      string  `toml:"border_color"`
	Opacity          float64 `toml:"opacity"`
	Menubar          bool    `toml:"menubar"`
	Dock             bool    `toml:"dock"`
}

type PerformanceConfig struct {
	MaxHintsDisplayed    int `toml:"max_hints_displayed"`
	DebounceMs           int `toml:"debounce_ms"`
	CacheDurationMs      int `toml:"cache_duration_ms"`
	MaxConcurrentQueries int `toml:"max_concurrent_queries"`
}

type LoggingConfig struct {
	LogLevel          string `toml:"log_level"`
	LogFile           string `toml:"log_file"`
	StructuredLogging bool   `toml:"structured_logging"`
}

type ElectronSupportConfig struct {
	Enable            bool     `toml:"enable"`
	AdditionalBundles []string `toml:"additional_bundles"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			HintCharacters:            "asdfghjkl",
			HintStyle:                 "alphabet",
			AccessibilityCheckOnStart: true,
		},
		Accessibility: AccessibilityConfig{
			ClickableRoles: []string{
				"AXButton",
				"AXCheckBox",
				"AXRadioButton",
				"AXPopUpButton",
				"AXMenuItem",
				"AXMenuBarItem",
				"AXDockItem",
				"AXApplicationDockItem",
				"AXLink",
				"AXTextField",
				"AXTextArea",
			},
			ScrollableRoles: []string{
				"AXScrollArea",
			},
			ElectronSupport: ElectronSupportConfig{
				Enable:            true,
				AdditionalBundles: []string{},
			},
		},
		Hotkeys: HotkeysConfig{
			ActivateHintMode:            "Cmd+Shift+Space",
			ActivateHintModeWithActions: "Cmd+Shift+A",
			ActivateScrollMode:          "Cmd+Shift+J",
			ReloadConfig:                "Cmd+Shift+R",
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
			FontSize:         14,
			FontFamily:       "SF Mono",
			BackgroundColor:  "#FFD700",
			TextColor:        "#000000",
			MatchedTextColor: "#0066CC",
			BorderRadius:     4,
			Padding:          4,
			BorderWidth:      1,
			BorderColor:      "#000000",
			Opacity:          0.95,
			Menubar:          false,
			Dock:             false,
		},
		Performance: PerformanceConfig{
			MaxHintsDisplayed:    200,
			DebounceMs:           50,
			CacheDurationMs:      100,
			MaxConcurrentQueries: 10,
		},
		Logging: LoggingConfig{
			LogLevel:          "info",
			LogFile:           "",
			StructuredLogging: true,
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

	// Validate performance settings
	if c.Performance.MaxHintsDisplayed < 1 {
		return fmt.Errorf("performance.max_hints_displayed must be at least 1")
	}
	if c.Performance.MaxConcurrentQueries < 1 {
		return fmt.Errorf("performance.max_concurrent_queries must be at least 1")
	}

	for _, role := range c.Accessibility.ClickableRoles {
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("accessibility.clickable_roles cannot contain empty values")
		}
	}

	for _, role := range c.Accessibility.ScrollableRoles {
		if strings.TrimSpace(role) == "" {
			return fmt.Errorf("accessibility.scrollable_roles cannot contain empty values")
		}
	}

	for _, bundle := range c.Accessibility.ElectronSupport.AdditionalBundles {
		if strings.TrimSpace(bundle) == "" {
			return fmt.Errorf("accessibility.electron_support.additional_bundles cannot contain empty values")
		}
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
