# GoVim

Keyboard-driven navigation for macOS - click and scroll without a mouse.

## Features

- **Hint Mode** - Show labels on clickable elements, type to click
- **Scroll Mode** - Vim-style scrolling (j/k/h/l) anywhere
- **Universal** - Works in browsers, native apps, Electron apps
- **Electron & Chrome Support** - Built-in support for VS Code, Windsurf, Chrome, and more
- **Fast** - Event tap for key capture, CGEvent for scrolling
- **Configurable** - TOML config file

## Quick Start

```bash
# Build
just build

# Run
./bin/govim
```

**Required**: Grant Accessibility permissions in System Settings → Privacy & Security → Accessibility

## Usage

### Hint Mode

1. Press `Cmd+Shift+Space`
2. Type the label shown on the element (e.g., "aa", "ab")
3. Element is clicked automatically
4. Press `Escape` to exit

### Scroll Mode

1. Press `Cmd+Shift+J`
2. Use `j`/`k` to scroll down/up, `h`/`l` for left/right
3. Press `Escape` to exit

### Menu Bar

Click the ⌨️ icon to quit GoVim.

## Configuration

Config file: `~/Library/Application Support/govim/config.toml`

### Example Configuration

```toml
[general]
hint_characters = "asdfghjkl"
hint_style = "alphabet"

[accessibility]
# Global roles used for all apps
clickable_roles = ["AXButton", "AXCheckBox", "AXLink", ...]
scrollable_roles = ["AXScrollArea"]

# Per-app configurations (optional)
# Add additional roles for specific apps
[[accessibility.app_configs]]
bundle_id = "com.apple.Safari"
additional_clickable_roles = ["AXGroup", "AXImage"]
additional_scrollable_roles = ["AXWebArea"]

[accessibility.electron_support]
enable = true
additional_bundles = ["com.google.Chrome"]

[hotkeys]
activate_hint_mode = "Cmd+Shift+Space"
activate_scroll_mode = "Cmd+Shift+J"

[scroll]
scroll_speed = 50
highlight_color = "#FF0000"

[hints]
font_size = 14
background_color = "#FFD700"
text_color = "#000000"
```

See `configs/default-config.toml` for all available options.

### Per-App Role Configuration

GoVim supports global and per-app accessibility role configurations:

- **Global roles** (`clickable_roles` and `scrollable_roles`) are used for all applications
- **Per-app roles** allow you to add additional roles for specific applications
- The final roles used = global roles + app-specific additional roles (merged)

To find an app's bundle ID:
```bash
osascript -e 'id of app "AppName"'
```

Example per-app configuration:
```toml
[[accessibility.app_configs]]
bundle_id = "com.microsoft.VSCode"
additional_clickable_roles = ["AXStaticText"]
additional_scrollable_roles = ["AXGroup"]
```

### Electron & Chrome Support

GoVim includes built-in support for Electron apps (VS Code, Windsurf, Slack, etc.) and Chromium browsers.

- `accessibility.electron_support.enable` toggles automatic enabling of Electron accessibility hooks for legacy apps.
- `accessibility.electron_support.additional_bundles` accepts exact bundle IDs or `prefix*` wildcards for extra Electron apps that require manual accessibility.

## CLI Usage

```bash
# Activate hint mode
govim hint

# Activate scroll mode
govim scroll

# Check status
govim status

# Reload configuration
govim reload-config

# List UI elements (debugging)
govim list-elements --app "Finder"
```

## Architecture

GoVim is built with:
- **Go**: Core application logic
- **CGo/Objective-C**: macOS Accessibility API integration
- **Native macOS APIs**: For overlay rendering and hotkey management

### Project Structure

```
govim/
├── cmd/govim/           # Main entry point
├── internal/
│   ├── accessibility/   # Accessibility API wrappers
│   ├── hints/          # Hint generation and display
│   ├── scroll/         # Scroll mode implementation
│   ├── hotkeys/        # Hotkey management
│   ├── config/         # Configuration parsing
│   ├── cli/            # CLI interface
│   ├── menubar/        # Menu bar integration
│   └── bridge/         # CGo/Objective-C bridges
├── configs/            # Default configuration
└── scripts/            # Build and packaging scripts
```

## Performance

- Hint generation: <100ms for typical windows
- UI element detection: <50ms
- Hotkey response: <10ms
- Idle CPU usage: <1%
- Memory usage: <50MB typical

## Troubleshooting

### Accessibility Permissions

If GoVim isn't working, ensure it has Accessibility permissions:
1. Open System Settings
2. Go to Privacy & Security → Accessibility
3. Enable GoVim

### Configuration Issues

Reload configuration without restarting:
```bash
govim reload-config
```

Check configuration syntax:
```bash
govim validate-config
```

### Logs

Logs are stored at: `~/Library/Logs/govim/app.log`

Enable debug logging in your config for troubleshooting:
```toml
[logging]
log_level = "debug"
```

### Electron Apps (Windsurf, VS Code, etc.)

If hints aren't appearing in Electron app content areas:

1. Ensure Electron support is enabled (default):
   ```toml
   [accessibility.electron_support]
   enable = true
   ```

2. Check logs for: `"App requires Electron support"` and `"Enabled AXManualAccessibility"`

3. Supported apps include: Windsurf, VS Code, Slack, GitHub Desktop, Zoom, Obsidian, Teams

### Too Many Hints / Long Hint Labels

When there are many clickable elements, hints will use 3 characters (e.g., "AAA", "AAS") to avoid prefix conflicts. All hints will have the same length, so you can type the full 3-character code without ambiguity.

To reduce hint length:

1. **Reduce max hints** to stay within 2-character range:
   ```toml
   [performance]
   max_hints_displayed = 80  # With 9 chars, this keeps hints at 2 chars
   ```

2. **Add more hint characters** for more 2-char combinations:
   ```toml
   [general]
   hint_characters = "asdfghjklqwertyuiop"  # 19 chars = 361 two-char combos
   ```

**Hint capacity**: 9 chars = 81 two-char hints, 12 chars = 144 two-char hints, 15 chars = 225 two-char hints.

### Chrome/Chromium Browsers

To enable hints in Chrome:

1. Add to your config:
   ```toml
   [accessibility.electron_support]
   enable = true
   additional_bundles = [
       "com.google.Chrome",
       "com.brave.Browser",
   ]
   ```

2. Reload config: `Cmd+Shift+R`

## Development

### Building

```bash
just build      # Development build
just release    # Optimized release build
just test       # Run tests
just lint       # Run linter
```

### Testing

```bash
just test           # Unit tests
just test-race      # Race detection
```

## License

MIT License - see LICENSE file for details

## Acknowledgments

Inspired by:
- [Homerow](https://www.homerow.app/)
- [Vimac](https://github.com/dexterleng/vimac)
- [Shortcat](https://shortcat.app/)
- [Vimium](https://github.com/philc/vimium)
