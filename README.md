# Neru

**Master your keyboard, refine your workflow**

Navigate macOS without touching your mouse - keyboard-driven productivity at its finest 🖱️⌨️

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Contributing](#contributing)

<https://github.com/user-attachments/assets/left-click-demo.mov>

---

## What is Neru?

Neru (練る) - a Japanese word meaning "to refine, polish, and master through practice" - is a free, open-source keyboard navigation tool for macOS. Navigate, click, and scroll anywhere on your screen without ever touching your mouse.

**Grid-based navigation** is the foundation of Neru. Unlike hint-based systems that rely on accessibility trees (which break in Electron apps, Chromium, Mission Control, menubar items), Neru's grid approach:

- ✅ Works everywhere - native apps, Electron, browsers, menubar, Mission Control, Dock
- ✅ Fast and reliable - no waiting for accessibility tree traversal
- ✅ Simple to maintain - no complex app-specific compatibility layers
- ✅ Always accurate - clicks exactly where you see the hint

**Why Neru?**

- 🆓 **Always free** - No paywalls, no subscriptions, no "upgrade to pro"
- 🎯 **Universal compatibility** - Works with native macOS apps, Electron apps, and all browsers
- ⚡ **Lightning fast** - Native performance with instant response
- 🛠️ **Power-user friendly** - Text-based config for version control and dotfile management
- 🤝 **Community-owned** - Your contributions shape the project
- 🔧 **Scriptable** - CLI commands enable automation and integration

### Free Alternative To

Neru is a capable replacement for:

- [Homerow](https://www.homerow.app/) - Modern keyboard navigation (paid)
- [Shortcat](https://shortcat.app/) - Keyboard productivity tool (discontinued? not sure...)
- [Vimac](https://github.com/dexterleng/vimac) - Vim-style navigation (unmaintained)

---

## Features

- 🎯 **Grid-based hints** - Click any visible element using keyboard labels
- 🎬 **Action mode** - Choose click type: left, right, double, middle, drag
- 📜 **Scroll mode** - Vim-style scrolling (`j`/`k`, `gg`/`G`) in any application
- 🌐 **Universal support** - Native apps, Electron, Chrome, Firefox, menubar, Dock, Mission Control
- ⚡ **Native performance** - Built with Objective-C and Go for instant response
- 🛠️ **Highly customizable** - Configure everything via TOML
- 🚫 **App exclusion** - Disable Neru in specific applications
- 💬 **IPC control** - Control via CLI for scripting and automation
- 📦 **No GUI bloat** - Configuration over UI for maintainability

---

## Installation

### Homebrew (Recommended)

```bash
brew tap y3owk1n/tap
brew install --cask y3owk1n/tap/neru
```

### Nix

See [docs/INSTALLATION.md](docs/INSTALLATION.md#nix) for Nix configuration details.

### From Source

```bash
git clone https://github.com/y3owk1n/neru.git
cd neru
just release # CLI only
just bundle  # App bundle
```

See [docs/INSTALLATION.md](docs/INSTALLATION.md) for detailed installation instructions and troubleshooting.

### Grant Permissions

After installation, grant Accessibility permissions:

1. Open **System Settings**
2. Navigate to **Privacy & Security → Accessibility**
3. Enable **Neru**

---

## Quick Start

```bash
# Start Neru
open -a Neru  # App bundle
# or
neru launch   # CLI

# Try the default hotkeys:
# Cmd+Shift+Space - Hint mode (left click)
# Cmd+Shift+A     - Hints with Action mode (choose click type)
# Cmd+Shift+J     - Scroll mode
# Cmd+Shift+G     - Grid mode (left click)
```

**Basic workflow:**

For hints:

1. Press hotkey to activate hints
2. Type the label characters (e.g., "aa", "ab") - use `delete` to fix typos
3. Element is clicked, cursor optionally restores to original position
4. Press `Esc` anytime to cancel

For grid:

1. Press hotkey to activate grid mode
2. Click the cell you want to click
3. Refine the final selection within the selected cell
4. Element is clicked, cursor optionally restores to original position
5. Press `Esc` anytime to cancel

---

## Documentation

- **[Installation Guide](docs/INSTALLATION.md)** - Detailed installation for Homebrew, Nix, and source builds
- **[Configuration](docs/CONFIGURATION.md)** - Complete configuration reference with examples
- **[CLI Usage](docs/CLI.md)** - Command-line interface and IPC control
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Development](docs/DEVELOPMENT.md)** - Building, testing, and contributing

### Configuration Files

Neru uses TOML for configuration. Default locations (in order of preference):

1. `~/.config/neru/config.toml` (XDG standard - **recommended for dotfiles**)
2. `~/Library/Application Support/neru/config.toml` (macOS convention)
3. Custom path: `neru launch --config /path/to/config.toml`

**No config file?** Neru uses sensible defaults.

See [configs/default-config.toml](configs/default-config.toml) for all options, or check [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for examples.

---

## Design Philosophy

### Why No GUI?

Neru intentionally avoids a GUI settings panel because:

✅ **Config files are superior for power users**

- Faster to edit than clicking through settings
- Version control friendly (git, dotfiles)
- Easily shared and documented
- More powerful than UI constraints allow

✅ **Reduces maintenance burden**

- No UI code to maintain
- Focus on core functionality
- Smaller, simpler codebase

✅ **Menubar provides essentials**

- Quick access to common actions
- Status information
- Enable/disable toggle

This is an intentional choice to keep Neru lean, maintainable, and focused on what matters: **keyboard-driven productivity**.

### Grid-Based > Hint-Based

Neru uses a **grid-based approach** for hint placement, not accessibility tree traversal:

| Grid-Based (Neru)           | Hint-Based (Traditional)      |
| --------------------------- | ----------------------------- |
| ✅ Works everywhere         | ❌ Breaks in Electron         |
| ✅ Works in menubar         | ❌ No menubar support         |
| ✅ Works in Mission Control | ❌ No Mission Control         |
| ✅ Fast (instant)           | ❌ Slower (tree walk)         |
| ✅ Simple maintenance       | ❌ Complex app-specific fixes |
| ✅ Always accurate          | ❌ Misaligned hints           |
| ✅ No side effects          | ❌ Can break tiling WMs       |

Grid-based navigation means Neru doesn't depend on apps exposing proper accessibility information. It works by overlaying a visual grid and clicking at exact screen coordinates - simple, reliable, universal.

**Important:** Neru includes optional accessibility support for Chromium and Firefox (disabled by default) that can help with hint detection. However, enabling this may cause side effects with tiling window managers (yabai, Amethyst, etc.). If you use a tiling WM, keep `additional_ax_support.enable = false` unless absolutely necessary.

---

## Project Status

> [!NOTE]
> Neru is a personal project maintained on a best-effort basis. **Pull requests are more likely to be reviewed than feature requests or issues**, unless I'm experiencing the same problem.

This project thrives on community contributions. I'm happy to merge PRs that align with the project's goals. Neru stays current through collective effort rather than solo maintenance.

### Roadmap / Future Ideas

- [ ] Auto-deduplicate hints targeting the same point
- [ ] Continuous clicks (like Homerow)
- [ ] Test suites (contributions welcome!)
- [ ] Launch agent with `start-service`/`stop-service` commands
- [ ] Better app icon and menubar icon
- [ ] Full vim mode (like [Vimnav.spoon](https://github.com/y3owk1n/vimnav.spoon))
- [ ] Expose mouse action to CLI (like `neru left_click`) and perform the action on the current cursor position

**Known Issues:**

- Hold/unhold actions don't work in Finder.app (help is appreciated!)

---

## Contributing

Contributions are welcome! Here's how:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test thoroughly (`just test && just lint`)
5. Commit with clear messages
6. Push and open a Pull Request

**Guidelines:**

- Keep PRs focused on a single change
- Add tests for new features
- Update documentation
- Follow existing code style

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for build instructions and architecture details.

---

## Compatibility

Neru works with:

- ✅ All native macOS apps (Finder, Safari, System Settings, etc.)
- ✅ Electron apps (VS Code, Windsurf, Cursor, Slack, Spotify, Obsidian)
- ✅ Chromium browsers (Chrome, Brave, Arc)
- ✅ Firefox browsers (Firefox, Zen)
- ✅ Adobe Creative Suite (Illustrator, Photoshop)
- ✅ Menubar applications
- ✅ Dock and Mission Control

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) if you encounter issues with specific apps.

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

Neru is inspired by these excellent projects:

- [Homerow](https://www.homerow.app/) - Modern keyboard navigation for macOS
- [Vimac](https://github.com/dexterleng/vimac) - Vim-style keyboard navigation
- [Shortcat](https://shortcat.app/) - Keyboard productivity tool
- [Vimium](https://github.com/philc/vimium) - Vim bindings for browsers
- [Vifari](https://github.com/dzirtusss/vifari) - Vimium/Vimari for Safari

---

## Support

- 📖 Check [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common issues
- 🐛 Open an issue for bugs (but PRs are preferred!)
- ⭐ Star the repo if you find Neru useful!

Made with ❤️ by [y3owk1n](https://github.com/y3owk1n)
