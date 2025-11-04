package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	ExcludedApps               []string `toml:"excluded_apps"`
	IncludeMenubarHints        bool     `toml:"include_menubar_hints"`
	IncludeDockHints           bool     `toml:"include_dock_hints"`
	IncludeNCHints             bool     `toml:"include_nc_hints"`
	RestorePosAfterLeftClick   bool     `toml:"restore_pos_after_left_click"`
	RestorePosAfterRightClick  bool     `toml:"restore_pos_after_right_click"`
	RestorePosAfterMiddleClick bool     `toml:"restore_pos_after_middle_click"`
	RestorePosAfterDoubleClick bool     `toml:"restore_pos_after_double_click"`
}

type AccessibilityConfig struct {
	AccessibilityCheckOnStart bool                  `toml:"accessibility_check_on_start"`
	ClickableRoles            []string              `toml:"clickable_roles"`
	ScrollableRoles           []string              `toml:"scrollable_roles"`
	ElectronSupport           ElectronSupportConfig `toml:"electron_support"`
	AppConfigs                []AppConfig           `toml:"app_configs"`
}

type AppConfig struct {
	BundleID             string   `toml:"bundle_id"`
	AdditionalClickable  []string `toml:"additional_clickable_roles"`
	AdditionalScrollable []string `toml:"additional_scrollable_roles"`
}

