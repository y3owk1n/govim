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
	Logging       LoggingConfig       `toml:"logging"`
}

type GeneralConfig struct {
	ExcludedApps                  []string `toml:"excluded_apps"`
	IncludeMenubarHints           bool     `toml:"include_menubar_hints"`
	AdditionalMenubarHintsTargets []string `toml:"additional_menubar_hints_targets"`
	IncludeDockHints              bool     `toml:"include_dock_hints"`
	IncludeNCHints                bool     `toml:"include_nc_hints"`
	RestorePosAfterLeftClick      bool     `toml:"restore_pos_after_left_click"`
	RestorePosAfterRightClick     bool     `toml:"restore_pos_after_right_click"`
	RestorePosAfterMiddleClick    bool     `toml:"restore_pos_after_middle_click"`
	RestorePosAfterDoubleClick    bool     `toml:"restore_pos_after_double_click"`
}

type AccessibilityConfig struct {
	AccessibilityCheckOnStart bool                `toml:"accessibility_check_on_start"`
	ClickableRoles            []string            `toml:"clickable_roles"`
	ScrollableRoles           []string            `toml:"scrollable_roles"`
	IgnoreClickableCheck      bool                `toml:"ignore_clickable_check"`
	AdditionalAXSupport       AdditionalAXSupport `toml:"additional_ax_support"`
	AppConfigs                []AppConfig         `toml:"app_configs"`
}

type AppConfig struct {
	BundleID             string   `toml:"bundle_id"`
	AdditionalClickable  []string `toml:"additional_clickable_roles"`
	AdditionalScrollable []string `toml:"additional_scrollable_roles"`
	IgnoreClickableCheck bool     `toml:"ignore_clickable_check"`
}

type HotkeysConfig struct {
	// Bindings holds hotkey -> action mappings parsed from the [hotkeys] table.
	// Supported TOML format (preferred):
	// [hotkeys]
	// "Cmd+Shift+Space" = "hints"
	// Values are strings. The special exec prefix is supported: "exec /usr/bin/say hi"
	Bindings map[string]string
}

type ScrollConfig struct {
	ScrollStep          int    `toml:"scroll_step"`
	ScrollStepHalf      int    `toml:"scroll_step_half"`
	ScrollStepFull      int    `toml:"scroll_step_full"`
	HighlightScrollArea bool   `toml:"highlight_scroll_area"`
	HighlightColor      string `toml:"highlight_color"`
	HighlightWidth      int    `toml:"highlight_width"`
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
	ActionBackgroundColor  string `toml:"action_background_color"`
	ActionTextColor        string `toml:"action_text_color"`
	ActionMatchedTextColor string `toml:"action_matched_text_color"`
	ActionBorderColor      string `toml:"action_border_color"`
	// Scroll hints specific colors and opacity
	ScrollHintsBackgroundColor  string `toml:"scroll_hints_background_color"`
	ScrollHintsTextColor        string `toml:"scroll_hints_text_color"`
	ScrollHintsMatchedTextColor string `toml:"scroll_hints_matched_text_color"`
	ScrollHintsBorderColor      string `toml:"scroll_hints_border_color"`
}

type LoggingConfig struct {
	LogLevel          string `toml:"log_level"`
	LogFile           string `toml:"log_file"`
	StructuredLogging bool   `toml:"structured_logging"`
}

