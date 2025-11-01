# GoVim 🚀

<div align="center">

[![License](https://img.shields.io/github/license/y3owk1n/govim)](LICENSE)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
![Go Version](https://img.shields.io/github/go-mod/go-version/y3owk1n/govim)
[![Latest Release](https://img.shields.io/github/v/release/y3owk1n/govim)](https://github.com/y3owk1n/govim/releases)

**Keyboard-driven navigation for macOS - Navigate and click without touching your mouse 🖱️**

[Installation](#installation) •
[Features](#features) •
[Usage](#usage) •
[Configuration](#configuration)

</div>

## ✨ Features

- 🎯 **Hint Mode** - Click any UI element using keyboard shortcuts
- 📜 **Scroll Mode** - Vim-style scrolling in any application
- 🌐 **Universal** - Works with native macOS apps and Electron apps
- ⚡ **Performance** - Built with native macOS APIs for instant response
- 🛠️ **Customizable** - Configure hints, hotkeys, and behaviors via TOML
- 🎨 **Minimal UI** - Non-intrusive hints that don't get in your way

## 🚀 Installation

### Homebrew (Recommended)

```bash
brew tap y3owk1n/tap
brew install y3owk1n/tap/govim
```

### Manual Build

```bash
# Clone the repository
git clone https://github.com/y3owk1n/govim.git
cd govim

# Build
just build

# Run
./bin/govim
```

### Required Permissions

⚠️ Grant Accessibility permissions in:
System Settings → Privacy & Security → Accessibility

## 🎮 Usage

### Hint Mode

1. Press `Cmd+Shift+Space` (default)
2. Clickable elements show hint labels
3. Type the label to click (e.g., "aa", "ab")
4. Press `Esc` to exit

### Scroll Mode

1. Press `Cmd+Shift+J` (default)
2. Use Vim-style navigation:
   - `j` / `k` - Scroll down/up
   - `h` / `l` - Scroll left/right
   - `c-d` / `c-u` - Page down/up
   - `gg` / `G` - Top/bottom
3. Press `Esc` to exit

### Menu Bar

The ⌨️ icon in your menu bar provides:

- Quit option

## ⚙️ Configuration

Config file location can be any of the following:

- Macos Convention: `~/Library/Application Support/govim/config.toml`
- XDG: `~/.config/govim/config.toml`

or any other location specified via the `--config` flag.

### Example Configuration

```toml
[general]
hint_characters = "asdfghjkl"  # Characters used for hints
hint_style = "alphabet"        # "alphabet" or "numeric"

[accessibility]
# Clickable elements
clickable_roles = [
    "AXButton",
    "AXCheckBox",
    "AXMenuItem",
    "AXRadioButton",
    "AXLink"
]
scrollable_roles = ["AXScrollArea"]

[hotkeys]
activate_hint_mode = "Cmd+Shift+Space"
activate_scroll_mode = "Cmd+Shift+J"

[hints]
font_size = 14
background_color = "#FFD700"
text_color = "#000000"
opacity = 0.9

[scroll]
scroll_speed = 50
```

See [`configs/default-config.toml`](configs/default-config.toml) for all available options.

### Application Support

GoVim works with:

- Native macOS applications
- Electron-based apps (VS Code, Chrome, etc.)
- Web browsers
- System UI elements

For app-specific settings, find the bundle ID using:

```bash
osascript -e 'id of app "App Name"'
```

## 🤝 Contributing

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

GoVim provides a comprehensive CLI with IPC (Inter-Process Communication) for controlling the running daemon.

### Launch Commands

Start the GoVim daemon:

```bash
# Launch daemon (default)
govim

# Launch with custom config
govim --config /path/to/config.toml

# Explicit launch command (same as above)
govim launch
govim launch --config /path/to/config.toml
```

### Control Commands

Control the running daemon (requires GoVim to be running):

```bash
# Start/resume GoVim if paused
govim start

# Pause GoVim (doesn't quit, just disables functionality)
govim stop

# Show current status
govim status
```

### Mode Commands

Activate different modes (requires GoVim to be running):

```bash
# Activate hint mode (direct click)
govim hints

# Activate hint mode (with action)
govim hints_action

# Activate scroll mode
govim scroll
```

#### Electron Apps Support

GoVim includes special support for Electron-based applications like VS Code and Chrome, with:

- Optimized element detection
- Improved scrolling behavior
- Better hint placement

## 🤝 Contributing

Contributions welcome! Here's how:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `just test` and `just lint` to verify
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
Made with ❤️ by <a href="https://github.com/y3owk1n">y3owk1n</a>
</div>
```

### Examples

```bash
# Start GoVim
govim

# In another terminal, check status
govim status
# Output:
#   GoVim Status:
#     Status: running
#     Mode: idle
#     Config: /Users/you/Library/Application Support/govim/config.toml

# Activate hints mode via CLI
govim hints

# Return to idle
govim idle

# Pause GoVim
govim stop

# Resume
govim start
```

### Error Handling

All control and mode commands will fail gracefully if GoVim is not running:

```bash
$ govim status
Error: govim is not running
Start it first with: govim launch
```

### IPC Architecture

The CLI uses Unix domain sockets (`/tmp/govim.sock`) for communication with the daemon. This allows:

- Fast, reliable communication
- Multiple CLI commands while daemon runs
- Proper error handling when daemon is not running

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
│   ├── cli/            # CLI commands (cobra-based)
│   ├── ipc/            # IPC server/client for CLI communication
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
just build                  # Development build (auto-detects version from git)
just release                # Optimized release build
just build-version v1.0.0   # Build with custom version
just test                   # Run tests
just lint                   # Run linter
```

Version information is automatically injected at build time from git tags. For custom versions:

```bash
# Build with specific version
just build-version v1.0.0

# Manual build with ldflags
go build -ldflags="-s -w -X github.com/y3owk1n/govim/internal/cli.Version=v1.0.0" -o bin/govim cmd/govim/main.go
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
