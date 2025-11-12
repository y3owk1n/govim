# Configuration Guide

Neru uses TOML for configuration. This guide covers all available options with examples.

## Configuration File Locations

Neru searches for configuration in the following order:

1. `~/.config/neru/config.toml` (XDG standard - **recommended for dotfiles**)
2. `~/Library/Application Support/neru/config.toml` (macOS convention)
3. Custom path: `neru launch --config /path/to/config.toml`

**No config file?** Neru uses sensible defaults. See [../configs/default-config.toml](../configs/default-config.toml).

---

## Table of Contents

- [Hotkeys](#hotkeys)
- [General Settings](#general-settings)
- [Hint Mode](#hint-mode)
- [Grid Mode](#grid-mode)
- [Scroll Configuration](#scroll-configuration)
- [Logging](#logging)
- [Complete Example](#complete-example)

---

## Hotkeys

Bind global hotkeys to Neru actions. Comment out to disable.

```toml
[hotkeys]
# Hint modes
"Cmd+Shift+Space" = "hints left_click"
"Cmd+Shift+A" = "hints context_menu"
"Cmd+Shift+J" = "action scroll"

# Grid mode
"Cmd+Shift+G" = "grid left_click"

# Additional actions (uncomment to enable)
# "Ctrl+R" = "hints right_click"
# "Ctrl+D" = "hints double_click"
# "Cmd+Shift+M" = "grid move_mouse"
# "Cmd+Shift+S" = "grid scroll"

# Execute shell commands
# "Cmd+Alt+T" = "exec open -a Terminal"
# "Cmd+Alt+N" = "exec osascript -e 'display notification \"Hello!\" with title \"Neru\"'"
```

### Hotkey Syntax

**Modifiers:** `Cmd`, `Ctrl`, `Alt`/`Option`, `Shift`
**Format:** `"Modifier1+Modifier2+Key" = "action"`

**Shell Commands:**

```toml
"Cmd+Alt+T" = "exec open -a Terminal"
"Cmd+Alt+C" = "exec open -a 'Visual Studio Code'"
"Cmd+Alt+S" = "exec ~/scripts/screenshot.sh"
```

**Alternative:** Use [skhd](https://github.com/koekeishiya/skhd) for hotkey management:

```bash
# ~/.config/skhd/skhdrc
ctrl - f : neru hints left_click
ctrl - g : neru hints context_menu
```

---

## General Settings

```toml
[general]
# Exclude apps by bundle ID
excluded_apps = [
    "com.apple.Terminal",
    "com.googlecode.iterm2",
    "com.microsoft.rdc.macos",
]
accessibility_check_on_start = true
```

**Finding Bundle IDs:**

```bash
osascript -e 'id of app "Safari"'  # com.apple.Safari
osascript -e 'id of app "VS Code"' # com.microsoft.VSCode
```

---

## Hint Mode

Hint mode overlays clickable labels on UI elements using macOS accessibility APIs.

### Basic Configuration

```toml
[hints]
enabled = true
hint_characters = "asdfghjkl"  # At least 2 distinct characters

# Visual styling
font_size = 12                 # Range: 6-72
font_family = ""               # Empty = system default
border_radius = 4
padding = 4
border_width = 1
opacity = 0.95                 # Range: 0.0-1.0
```

**Choosing hint characters:**

- Use home row keys for comfort: `"asdfghjkl"`
- Left hand only: `"asdfqwertzxcv"`
- Custom: `"fjdksla"`

### Hint Visibility Options

```toml
[hints]
# Show hints in menubar
include_menubar_hints = false

# Target specific menubar apps (requires include_menubar_hints = true)
additional_menubar_hints_targets = [
    "com.apple.TextInputMenuAgent",
    "com.apple.controlcenter",
    "com.apple.systemuiserver",
]

# Show hints in Dock and Mission Control
include_dock_hints = false

# Show hints in notification popups
include_nc_hints = false
```

### Per-Action Customization

Customize colors and behavior for each action type:

```toml
[hints.left_click_hints]
background_color = "#FFD700"    # Gold
text_color = "#000000"
matched_text_color = "#737373"  # Typed characters
border_color = "#000000"
restore_cursor = false          # Return cursor to original position

[hints.right_click_hints]
background_color = "#FF6B6B"
text_color = "#FFFFFF"
matched_text_color = "#CCCCCC"
border_color = "#CC0000"
restore_cursor = false

[hints.double_click_hints]
background_color = "#4ECDC4"
text_color = "#000000"
matched_text_color = "#555555"
border_color = "#2BA39C"
restore_cursor = true

# Additional action types: triple_click, middle_click, mouse_up,
# mouse_down, move_mouse, scroll, context_menu
```

**Cursor restoration:** When `restore_cursor = true`, the cursor returns to its original position after the action completes.

### Accessibility Configuration

Define which UI elements are clickable or scrollable:

```toml
[hints]
# Global clickable roles
clickable_roles = [
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
]

# Global scrollable roles
scrollable_roles = [
    "AXWebArea",
    "AXScrollArea",
    "AXTable",
    "AXRow",
    "AXColumn",
    "AXOutline",
    "AXList",
    "AXGroup",
]

# ⚠️ Make all elements clickable (use with caution)
ignore_clickable_check = false
```

### Per-App Overrides

Customize accessibility for specific apps:

```toml
[[hints.app_configs]]
bundle_id = "com.google.Chrome"
additional_clickable_roles = ["AXTabGroup"]
additional_scrollable_roles = []
ignore_clickable_check = false

[[hints.app_configs]]
bundle_id = "com.adobe.illustrator"
additional_clickable_roles = ["AXStaticText", "AXImage"]
ignore_clickable_check = true
```

**How it works:**

- App-specific `additional_*_roles` are **merged** with global roles
- App-specific `ignore_clickable_check` overrides the global setting

### Enhanced Browser Support

Enable improved accessibility for Electron, Chromium, and Firefox apps:

```toml
[hints.additional_ax_support]
enable = false  # Off by default

# Automatically supported (no need to add):
# Electron: VS Code, Windsurf, Cursor, Slack, Spotify, Obsidian
# Chromium: Chrome, Brave, Arc, Helium
# Firefox: Firefox, Zen

# Add custom apps
additional_electron_bundles = ["com.example.electronapp"]
additional_chromium_bundles = ["com.example.custombrowser"]
additional_firefox_bundles = ["com.example.firefoxfork"]
```

**⚠️ Tiling Window Manager Warning:**

Enabling accessibility support for Chromium/Firefox can interfere with tiling window managers (yabai, Amethyst, Aerospace). Symptoms include:

- Windows resist tiling
- Layout glitches
- Windows snap to wrong positions

**Recommendation:** If you use a tiling WM, keep `enable = false` and use grid mode instead.

---

## Grid Mode

Grid mode provides a universal, accessibility-independent way to click anywhere on screen using coordinate-based selection.

### Basic Configuration

```toml
[grid]
enabled = true
characters = "abcdefghijklmnpqrstuvwxyz"

# Visual styling
font_size = 12
font_family = ""
opacity = 0.7
border_width = 1
```

**Cell sizing:** Automatically optimized based on screen resolution. Uses 2-4 character labels with square cells.

### Grid Behavior

```toml
[grid]
live_match_update = true    # Highlight matches as you type
subgrid_enabled = true      # Show 3x3 precision layer after selection
hide_unmatched = true       # Hide non-matching cells while typing

# Subgrid keys (requires at least 9 characters)
sublayer_keys = "abcdefghijklmnpqrstuvwxyz"
```

**Workflow:**

1. Press grid hotkey (e.g., `Cmd+Shift+G`)
2. Type main grid coordinate (2-4 characters)
3. If subgrid enabled, type subgrid position (1 character, a-i)
4. Action executes at selected location

### Per-Action Configuration

```toml
[grid.left_click]
restore_cursor = false
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.right_click]
restore_cursor = false
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.double_click]
restore_cursor = false
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.triple_click]
restore_cursor = false
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.middle_click]
restore_cursor = false
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.mouse_up]
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.mouse_down]
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.move_mouse]
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.scroll]
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"

[grid.context_menu]
background_color = "#abe9b3"
text_color = "#000000"
matched_text_color = "#000000"
matched_background_color = "#f8bd96"
matched_border_color = "#f8bd96"
border_color = "#abe9b3"
```

---

## Scroll Configuration

Configure Vim-style scrolling behavior for both standalone scrolling and hint/grid-based scrolling.

```toml
[scroll]
# Scroll amounts
scroll_step = 50              # j/k keys
scroll_step_half = 500        # Ctrl+D/U
scroll_step_full = 1000000    # gg/G (top/bottom)

# Visual feedback
highlight_scroll_area = true
highlight_color = "#FF0000"
highlight_width = 2
```

### Scroll Keys

- `j` / `k` - Scroll down/up
- `h` / `l` - Scroll left/right
- `Ctrl+d` / `Ctrl+u` - Half-page down/up
- `gg` - Jump to top (press `g` twice)
- `G` - Jump to bottom
- `Tab` - Return to hint/grid overlay (hints/grid mode only)
- `Esc` - Exit scroll mode

### Scroll Modes

**Standalone scroll** (`neru action scroll`):

- Scrolls at current cursor position
- No location selection required
- Only `Esc` to exit

**Hint/Grid scroll** (`neru hints scroll` or `neru grid scroll`):

- Select location first, then scroll
- Press `Tab` to reselect location
- Press `Esc` to exit

---

## Logging

```toml
[logging]
log_level = "info"          # "debug", "info", "warn", "error"
log_file = ""               # Empty = ~/Library/Logs/neru/app.log
structured_logging = true   # JSON format for file logs
```

**Default location:** `~/Library/Logs/neru/app.log`

**Debug logging:**

```toml
[logging]
log_level = "debug"
```

Then check logs:

```bash
tail -f ~/Library/Logs/neru/app.log
```

---

## Complete Example

A full configuration example:

```toml
# ~/.config/neru/config.toml

[hotkeys]
"Cmd+Shift+Space" = "hints left_click"
"Cmd+Shift+A" = "hints context_menu"
"Cmd+Shift+J" = "action scroll"
"Cmd+Shift+G" = "grid left_click"

[general]
excluded_apps = ["com.apple.Terminal", "com.googlecode.iterm2"]
accessibility_check_on_start = true

[hints]
enabled = true
hint_characters = "asdfghjkl"
font_size = 14
border_radius = 6
padding = 5
opacity = 0.9
include_menubar_hints = false
include_dock_hints = false
include_nc_hints = false
clickable_roles = ["AXButton", "AXLink", "AXTextField", "AXCheckBox"]
scrollable_roles = ["AXWebArea", "AXScrollArea", "AXTable"]
ignore_clickable_check = false

[[hints.app_configs]]
bundle_id = "com.google.Chrome"
additional_clickable_roles = ["AXTabGroup"]

[hints.additional_ax_support]
enable = false

[hints.left_click_hints]
background_color = "#FFD700"
text_color = "#000000"
matched_text_color = "#737373"
border_color = "#000000"
restore_cursor = false

[grid]
enabled = true
characters = "abcdefghijklmnpqrstuvwxyz"
font_size = 12
opacity = 0.7
live_match_update = true
subgrid_enabled = true
hide_unmatched = true

[grid.left_click]
restore_cursor = false

[scroll]
scroll_step = 60
scroll_step_half = 400
scroll_step_full = 1000000
highlight_scroll_area = true
highlight_color = "#00FF00"
highlight_width = 3

[logging]
log_level = "info"
structured_logging = true
```

---

## Tips

**Version control your config:**

```bash
cd ~/.config/neru
git init
git add config.toml
git commit -m "Initial Neru configuration"
```

**Share with others:**

```bash
# Export
cp ~/.config/neru/config.toml ~/Downloads/neru-config.toml

# Import
cp ~/Downloads/neru-config.toml ~/.config/neru/config.toml
neru launch  # Restart to apply
```

**Test changes:**

1. Edit `~/.config/neru/config.toml`
2. Restart: `pkill neru && neru launch`
3. Test your changes

---

## Next Steps

- See [CLI.md](CLI.md) for command-line usage
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) if configs aren't working
- Review [default-config.toml](../configs/default-config.toml) for all options