type AdditionalAXSupport struct {
	Enable                    bool     `toml:"enable"`
	AdditionalElectronBundles []string `toml:"additional_electron_bundles"`
	AdditionalChromiumBundles []string `toml:"additional_chromium_bundles"`
	AdditionalFirefoxBundles  []string `toml:"additional_firefox_bundles"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			ExcludedApps:        []string{},
			IncludeMenubarHints: false,
			AdditionalMenubarHintsTargets: []string{
				"com.apple.TextInputMenuAgent",
				"com.apple.controlcenter",
				"com.apple.systemuiserver",
			},
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
				"AXWebArea",
				"AXScrollArea",
				"AXTable",
				"AXRow",
				"AXColumn",
				"AXOutline",
				"AXList",
				"AXGroup",
			},
			IgnoreClickableCheck: false,
			AdditionalAXSupport: AdditionalAXSupport{
				Enable:                    false,
				AdditionalElectronBundles: []string{},
				AdditionalChromiumBundles: []string{},
				AdditionalFirefoxBundles:  []string{},
			},
		},
		Hotkeys: HotkeysConfig{
			Bindings: map[string]string{
				"Cmd+Shift+Space": "hints",
				"Cmd+Shift+A":     "hints_action",
				"Cmd+Shift+J":     "scroll",
			},
		},
		Hints: HintsConfig{
			HintCharacters:   "asdfghjkl",
			FontSize:         12,
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
			// Defaults for scroll hints colors to visually differentiate
			ScrollHintsBackgroundColor:  "#2ECC71",
			ScrollHintsTextColor:        "#000000",
			ScrollHintsMatchedTextColor: "#145A32",
			ScrollHintsBorderColor:      "#000000",
		},
		Scroll: ScrollConfig{
			ScrollStep:          50,
			ScrollStepHalf:      500,
			ScrollStepFull:      1000000,
			HighlightScrollArea: true,
			HighlightColor:      "#FF0000",
			HighlightWidth:      2,
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

	// Parse TOML file into the typed config
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Decode the hotkeys table into a generic map and populate cfg.Hotkeys.Bindings.
	// We only support the inline mapping format and reject the old nested `bindings` map.
	var raw map[string]map[string]interface{}
	if _, err := toml.DecodeFile(path, &raw); err == nil {
		if hot, ok := raw["hotkeys"]; ok {
			if _, hasBindings := hot["bindings"]; hasBindings {
				return nil, fmt.Errorf("hotkeys.bindings is not supported; use inline mapping under [hotkeys], e.g. \"Cmd+Shift+Space\" = \"hints\"")
			}
			if cfg.Hotkeys.Bindings == nil {
				cfg.Hotkeys.Bindings = map[string]string{}
			}
			for k, v := range hot {
				// Only accept string values for actions
				str, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("hotkeys.%s must be a string action", k)
				}
				cfg.Hotkeys.Bindings[k] = str
			}
		}
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

	// Try ~/.config/neru/config.toml
	configPath := filepath.Join(homeDir, ".config", "neru", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Try ~/Library/Application Support/neru/config.toml
	configPath = filepath.Join(homeDir, "Library", "Application Support", "neru", "config.toml")
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
	return filepath.Join(homeDir, "Library", "Application Support", "neru", "config.toml")
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
	// Hotkeys are validated in the bindings validation section below

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
	// Validate scroll hints colors
	if err := validateColor(c.Hints.ScrollHintsBackgroundColor, "hints.scroll_hints_background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHintsTextColor, "hints.scroll_hints_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHintsMatchedTextColor, "hints.scroll_hints_matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHintsBorderColor, "hints.scroll_hints_border_color"); err != nil {
		return err
	}

	// Validate scroll settings
	if c.Scroll.ScrollStep < 1 {
		return fmt.Errorf("scroll.scroll_speed must be at least 1")
	}
	if c.Scroll.ScrollStepHalf < 1 {
		return fmt.Errorf("scroll.half_page_multiplier must be at least 1")
	}
	if c.Scroll.ScrollStepFull < 1 {
		return fmt.Errorf("scroll.full_page_multiplier must be at least 1")
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

	for _, bundle := range c.Accessibility.AdditionalAXSupport.AdditionalElectronBundles {
		if strings.TrimSpace(bundle) == "" {
			return fmt.Errorf("accessibility.electron_support.additional_electron_bundles cannot contain empty values")
		}
	}

	for _, bundle := range c.Accessibility.AdditionalAXSupport.AdditionalChromiumBundles {
		if strings.TrimSpace(bundle) == "" {
			return fmt.Errorf("accessibility.electron_support.additional_chromium_bundles cannot contain empty values")
		}
	}

	for _, bundle := range c.Accessibility.AdditionalAXSupport.AdditionalFirefoxBundles {
		if strings.TrimSpace(bundle) == "" {
			return fmt.Errorf("accessibility.electron_support.additional_firefox_bundles cannot contain empty values")
		}
	}

	// Validate app configs
	for i, appConfig := range c.Accessibility.AppConfigs {
		if strings.TrimSpace(appConfig.BundleID) == "" {
			return fmt.Errorf("accessibility.app_configs[%d].bundle_id cannot be empty", i)
		}

		// Validate hotkey bindings
		for k, v := range c.Hotkeys.Bindings {
			if strings.TrimSpace(k) == "" {
				return fmt.Errorf("hotkeys.bindings contains an empty key")
			}
			if err := validateHotkey(k, "hotkeys.bindings"); err != nil {
				return err
			}
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("hotkeys.bindings[%s] cannot be empty", k)
			}
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
