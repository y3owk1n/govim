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
- [Hint Appearance](#hint-appearance)
- [Per-Action Hint Customization](#per-action-hint-customization)
- [Scroll Configuration](#scroll-configuration)
- [Accessibility](#accessibility)
- [Logging](#logging)
- [Complete Example](#complete-example)

---

## Hotkeys

Bind global hotkeys to Neru actions. Set to `""` or comment out to disable.

```toml
[hotkeys]
# Hint modes
"Cmd+Shift+Space" = "hints left_click"
"Cmd+Shift+A" = "hints context_menu"  # Action selection mode
"Cmd+Shift+J" = "hints scroll"
"Cmd+Shift+R" = "hints right_click"
"Cmd+Shift+D" = "hints double_click"
"Cmd+Shift+T" = "hints triple_click"
"Cmd+Shift+H" = "hints middle_click"
"Cmd+Shift+M" = "hints move_mouse"

# Grid mode
"Cmd+Shift+Space" = "grid left_click"
"Cmd+Shift+A" = "grid context_menu"  # Action selection mode
"Cmd+Shift+J" = "grid scroll"
"Cmd+Shift+R" = "grid right_click"
"Cmd+Shift+D" = "grid double_click"
"Cmd+Shift+T" = "grid triple_click"
"Cmd+Shift+H" = "grid middle_click"
"Cmd+Shift+M" = "grid move_mouse"

# Execute shell commands
"Cmd+Alt+T" = "exec open -a Terminal"
"Cmd+Alt+N" = "exec osascript -e 'display notification \"Hello!\" with title \"Neru\"'"
```

### Hotkey Syntax

**Modifiers:**

- `Cmd` - Command (⌘)
- `Ctrl` - Control
- `Alt` / `Option` - Option (⌥)
- `Shift` - Shift (⇧)

**Format:** `"Modifier1+Modifier2+Key" = "action"`

**Examples:**

```toml
"Ctrl+F" = "hints left_click"
"Cmd+Ctrl+Space" = "hints context_menu"
"Alt+J" = "hints scroll"
```

### Shell Commands

Execute shell commands from hotkeys using `exec` prefix:

```toml
# Open applications
"Cmd+Alt+T" = "exec open -a Terminal"
"Cmd+Alt+C" = "exec open -a 'Visual Studio Code'"

# Run scripts
"Cmd+Alt+S" = "exec ~/scripts/screenshot.sh"

# AppleScript
"Cmd+Alt+N" = "exec osascript -e 'display notification \"Task complete!\"'"
```

**Important:** Escape quotes properly in shell commands.

### Alternative: Use skhd

Instead of Neru hotkeys, use [skhd](https://github.com/koekeishiya/skhd) or similar:

```bash
# ~/.config/skhd/skhdrc
ctrl - f : neru hints left_click
ctrl - g : neru hints context_menu
ctrl - s : neru hints scroll
```

This approach separates hotkey management from Neru.

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

# The following are all hints based configuration, ignore them if you're just using grid mode

# Show hints in menubar
include_menubar_hints = true

# Additional menubar apps to target (requires include_menubar_hints = true)
additional_menubar_hints_targets = [
    "com.apple.TextInputMenuAgent",    # Input menu
    "com.apple.controlcenter",          # Control Center
    "com.apple.systemuiserver",         # System UI (Siri, etc.)
    "net.kovidgoyal.kitty",             # Example: Kitty terminal
]

# Show hints in Dock (also includes Mission Control)
include_dock_hints = true

# Show hints in notification popups
include_nc_hints = true
```

### Finding Bundle IDs

```bash
# Get bundle ID for running app
osascript -e 'id of app "AppName"'

# Examples
osascript -e 'id of app "Safari"'       # com.apple.Safari
osascript -e 'id of app "VS Code"'      # com.microsoft.VSCode
```

---

## Hint Appearance

Configure the visual appearance of hints globally.

```toml
[hints]
# Characters used for hint labels (at least 2 distinct characters)
hint_characters = "asdfghjkl"

# Font settings
font_size = 12              # Range: 6-72
font_family = ""            # Empty string = system default

# Visual styling
border_radius = 4           # Pixels
padding = 4                 # Pixels
border_width = 1            # Pixels
opacity = 0.95              # Range: 0.0-1.0 (1.0 = fully opaque)
```

### Hint Characters

Choose characters that:

- Are easy to type on your keyboard layout
- Are visually distinct
- You can touch-type comfortably

**Examples:**

```toml
# Home row (recommended)
hint_characters = "asdfghjkl"

# QWERTY left hand
hint_characters = "asdfqwertzxcv"

# Custom
hint_characters = "fjdksla"
```

---

## Per-Action Hint Customization

Customize colors and behavior for each action type.

```toml
[hints.left_click_hints]
background_color = "#FFD700"    # Gold
text_color = "#000000"          # Black
matched_text_color = "#737373"  # Gray (typed characters)
border_color = "#000000"        # Black
restore_cursor = false          # Restore cursor to original position after click

[hints.right_click_hints]
background_color = "#FF6B6B"    # Red
text_color = "#FFFFFF"
matched_text_color = "#CCCCCC"
border_color = "#CC0000"
restore_cursor = false

[hints.double_click_hints]
background_color = "#4ECDC4"    # Teal
text_color = "#000000"
matched_text_color = "#555555"
border_color = "#2BA39C"
restore_cursor = true

[hints.triple_click_hints]
background_color = "#95E1D3"    # Mint
text_color = "#000000"
matched_text_color = "#555555"
border_color = "#6DCFC2"
restore_cursor = false

[hints.middle_click_hints]
background_color = "#FFA07A"    # Light salmon
text_color = "#000000"
matched_text_color = "#666666"
border_color = "#FF7F50"
restore_cursor = false

[hints.mouse_up_hints]
background_color = "#DDA0DD"    # Plum
text_color = "#000000"
matched_text_color = "#888888"
border_color = "#BA55D3"

[hints.mouse_down_hints]
background_color = "#F0E68C"    # Khaki
text_color = "#000000"
matched_text_color = "#777777"
border_color = "#DAA520"

[hints.move_mouse_hints]
background_color = "#87CEEB"    # Sky blue
text_color = "#000000"
matched_text_color = "#666666"
border_color = "#4682B4"

[hints.scroll_hints]
background_color = "#98FB98"    # Pale green
text_color = "#000000"
matched_text_color = "#555555"
border_color = "#00FA9A"

[hints.context_menu_hints]
background_color = "#DEB887"    # Burlywood
text_color = "#000000"
matched_text_color = "#666666"
border_color = "#D2691E"
```

### Cursor Restoration

When `restore_cursor = true`, the cursor returns to its original position after the action completes. Useful for:

- Quick actions where you want to maintain context
- Workflows where cursor position matters
- Preventing accidental cursor drift

---

## Scroll Configuration

Configure Vim-style scrolling behavior.

```toml
[scroll]
# Base scroll amount (j/k keys)
scroll_step = 50

# Half-page scroll (Ctrl+D/U)
scroll_step_half = 500

# Full-page scroll (gg/G)
# Use a very large number for "scroll to top/bottom"
scroll_step_full = 1000000

# Visual feedback
highlight_scroll_area = true    # Show border around active scroll area
highlight_color = "#FF0000"     # Border color
highlight_width = 2             # Border width in pixels
```

### Scroll Keys

Once in scroll mode:

- `j` / `k` - Scroll down/up by `scroll_step`
- `h` / `l` - Scroll left/right by `scroll_step`
- `Ctrl+d` / `Ctrl+u` - Half-page down/up
- `gg` - Jump to top
- `G` - Jump to bottom
- `Esc` - Exit scroll mode
- `Tab` - Toggle between direct scroll and hint/grid selection

---

## Accessibility

Configure which UI elements Neru can interact with.

### Global Clickable Roles

> [!NOTE]
> This is only used for hints mode, no effect on grid mode.

```toml
[accessibility]
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

# Skip role checking (all elements become clickable)
# ⚠️ Enable only if you know what you're doing
ignore_clickable_check = false
```

### Global Scrollable Roles

> [!NOTE]
> This is only used for hints mode, no effect on grid mode.

```toml
[accessibility]
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
```

### Per-App Customization

> [!NOTE]
> This is only used for hints mode, no effect on grid mode.

Override or extend accessibility settings for specific apps:

```toml
[[accessibility.app_configs]]
bundle_id = "com.google.Chrome"
additional_clickable_roles = ["AXTabGroup"]  # Click Chrome tab groups
additional_scrollable_roles = []
ignore_clickable_check = false  # Override global setting

[[accessibility.app_configs]]
bundle_id = "com.adobe.illustrator"
additional_clickable_roles = ["AXStaticText", "AXImage"]
ignore_clickable_check = true  # Make everything clickable
```

**How it works:**

- Global roles apply to all apps
- Per-app `additional_*_roles` are **merged** with global roles
- Per-app `ignore_clickable_check` overrides the global setting

### Additional AX Support (Electron, Chromium, Firefox)

> [!NOTE]
> This is only used for hints mode, no effect on grid mode.

Enable enhanced accessibility for apps that need it:

```toml
[accessibility.additional_ax_support]
enable = false  # Off by default

# Automatically supported apps (no need to add):
# Electron: VS Code, Windsurf, Cursor, Slack, Spotify, Obsidian
# Chromium: Chrome, Brave, Arc, Helium
# Firefox: Firefox, Zen

# Add custom Electron apps
additional_electron_bundles = [
    "com.example.electronapp",
]

# Add custom Chromium apps
additional_chromium_bundles = [
    "com.example.custombrowser",
]

# Add custom Firefox apps
additional_firefox_bundles = [
    "com.example.firefoxfork",
]
```

**When to enable:**

- You need to use hints mode
- You are using electron, chromium, or firefox

**⚠️ Tiling Window Manager Warning:**

Enabling accessibility support for Chromium and Firefox (`AXEnhancedUserInterface`) can cause side effects with tiling window managers like:

- **yabai** - Windows may not tile correctly
- **Amethyst** - Layout issues
- **Arospace** - Weird animation when activating affected apps
- Other tiling WMs

**Symptoms:**

- Browser windows resist tiling
- Windows snap to wrong positions
- Tiling rules ignored
- Layout glitches

**Recommendation:**
If you use a tiling window manager, keep `enable = false` and rely on Neru's grid-based approach instead. The grid method works well in browsers without accessibility modifications.

**Note:** This is a grid-based workaround for apps with poor accessibility tree support.

---

## Logging

Configure logging verbosity for debugging.

```toml
[logging]
# Log levels: "debug", "info", "warn", "error"
log_level = "info"
```

**Log location:** `~/Library/Logs/neru/app.log`

**Enable debug logging:**

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

A full configuration example combining common settings:

```toml
# ~/.config/neru/config.toml

# Hotkeys
[hotkeys]
"Ctrl+F" = "hints left_click"
"Ctrl+G" = "hints context_menu"
"Ctrl+S" = "hints scroll"
"Ctrl+R" = "hints right_click"
"Ctrl+D" = "hints double_click"
"Ctrl+Alt+T" = "exec open -a Terminal"

# General
[general]
excluded_apps = [
    "com.apple.Terminal",
    "com.googlecode.iterm2",
]
include_menubar_hints = true
additional_menubar_hints_targets = ["net.kovidgoyal.kitty"]
include_dock_hints = true
include_nc_hints = true

# Hint appearance
[hints]
hint_characters = "asdfghjkl"
font_size = 14
border_radius = 6
padding = 5
opacity = 0.9

# Left click hints
[hints.left_click_hints]
background_color = "#FFD700"
text_color = "#000000"
matched_text_color = "#737373"
border_color = "#000000"
restore_cursor = true

# Scroll configuration
[scroll]
scroll_step = 60
scroll_step_half = 400
highlight_scroll_area = true
highlight_color = "#00FF00"
highlight_width = 3

# Accessibility
[accessibility]
clickable_roles = [
    "AXButton",
    "AXLink",
    "AXTextField",
    "AXCheckBox",
]

[[accessibility.app_configs]]
bundle_id = "com.google.Chrome"
additional_clickable_roles = ["AXTabGroup"]

# Electron/Chromium support
[accessibility.additional_ax_support]
enable = true

# Logging
[logging]
log_level = "info"
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
