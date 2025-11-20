package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"go.uber.org/zap"
)

// ActionConfig defines the visual and behavioral settings for action mode.
type ActionConfig struct {
	HighlightColor string `toml:"highlight_color"`
	HighlightWidth int    `toml:"highlight_width"`

	LeftClickKey   string `toml:"left_click_key"`
	RightClickKey  string `toml:"right_click_key"`
	MiddleClickKey string `toml:"middle_click_key"`
	MouseDownKey   string `toml:"mouse_down_key"`
	MouseUpKey     string `toml:"mouse_up_key"`
}

// Config represents the complete application configuration structure.
type Config struct {
	General      GeneralConfig      `toml:"general"`
	Hotkeys      HotkeysConfig      `toml:"hotkeys"`
	Hints        HintsConfig        `toml:"hints"`
	Grid         GridConfig         `toml:"grid"`
	Scroll       ScrollConfig       `toml:"scroll"`
	Action       ActionConfig       `toml:"action"`
	Logging      LoggingConfig      `toml:"logging"`
	SmoothCursor SmoothCursorConfig `toml:"smooth_cursor"`
}

// GeneralConfig defines general application-wide settings.
type GeneralConfig struct {
	ExcludedApps              []string `toml:"excluded_apps"`
	AccessibilityCheckOnStart bool     `toml:"accessibility_check_on_start"`
	RestoreCursorPosition     bool     `toml:"restore_cursor_position"`
}

// AppConfig defines application-specific settings for role customization.
type AppConfig struct {
	BundleID             string   `toml:"bundle_id"`
	AdditionalClickable  []string `toml:"additional_clickable_roles"`
	IgnoreClickableCheck bool     `toml:"ignore_clickable_check"`
}

// HotkeysConfig defines hotkey mappings and their associated actions.
type HotkeysConfig struct {
	// Bindings holds hotkey -> action mappings parsed from the [hotkeys] table.
	// Supported TOML format (preferred):
	// [hotkeys]
	// "Cmd+Shift+Space" = "hints"
	// Values are strings. The special exec prefix is supported: "exec /usr/bin/say hi"
	Bindings map[string]string `toml:"bindings"`
}

// ScrollConfig defines the behavior and appearance settings for scroll mode.
type ScrollConfig struct {
	ScrollStep          int    `toml:"scroll_step"`
	ScrollStepHalf      int    `toml:"scroll_step_half"`
	ScrollStepFull      int    `toml:"scroll_step_full"`
	HighlightScrollArea bool   `toml:"highlight_scroll_area"`
	HighlightColor      string `toml:"highlight_color"`
	HighlightWidth      int    `toml:"highlight_width"`
}

// HintsConfig defines the visual and behavioral settings for hints mode.
type HintsConfig struct {
	Enabled        bool    `toml:"enabled"`
	HintCharacters string  `toml:"hint_characters"`
	FontSize       int     `toml:"font_size"`
	FontFamily     string  `toml:"font_family"`
	BorderRadius   int     `toml:"border_radius"`
	Padding        int     `toml:"padding"`
	BorderWidth    int     `toml:"border_width"`
	Opacity        float64 `toml:"opacity"`

	BackgroundColor  string `toml:"background_color"`
	TextColor        string `toml:"text_color"`
	MatchedTextColor string `toml:"matched_text_color"`
	BorderColor      string `toml:"border_color"`

	IncludeMenubarHints           bool     `toml:"include_menubar_hints"`
	AdditionalMenubarHintsTargets []string `toml:"additional_menubar_hints_targets"`
	IncludeDockHints              bool     `toml:"include_dock_hints"`
	IncludeNCHints                bool     `toml:"include_nc_hints"`

	ClickableRoles       []string `toml:"clickable_roles"`
	IgnoreClickableCheck bool     `toml:"ignore_clickable_check"`

	AppConfigs []AppConfig `toml:"app_configs"`

	AdditionalAXSupport AdditionalAXSupport `toml:"additional_ax_support"`
}

// GridConfig defines the visual and behavioral settings for grid mode.
type GridConfig struct {
	Enabled bool `toml:"enabled"`

	Characters   string `toml:"characters"`
	SublayerKeys string `toml:"sublayer_keys"`

	FontSize    int     `toml:"font_size"`
	FontFamily  string  `toml:"font_family"`
	Opacity     float64 `toml:"opacity"`
	BorderWidth int     `toml:"border_width"`

	BackgroundColor        string `toml:"background_color"`
	TextColor              string `toml:"text_color"`
	MatchedTextColor       string `toml:"matched_text_color"`
	MatchedBackgroundColor string `toml:"matched_background_color"`
	MatchedBorderColor     string `toml:"matched_border_color"`
	BorderColor            string `toml:"border_color"`

	LiveMatchUpdate bool `toml:"live_match_update"`
	HideUnmatched   bool `toml:"hide_unmatched"`
}

