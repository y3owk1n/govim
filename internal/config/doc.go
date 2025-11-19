// Package config provides comprehensive configuration management for the Neru application.
//
// This package handles all aspects of Neru's configuration system, including loading,
// parsing, validation, and default value management. Configuration is stored in TOML
// format and supports multiple locations with a well-defined precedence order.
//
// Key Features:
//   - Multi-location Support: Configuration can be loaded from multiple standard locations
//   - ~/.config/neru/config.toml (XDG standard - recommended for dotfiles)
//   - ~/Library/Application Support/neru/config.toml (macOS convention)
//   - Custom path via --config flag
//   - TOML Format: Uses TOML for human-readable, structured configuration
//   - Validation: Comprehensive validation of configuration values
//   - Defaults: Sensible defaults for all configuration options
//   - Hot Reloading: Configuration changes require daemon restart to take effect
//
// Configuration Structure:
//   - [general]: Core application settings
//   - excluded_apps: List of bundle IDs to exclude from Neru
//   - accessibility_check_on_start: Whether to check accessibility permissions on startup
//   - restore_cursor_position: Whether to restore cursor position after exiting modes
//   - [hotkeys]: Global hotkey bindings
//   - Key-value pairs mapping hotkeys to actions
//   - [hints]: Hint mode configuration
//   - enabled: Whether hint mode is enabled
//   - hint_characters: Characters to use for hint labels
//   - Visual appearance settings (colors, fonts, etc.)
//   - [hints.additional_ax_support]: Enhanced browser support
//   - [grid]: Grid mode configuration
//   - enabled: Whether grid mode is enabled
//   - characters: Characters to use for grid labels
//   - Visual appearance settings
//   - Behavior settings (subgrid, live matching, etc.)
//   - [scroll]: Scroll mode configuration
//   - Scroll step sizes for different actions
//   - Visual feedback settings
//   - [action]: Action mode configuration
//   - Visual feedback settings
//   - Action key mappings
//   - [smooth_cursor]: Smooth cursor movement settings
//   - move_mouse_enabled: Whether to enable smooth mouse movement
//   - steps: Number of steps for smooth movement
//   - delay: Delay between steps in milliseconds
//   - [logging]: Logging configuration
//   - log_level: Logging verbosity level
//   - log_file: Log file path
//   - structured_logging: Whether to use structured (JSON) logging
//
// The configuration system is designed to be both powerful and user-friendly,
// allowing for extensive customization while providing sensible defaults for
// users who prefer minimal configuration.
package config