type HotkeysConfig struct {
	ActivateHintMode            string `toml:"activate_hint_mode"`
	ActivateHintModeWithActions string `toml:"activate_hint_mode_with_actions"`
	ActivateScrollMode          string `toml:"activate_scroll_mode"`
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
	HintCharacters   string  `toml:"hint_characters"`
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
	// Action overlay specific colors and opacity
	ActionBackgroundColor  string  `toml:"action_background_color"`
	ActionTextColor        string  `toml:"action_text_color"`
	ActionMatchedTextColor string  `toml:"action_matched_text_color"`
	ActionBorderColor      string  `toml:"action_border_color"`
	ActionOpacity          float64 `toml:"action_opacity"`
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
			ExcludedApps:               []string{},
			IncludeMenubarHints:        false,
			IncludeDockHints:           false,
			IncludeNCHints:             false,
			RestorePosAfterLeftClick:   false,
			RestorePosAfterRightClick:  false,
			RestorePosAfterMiddleClick: false,
			RestorePosAfterDoubleClick: false,
		},
		Accessibility: AccessibilityConfig{
			AccessibilityCheckOnStart: true,
			ClickableRoles: []string{
				"AXButton",
				"AXComboBox",
				"AXCheckBox",
				"AXRadioButton",
				"AXLink",
				"AXPopUpButton",
				"AXTextField",
				"AXSlider",
				"AXTabButton",
				"AXSwitch",
				"AXDisclosureTriangle",
				"AXTextArea",
				"AXMenuButton",
				"AXMenuItem",
				"AXCell",
				"AXRow",
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
			ActivateHintMode:            "",
			ActivateHintModeWithActions: "",
			ActivateScrollMode:          "",
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
			HintCharacters:   "asdfghjkl",
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
			// Defaults for action overlay colors to visually differentiate
			ActionBackgroundColor:  "#66CCFF",
			ActionTextColor:        "#000000",
			ActionMatchedTextColor: "#003366",
			ActionBorderColor:      "#000000",
			ActionOpacity:          0.95,
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
	// Validate hint characters
	if strings.TrimSpace(c.Hints.HintCharacters) == "" {
		return fmt.Errorf("hint_characters cannot be empty")
	}
	if len(c.Hints.HintCharacters) < 2 {
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
	if c.Hints.ActionOpacity < 0 || c.Hints.ActionOpacity > 1 {
		return fmt.Errorf("hints.action_opacity must be between 0 and 1")
	}

	// Validate performance settings
	if c.Performance.MaxHintsDisplayed < 1 {
		return fmt.Errorf("performance.max_hints_displayed must be at least 1")
	}
	if c.Performance.MaxConcurrentQueries < 1 {
		return fmt.Errorf("performance.max_concurrent_queries must be at least 1")
	}

	// Validate hotkeys
	if err := validateHotkey(c.Hotkeys.ActivateHintMode, "hotkeys.activate_hint_mode"); err != nil {
		return err
	}
	if err := validateHotkey(c.Hotkeys.ActivateHintModeWithActions, "hotkeys.activate_hint_mode_with_actions"); err != nil {
		return err
	}
	if err := validateHotkey(c.Hotkeys.ActivateScrollMode, "hotkeys.activate_scroll_mode"); err != nil {
		return err
	}

	// Validate colors
	if err := validateColor(c.Hints.BackgroundColor, "hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.TextColor, "hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MatchedTextColor, "hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.BorderColor, "hints.border_color"); err != nil {
		return err
	}
	// Validate action overlay colors
	if err := validateColor(c.Hints.ActionBackgroundColor, "hints.action_background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ActionTextColor, "hints.action_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ActionMatchedTextColor, "hints.action_matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ActionBorderColor, "hints.action_border_color"); err != nil {
		return err
	}
	if err := validateColor(c.Scroll.HighlightColor, "scroll.highlight_color"); err != nil {
		return err
	}

	// Validate scroll settings
	if c.Scroll.ScrollSpeed < 1 {
		return fmt.Errorf("scroll.scroll_speed must be at least 1")
	}
	if c.Scroll.PageHeight < 100 {
		return fmt.Errorf("scroll.page_height must be at least 100")
	}
	if c.Scroll.HalfPageMultiplier <= 0 || c.Scroll.HalfPageMultiplier > 1 {
		return fmt.Errorf("scroll.half_page_multiplier must be between 0 and 1")
	}
	if c.Scroll.FullPageMultiplier <= 0 || c.Scroll.FullPageMultiplier > 1 {
		return fmt.Errorf("scroll.full_page_multiplier must be between 0 and 1")
	}

	// Validate hints settings
	if c.Hints.FontSize < 6 || c.Hints.FontSize > 72 {
		return fmt.Errorf("hints.font_size must be between 6 and 72")
	}
	if c.Hints.BorderRadius < 0 {
		return fmt.Errorf("hints.border_radius must be non-negative")
	}
	if c.Hints.Padding < 0 {
		return fmt.Errorf("hints.padding must be non-negative")
	}
	if c.Hints.BorderWidth < 0 {
		return fmt.Errorf("hints.border_width must be non-negative")
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

	// Validate app configs
	for i, appConfig := range c.Accessibility.AppConfigs {
		if strings.TrimSpace(appConfig.BundleID) == "" {
			return fmt.Errorf("accessibility.app_configs[%d].bundle_id cannot be empty", i)
		}
		for _, role := range appConfig.AdditionalClickable {
			if strings.TrimSpace(role) == "" {
				return fmt.Errorf("accessibility.app_configs[%d].additional_clickable_roles cannot contain empty values", i)
			}
		}
		for _, role := range appConfig.AdditionalScrollable {
			if strings.TrimSpace(role) == "" {
				return fmt.Errorf("accessibility.app_configs[%d].additional_scrollable_roles cannot contain empty values", i)
			}
		}
	}

	return nil
}

// GetClickableRolesForApp returns the merged clickable roles for a specific app.
// It combines global clickable roles with app-specific additional roles.
func (c *Config) GetClickableRolesForApp(bundleID string) []string {
	// Start with global roles
	rolesMap := make(map[string]struct{})
	for _, role := range c.Accessibility.ClickableRoles {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			rolesMap[trimmed] = struct{}{}
		}
	}

	// Add app-specific roles
	for _, appConfig := range c.Accessibility.AppConfigs {
		if appConfig.BundleID == bundleID {
			for _, role := range appConfig.AdditionalClickable {
				trimmed := strings.TrimSpace(role)
				if trimmed != "" {
					rolesMap[trimmed] = struct{}{}
				}
			}
			break
		}
	}

	// Add menubar roles if enabled
	if c.General.IncludeMenubarHints {
		rolesMap["AXMenuBarItem"] = struct{}{}
	}

	// Add dock roles if enabled
	if c.General.IncludeDockHints {
		rolesMap["AXDockItem"] = struct{}{}
	}

	// Convert map to slice
	roles := make([]string, 0, len(rolesMap))
	for role := range rolesMap {
		roles = append(roles, role)
	}
	return roles
}

// GetScrollableRolesForApp returns the merged scrollable roles for a specific app.
// It combines global scrollable roles with app-specific additional roles.
func (c *Config) GetScrollableRolesForApp(bundleID string) []string {
	// Start with global roles
	rolesMap := make(map[string]struct{})
	for _, role := range c.Accessibility.ScrollableRoles {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			rolesMap[trimmed] = struct{}{}
		}
	}

	// Add app-specific roles
	for _, appConfig := range c.Accessibility.AppConfigs {
		if appConfig.BundleID == bundleID {
			for _, role := range appConfig.AdditionalScrollable {
				trimmed := strings.TrimSpace(role)
				if trimmed != "" {
					rolesMap[trimmed] = struct{}{}
				}
			}
			break
		}
	}

	// Convert map to slice
	roles := make([]string, 0, len(rolesMap))
	for role := range rolesMap {
		roles = append(roles, role)
	}
	return roles
}

// IsAppExcluded checks if the given bundle ID is in the excluded apps list
func (c *Config) IsAppExcluded(bundleID string) bool {
	if bundleID == "" {
		return false
	}

	// Normalize bundle ID for case-insensitive comparison
	bundleID = strings.ToLower(strings.TrimSpace(bundleID))

	for _, excludedApp := range c.General.ExcludedApps {
		excludedApp = strings.ToLower(strings.TrimSpace(excludedApp))
		if excludedApp == bundleID {
			return true
		}
	}
	return false
}

// Save saves the configuration to the specified path
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	var closeErr error
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && closeErr == nil {
			closeErr = fmt.Errorf("failed to close config file: %w", cerr)
		}
	}()

	// Encode to TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return closeErr
}

