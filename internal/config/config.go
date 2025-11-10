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
	Grid          GridConfig          `toml:"grid"`
	Logging       LoggingConfig       `toml:"logging"`
}

type GeneralConfig struct {
	ExcludedApps                  []string `toml:"excluded_apps"`
	IncludeMenubarHints           bool     `toml:"include_menubar_hints"`
	AdditionalMenubarHintsTargets []string `toml:"additional_menubar_hints_targets"`
	IncludeDockHints              bool     `toml:"include_dock_hints"`
	IncludeNCHints                bool     `toml:"include_nc_hints"`
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
	HintCharacters   string                             `toml:"hint_characters"`
	FontSize         int                                `toml:"font_size"`
	FontFamily       string                             `toml:"font_family"`
	BorderRadius     int                                `toml:"border_radius"`
	Padding          int                                `toml:"padding"`
	BorderWidth      int                                `toml:"border_width"`
	Opacity          float64                            `toml:"opacity"`
	Enabled          bool                               `toml:"enabled"`
	LeftClickHints   HintsActionConfigWithRestoreCursor `toml:"left_click_hints"`
	RightClickHints  HintsActionConfigWithRestoreCursor `toml:"right_click_hints"`
	DoubleClickHints HintsActionConfigWithRestoreCursor `toml:"double_click_hints"`
	TripleClickHints HintsActionConfigWithRestoreCursor `toml:"triple_click_hints"`
	MiddleClickHints HintsActionConfigWithRestoreCursor `toml:"middle_click_hints"`
	MouseUpHints     HintsActionConfig                  `toml:"mouse_up_hints"`
	MouseDownHints   HintsActionConfig                  `toml:"mouse_down_hints"`
	MoveMouseHints   HintsActionConfig                  `toml:"move_mouse_hints"`
	ScrollHints      HintsActionConfig                  `toml:"scroll_hints"`
	ContextMenuHints HintsActionConfig                  `toml:"context_menu_hints"`
}

type HintsActionConfig struct {
	BackgroundColor  string `toml:"background_color"`
	TextColor        string `toml:"text_color"`
	MatchedTextColor string `toml:"matched_text_color"`
	BorderColor      string `toml:"border_color"`
}

type HintsActionConfigWithRestoreCursor struct {
	HintsActionConfig
	RestoreCursor bool `toml:"restore_cursor"`
}

type GridActionConfig struct {
	RestoreCursor bool `toml:"restore_cursor"`
}

type LoggingConfig struct {
	LogLevel          string `toml:"log_level"`
	LogFile           string `toml:"log_file"`
	StructuredLogging bool   `toml:"structured_logging"`
}