// LoggingConfig defines the logging behavior and file management settings.
type LoggingConfig struct {
	LogLevel          string `toml:"log_level"`
	LogFile           string `toml:"log_file"`
	StructuredLogging bool   `toml:"structured_logging"`

	// New options for log rotation and file logging control
	DisableFileLogging bool `toml:"disable_file_logging"`
	MaxFileSize        int  `toml:"max_file_size"` // Size in MB
	MaxBackups         int  `toml:"max_backups"`   // Maximum number of old log files to retain
	MaxAge             int  `toml:"max_age"`       // Maximum number of days to retain old log files
}

// SmoothCursorConfig defines the smooth cursor movement settings.
type SmoothCursorConfig struct {
	MoveMouseEnabled bool `toml:"move_mouse_enabled"`
	Steps            int  `toml:"steps"`
	Delay            int  `toml:"delay"` // Delay in milliseconds
}

// AdditionalAXSupport defines accessibility support for specific application frameworks.
type AdditionalAXSupport struct {
	Enable                    bool     `toml:"enable"`
	AdditionalElectronBundles []string `toml:"additional_electron_bundles"`
	AdditionalChromiumBundles []string `toml:"additional_chromium_bundles"`
	AdditionalFirefoxBundles  []string `toml:"additional_firefox_bundles"`
}

// LoadResult contains the result of loading a configuration file.
type LoadResult struct {
	Config          *Config
	ValidationError error
	ConfigPath      string
}

// LoadWithValidation loads configuration from the specified path and returns both
// the config and any validation error separately. This allows callers to decide
// how to handle validation failures (e.g., show alert and use default config).
func LoadWithValidation(path string) *LoadResult {
	result := &LoadResult{
		Config:     DefaultConfig(),
		ConfigPath: path,
	}

	if path == "" {
		result.ConfigPath = FindConfigFile()
	}

	logger.Info("Loading config from", zap.String("path", result.ConfigPath))

	_, err := os.Stat(result.ConfigPath)
	if os.IsNotExist(err) {
		logger.Info("Config file not found, using default configuration")
		return result
	}

	_, err = toml.DecodeFile(result.ConfigPath, result.Config)
	if err != nil {
		result.ValidationError = fmt.Errorf("failed to parse config file: %w", err)
		result.Config = DefaultConfig()
		return result
	}

	var raw map[string]map[string]any
	_, err = toml.DecodeFile(result.ConfigPath, &raw)
	if err == nil {
		if hot, ok := raw["hotkeys"]; ok {
			if len(hot) > 0 {
				// Clear default bindings when user provides hotkeys config
				result.Config.Hotkeys.Bindings = map[string]string{}
			}
			for key, value := range hot {
				str, ok := value.(string)
				if !ok {
					result.ValidationError = fmt.Errorf("hotkeys.%s must be a string action", key)
					result.Config = DefaultConfig()
					return result
				}
				result.Config.Hotkeys.Bindings[key] = str
			}
		}
	}

	err = result.Config.Validate()
	if err != nil {
		result.ValidationError = fmt.Errorf("invalid configuration: %w", err)
		result.Config = DefaultConfig()
		return result
	}

	logger.Info("Configuration loaded successfully")
	return result
}

// DefaultConfig returns the default application configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			ExcludedApps:              []string{},
			AccessibilityCheckOnStart: true,
			RestoreCursorPosition:     false,
		},
		Hotkeys: HotkeysConfig{
			Bindings: map[string]string{
				"Cmd+Shift+Space": "hints",
				"Cmd+Shift+G":     "grid",
				"Cmd+Shift+S":     "action scroll",
			},
		},
		Hints: HintsConfig{
			Enabled:        true,
			HintCharacters: "asdfghjkl",
			FontSize:       12,
			FontFamily:     "SF Mono",
			BorderRadius:   4,
			Padding:        4,
			BorderWidth:    1,
			Opacity:        0.95,

			BackgroundColor:  "#FFD700",
			TextColor:        "#000000",
			MatchedTextColor: "#737373",
			BorderColor:      "#000000",

			IncludeMenubarHints: false,
			AdditionalMenubarHintsTargets: []string{
				"com.apple.TextInputMenuAgent",
				"com.apple.controlcenter",
				"com.apple.systemuiserver",
			},
			IncludeDockHints: false,
			IncludeNCHints:   false,

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
			IgnoreClickableCheck: false,

			AppConfigs: []AppConfig{},

			AdditionalAXSupport: AdditionalAXSupport{
				Enable:                    false,
				AdditionalElectronBundles: []string{},
				AdditionalChromiumBundles: []string{},
				AdditionalFirefoxBundles:  []string{},
			},
		},
		Grid: GridConfig{
			Enabled: true,

			Characters:   "abcdefghijklmnpqrstuvwxyz",
			SublayerKeys: "abcdefghijklmnpqrstuvwxyz",

			FontSize:    12,
			FontFamily:  "SF Mono",
			Opacity:     0.7,
			BorderWidth: 1,

			BackgroundColor:        "#abe9b3",
			TextColor:              "#000000",
			MatchedTextColor:       "#f8bd96",
			MatchedBackgroundColor: "#f8bd96",
			MatchedBorderColor:     "#f8bd96",
			BorderColor:            "#abe9b3",

			LiveMatchUpdate: true,
			HideUnmatched:   true,
		},
		Scroll: ScrollConfig{
			ScrollStep:          50,
			ScrollStepHalf:      500,
			ScrollStepFull:      1000000,
			HighlightScrollArea: true,
			HighlightColor:      "#FF0000",
			HighlightWidth:      2,
		},
		Action: ActionConfig{
			HighlightColor: "#00FF00",
			HighlightWidth: 3,

			// Default action key mappings
			LeftClickKey:   "l",
			RightClickKey:  "r",
			MiddleClickKey: "m",
			MouseDownKey:   "i",
			MouseUpKey:     "u",
		},
		Logging: LoggingConfig{
			LogLevel:           "info",
			LogFile:            "",
			StructuredLogging:  true,
			DisableFileLogging: false,
			MaxFileSize:        10, // 10MB
			MaxBackups:         5,  // Keep 5 old log files
			MaxAge:             30, // Keep log files for 30 days
		},
		SmoothCursor: SmoothCursorConfig{
			MoveMouseEnabled: false,
			Steps:            10,
			Delay:            1, // 1ms delay between steps
		},
	}
}