// validateHotkey validates a hotkey string format
func validateHotkey(hotkey, fieldName string) error {
	if strings.TrimSpace(hotkey) == "" {
		return nil // Allow empty hotkey to disable the action
	}

	// Hotkey format: [Modifier+]*Key
	// Valid modifiers: Cmd, Ctrl, Alt, Shift, Option
	// Examples: "Cmd+Shift+Space", "Ctrl+D", "F1"

	parts := strings.Split(hotkey, "+")
	if len(parts) == 0 {
		return fmt.Errorf("%s has invalid format: %s", fieldName, hotkey)
	}

	validModifiers := map[string]bool{
		"Cmd":    true,
		"Ctrl":   true,
		"Alt":    true,
		"Shift":  true,
		"Option": true,
	}

	// Check all parts except the last (which is the key)
	for i := 0; i < len(parts)-1; i++ {
		modifier := strings.TrimSpace(parts[i])
		if !validModifiers[modifier] {
			return fmt.Errorf("%s has invalid modifier '%s' in: %s (valid: Cmd, Ctrl, Alt, Shift, Option)",
				fieldName, modifier, hotkey)
		}
	}

	// Last part should be the key (non-empty)
	key := strings.TrimSpace(parts[len(parts)-1])
	if key == "" {
		return fmt.Errorf("%s has empty key in: %s", fieldName, hotkey)
	}

	return nil
}

// validateColor validates a color string (hex format)
func validateColor(color, fieldName string) error {
	if strings.TrimSpace(color) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Match hex color format: #RGB, #RRGGBB, #RRGGBBAA
	hexColorRegex := regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$`)

	if !hexColorRegex.MatchString(color) {
		return fmt.Errorf("%s has invalid hex color format: %s (expected #RGB, #RRGGBB, or #RRGGBBAA)",
			fieldName, color)
	}

	return nil
}