type GridConfig struct {
	Characters             string           `toml:"characters"`
	SublayerKeys           string           `toml:"sublayer_keys"`
	MinCellSize            int              `toml:"min_cell_size"`
	MaxCellSize            int              `toml:"max_cell_size"`
	FontSize               int              `toml:"font_size"`
	FontFamily             string           `toml:"font_family"`
	Opacity                float64          `toml:"opacity"`
	BackgroundColor        string           `toml:"background_color"`
	TextColor              string           `toml:"text_color"`
	MatchedTextColor       string           `toml:"matched_text_color"`
	MatchedBackgroundColor string           `toml:"matched_background_color"`
	MatchedBorderColor     string           `toml:"matched_border_color"`
	BorderColor            string           `toml:"border_color"`
	BorderWidth            int              `toml:"border_width"`
	LiveMatchUpdate        bool             `toml:"live_match_update"`
	SubgridEnabled         bool             `toml:"subgrid_enabled"`
	SubgridRows            int              `toml:"subgrid_rows"`
	SubgridCols            int              `toml:"subgrid_cols"`
	Enabled                bool             `toml:"enabled"`
	LeftClick              GridActionConfig `toml:"left_click"`
	RightClick             GridActionConfig `toml:"right_click"`
	DoubleClick            GridActionConfig `toml:"double_click"`
	TripleClick            GridActionConfig `toml:"triple_click"`
	MiddleClick            GridActionConfig `toml:"middle_click"`
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
			IncludeDockHints: false,
			IncludeNCHints:   false,
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
				"Cmd+Shift+Space": "hints left_click",
				"Cmd+Shift+A":     "hints context_menu",
				"Cmd+Shift+J":     "hints scroll",
			},
		},
		Hints: HintsConfig{
			HintCharacters: "asdfghjkl",
			FontSize:       12,
			FontFamily:     "SF Mono",
			BorderRadius:   4,
			Padding:        4,
			BorderWidth:    1,
			Opacity:        0.95,
			Enabled:        true,
			LeftClickHints: HintsActionConfigWithRestoreCursor{
				HintsActionConfig: HintsActionConfig{
					BackgroundColor:  "#FFD700",
					TextColor:        "#000000",
					MatchedTextColor: "#737373",
					BorderColor:      "#000000",
				},
				RestoreCursor: false,
			},
			RightClickHints: HintsActionConfigWithRestoreCursor{
				HintsActionConfig: HintsActionConfig{
					BackgroundColor:  "#FFD700",
					TextColor:        "#000000",
					MatchedTextColor: "#737373",
					BorderColor:      "#000000",
				},
				RestoreCursor: false,
			},
			DoubleClickHints: HintsActionConfigWithRestoreCursor{
				HintsActionConfig: HintsActionConfig{
					BackgroundColor:  "#FFD700",
					TextColor:        "#000000",
					MatchedTextColor: "#737373",
					BorderColor:      "#000000",
				},
				RestoreCursor: false,
			},
			TripleClickHints: HintsActionConfigWithRestoreCursor{
				HintsActionConfig: HintsActionConfig{
					BackgroundColor:  "#FFD700",
					TextColor:        "#000000",
					MatchedTextColor: "#737373",
					BorderColor:      "#000000",
				},
				RestoreCursor: false,
			},
			MiddleClickHints: HintsActionConfigWithRestoreCursor{
				HintsActionConfig: HintsActionConfig{
					BackgroundColor:  "#FFD700",
					TextColor:        "#000000",
					MatchedTextColor: "#737373",
					BorderColor:      "#000000",
				},
				RestoreCursor: false,
			},
			MouseUpHints: HintsActionConfig{
				BackgroundColor:  "#FFD700",
				TextColor:        "#000000",
				MatchedTextColor: "#737373",
				BorderColor:      "#000000",
			},
			MouseDownHints: HintsActionConfig{
				BackgroundColor:  "#FFD700",
				TextColor:        "#000000",
				MatchedTextColor: "#737373",
				BorderColor:      "#000000",
			},
			MoveMouseHints: HintsActionConfig{
				BackgroundColor:  "#FFD700",
				TextColor:        "#000000",
				MatchedTextColor: "#737373",
				BorderColor:      "#000000",
			},
			ScrollHints: HintsActionConfig{
				BackgroundColor:  "#2ECC71",
				TextColor:        "#000000",
				MatchedTextColor: "#145A32",
				BorderColor:      "#000000",
			},
			ContextMenuHints: HintsActionConfig{
				BackgroundColor:  "#66CCFF",
				TextColor:        "#000000",
				MatchedTextColor: "#005585",
				BorderColor:      "#000000",
			},
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
		Grid: GridConfig{
			Characters:             "asdfghjkl",
			SublayerKeys:           "",
			MinCellSize:            40,
			MaxCellSize:            200,
			FontSize:               12,
			FontFamily:             "SF Mono",
			Opacity:                0.85,
			BackgroundColor:        "#abe9b3",
			TextColor:              "#ffffff",
			MatchedTextColor:       "#ffffff",
			MatchedBackgroundColor: "#f8bd96",
			MatchedBorderColor:     "#f8bd96",
			BorderColor:            "#abe9b3",
			BorderWidth:            1,
			LiveMatchUpdate:        true,
			SubgridEnabled:         true,
			SubgridRows:            3,
			SubgridCols:            3,
			Enabled:                true,
			LeftClick:              GridActionConfig{RestoreCursor: false},
			RightClick:             GridActionConfig{RestoreCursor: false},
			DoubleClick:            GridActionConfig{RestoreCursor: false},
			TripleClick:            GridActionConfig{RestoreCursor: false},
			MiddleClick:            GridActionConfig{RestoreCursor: false},
		},
	}
}