// Load loads configuration from the specified path.
// For backward compatibility, this returns an error if validation fails.
// Use LoadWithValidation for graceful error handling.
func Load(path string) (*Config, error) {
	result := LoadWithValidation(path)
	if result.ValidationError != nil {
		return nil, result.ValidationError
	}
	return result.Config, nil
}

// FindConfigFile searches for config file in default locations.
func FindConfigFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Try ~/.config/neru/config.toml
	configPath := filepath.Join(homeDir, ".config", "neru", "config.toml")
	_, err = os.Stat(configPath)
	if err == nil {
		logger.Info("Found config at", zap.String("path", configPath))
		return configPath
	}

	// Try ~/Library/Application Support/neru/config.toml
	configPath = filepath.Join(homeDir, "Library", "Application Support", "neru", "config.toml")
	_, err = os.Stat(configPath)
	if err == nil {
		logger.Info("Found config at", zap.String("path", configPath))
		return configPath
	}

	logger.Info("No config file found in default locations")
	return ""
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// At least one mode must be enabled
	if !c.Hints.Enabled && !c.Grid.Enabled {
		return errors.New("at least one mode must be enabled: hints.enabled or grid.enabled")
	}

	// Validate hints configuration
	err := c.validateHints()
	if err != nil {
		return err
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.LogLevel] {
		return errors.New("log_level must be one of: debug, info, warn, error")
	}

	// Validate scroll settings
	if c.Scroll.ScrollStep < 1 {
		return errors.New("scroll.scroll_speed must be at least 1")
	}
	if c.Scroll.ScrollStepHalf < 1 {
		return errors.New("scroll.half_page_multiplier must be at least 1")
	}
	if c.Scroll.ScrollStepFull < 1 {
		return errors.New("scroll.full_page_multiplier must be at least 1")
	}

	// Validate app configs
	err = c.validateAppConfigs()
	if err != nil {
		return err
	}

	// Validate grid settings
	err = c.validateGrid()
	if err != nil {
		return err
	}

	// Validate action settings
	err = c.validateAction()
	if err != nil {
		return err
	}

	// Validate smooth cursor settings
	err = c.validateSmoothCursor()
	if err != nil {
		return err
	}

	return nil
}

// Save saves the configuration to the specified path.
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	var err error
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0o750)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file
	var closeErr error
	// #nosec G304 -- Path is validated and controlled by the application
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		cerr := file.Close()
		if cerr != nil && closeErr == nil {
			closeErr = fmt.Errorf("failed to close config file: %w", cerr)
		}
	}()

	// Encode to TOML
	encoder := toml.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return closeErr
}

// IsAppExcluded checks if the given bundle ID is in the excluded apps list.
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

// GetClickableRolesForApp returns the merged clickable roles for a specific app.
func (c *Config) GetClickableRolesForApp(bundleID string) []string {
	rolesMap := make(map[string]struct{})
	for _, role := range c.Hints.ClickableRoles {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			rolesMap[trimmed] = struct{}{}
		}
	}

	for _, appConfig := range c.Hints.AppConfigs {
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

	if c.Hints.IncludeMenubarHints {
		rolesMap["AXMenuBarItem"] = struct{}{}
	}

	if c.Hints.IncludeDockHints {
		rolesMap["AXDockItem"] = struct{}{}
	}

	roles := make([]string, 0, len(rolesMap))
	for role := range rolesMap {
		roles = append(roles, role)
	}
	return roles
}
