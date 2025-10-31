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
additional_clickable_roles = []

[accessibility.electron_support]
enable = true
additional_bundles = ["com.example.legacyapp*"]

[hotkeys]
activate_hint_mode = "Cmd+Shift+Space"
activate_scroll_mode = "Cmd+Shift+J"
exit_mode = "Escape"

[scroll]
scroll_speed = 50
smooth_scroll = true
highlight_color = "#FF0000"

[hints]
font_size = 14
background_color = "#FFD700"
text_color = "#000000"
```

See `configs/default-config.toml` for all available options.

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