// Load loads configuration from the specified path
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If path is empty, try default locations
	if path == "" {
		path = FindConfigFile()
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
func FindConfigFile() string {
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

// Validate validates the configuration
func (c *Config) Validate() error {
	// At least one mode must be enabled
	if !c.Hints.Enabled && !c.Grid.Enabled {
		return fmt.Errorf("at least one mode must be enabled: hints.enabled or grid.enabled")
	}
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
	if err := validateColor(c.Hints.LeftClickHints.BackgroundColor, "hints.left_click_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.LeftClickHints.TextColor, "hints.left_click_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.LeftClickHints.MatchedTextColor, "hints.left_click_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.LeftClickHints.BorderColor, "hints.left_click_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.RightClickHints.BackgroundColor, "hints.right_click_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.RightClickHints.TextColor, "hints.right_click_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.RightClickHints.MatchedTextColor, "hints.right_click_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.RightClickHints.BorderColor, "hints.right_click_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.DoubleClickHints.BackgroundColor, "hints.double_click_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.DoubleClickHints.TextColor, "hints.double_click_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.DoubleClickHints.MatchedTextColor, "hints.double_click_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.DoubleClickHints.BorderColor, "hints.double_click_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.TripleClickHints.BackgroundColor, "hints.triple_click_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.TripleClickHints.TextColor, "hints.triple_click_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.TripleClickHints.MatchedTextColor, "hints.triple_click_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.TripleClickHints.BorderColor, "hints.triple_click_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.MiddleClickHints.BackgroundColor, "hints.middle_click_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MiddleClickHints.TextColor, "hints.middle_click_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MiddleClickHints.MatchedTextColor, "hints.middle_click_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MiddleClickHints.BorderColor, "hints.middle_click_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.MouseUpHints.BackgroundColor, "hints.mouse_up_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseUpHints.TextColor, "hints.mouse_up_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseUpHints.MatchedTextColor, "hints.mouse_up_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseUpHints.BorderColor, "hints.mouse_up_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.MouseDownHints.BackgroundColor, "hints.mouse_down_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseDownHints.TextColor, "hints.mouse_down_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseDownHints.MatchedTextColor, "hints.mouse_down_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MouseDownHints.BorderColor, "hints.mouse_down_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.MoveMouseHints.BackgroundColor, "hints.move_mouse_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MoveMouseHints.TextColor, "hints.move_mouse_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MoveMouseHints.MatchedTextColor, "hints.move_mouse_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.MoveMouseHints.BorderColor, "hints.move_mouse_hints.border_color"); err != nil {
		return err
	}

	if err := validateColor(c.Hints.ScrollHints.BackgroundColor, "hints.scroll_hints.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHints.TextColor, "hints.scroll_hints.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHints.MatchedTextColor, "hints.scroll_hints.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Hints.ScrollHints.BorderColor, "hints.scroll_hints.border_color"); err != nil {
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

	// Validate grid settings
	if strings.TrimSpace(c.Grid.Characters) == "" {
		return fmt.Errorf("grid.characters cannot be empty")
	}
	if len(c.Grid.Characters) < 2 {
		return fmt.Errorf("grid.characters must contain at least 2 characters")
	}
	if c.Grid.MinCellSize < 1 {
		return fmt.Errorf("grid.min_cell_size must be at least 1")
	}
	if c.Grid.MaxCellSize > 0 && c.Grid.MaxCellSize < c.Grid.MinCellSize {
		return fmt.Errorf("grid.max_cell_size must be greater than or equal to min_cell_size")
	}
	if c.Grid.FontSize < 6 || c.Grid.FontSize > 72 {
		return fmt.Errorf("grid.font_size must be between 6 and 72")
	}
	if c.Grid.BorderWidth < 0 {
		return fmt.Errorf("grid.border_width must be non-negative")
	}
	if c.Grid.Opacity < 0 || c.Grid.Opacity > 1 {
		return fmt.Errorf("grid.opacity must be between 0 and 1")
	}
	if err := validateColor(c.Grid.BackgroundColor, "grid.background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Grid.TextColor, "grid.text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Grid.MatchedTextColor, "grid.matched_text_color"); err != nil {
		return err
	}
	if err := validateColor(c.Grid.BorderColor, "grid.border_color"); err != nil {
		return err
	}
	if err := validateColor(c.Grid.MatchedBackgroundColor, "grid.matched_background_color"); err != nil {
		return err
	}
	if err := validateColor(c.Grid.MatchedBorderColor, "grid.matched_border_color"); err != nil {
		return err
	}
	if c.Grid.SubgridEnabled {
		if c.Grid.SubgridRows < 1 || c.Grid.SubgridCols < 1 {
			return fmt.Errorf("grid.subgrid_rows and grid.subgrid_cols must be at least 1")
		}
		// Validate sublayer keys length (fallback to grid.characters) for rows*cols
		keys := strings.TrimSpace(c.Grid.SublayerKeys)
		if keys == "" {
			keys = c.Grid.Characters
		}
		required := c.Grid.SubgridRows * c.Grid.SubgridCols
		if len([]rune(keys)) < required {
			return fmt.Errorf("grid.sublayer_keys must contain at least %d characters (rows*cols) for subgrid selection", required)
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
