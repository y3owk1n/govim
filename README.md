# GoVim 🚀

<div align="center">

[![License](https://img.shields.io/github/license/y3owk1n/govim)](LICENSE)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
![Go Version](https://img.shields.io/github/go-mod/go-version/y3owk1n/govim)
[![Latest Release](https://img.shields.io/github/v/release/y3owk1n/govim)](https://github.com/y3owk1n/govim/releases)

**Navigate macOS without touching your mouse - keyboard-driven productivity at its finest 🖱️⌨️**

[Installation](#-installation) •
[Quick Start](#-quick-start) •
[Features](#-features) •
[Configuration](#%EF%B8%8F-configuration) •
[CLI Usage](#-cli-usage)

</div>

## 🎯 Project Philosophy

**GoVim is a free, open-source alternative to Homerow, Shortcat, and Vimac—and it always will be.** Built with Objective-C and Go, it's a working MVP that's proven itself capable of replacing Homerow in daily use for myself.

### Community-Driven Development

This project thrives on community contributions. I'm happy to merge pull requests that align with the project's goals, and I hope GoVim stays current through collective effort rather than solo maintenance, and slowly dies off with hundreds of issues.

### Configuration-First Approach

**No GUI settings panel.** GoVim embraces configuration files over UI settings because:

- ✅ Config files are faster and more powerful, plus easier for dotfile management
- ✅ Easy to version control and share
- ✅ No maintenance burden for UI code
- ✅ Menubar provides quick access to essential commands

This is an intentional design choice to keep the project lean, maintainable, and focused on what matters: keyboard-driven productivity.

**Key Advantages:**

- 🆓 **Always free** - No paywalls, no subscriptions, no "upgrade to pro"
- 🛠️ **Power-user friendly** - Text-based config is version-controllable and shareable
- 🤝 **Community-owned** - Your contributions shape the project's direction
- 🔧 **Scriptable** - CLI commands enable automation and integration

> [!NOTE]
> This is a personal project maintained on a best-effort basis. Pull requests are more likely to be reviewed than feature requests or issues, unless I'm experiencing the same problem.

## ✨ Features

- 🎯 **Hint Mode** - Click any UI element using keyboard shortcuts with visual labels
- 🎬 **Action Mode** - Choose specific click actions (left, right, double, middle click)
- 📜 **Scroll Mode** - Vim-style scrolling (`j`/`k`, `gg`/`G`, etc.) in any application
- 🌐 **Universal Support** - Works with native macOS apps, Electron apps, and Chrome/Firefox
- ⚡ **Performance** - Built with native macOS APIs for instant response
- 🛠️ **Highly Customizable** - Configure hints, hotkeys, and behaviors via TOML
- 🚫 **App Exclusion** - Exclude specific apps where GoVim shouldn't activate
- 🎨 **Minimal UI** - Non-intrusive hints that stay out of your way
- 💬 **IPC Control** - Control the daemon via CLI commands

### Consideration (NOT roadmap)

- Consider zIndex for hints on different layers? There might be multiple layers when we are considering `Menubar`, `Dock` and `Notification bar`
- Better UI representation for action menu (maybe auto edge detection like tooltip in browser, that will place itself around the element based on the space available around it)
- Still a bit slow when activating in windows that have huge lists, such as Mail.app without `unread` filter
- Maybe ditch Accessibility based actions and fully focus on pure mouse click simulation? Noticed that accessibility actions are not very reliable :(
- Find a way to auto deduplicate hints that are targeting the same point
- Scroll areas are hardcoded from 1-9, do we ever need more than that and make it dynamic upon query? We can still use <Tab> to cycle next tho
- Homerow supports `continuous clicks`, is that something important?
- Maybe a new action (move_mouse), that just move the mouse to the point after clicking the hint, maybe useful for trying to hover something?
- Maybe forces to certain input source? I normally just use English, but maybe it's useful for others
- Electron apps does not show hints on the first activation. Maybe there's some manual work need to be done after setting up electron support?
- Cant actually make scroll works in Electron apps, what is the best approach? There are tons of random `AXGroup` where some is scrollable and some is not...
- Sort of working in Mision Control, but it still shows hints from the frontmost app. How can we know that we are in mission control and ignore the frontmost app?
- Firefox seems fine but still the first activation shows very minimal hint after setting it up, same as electron apps
- Research if there's a good way to implemet visual hint mode where we can select text? Doing this with accessibility seems a little hard and vague
- Current `isClickable` heuristic implementation doesn't looks good enough to cover most cases, think of a way that we can determine if the target is actually clickable with a reliable way
- Add more actions to the menubar like `status`, `stop`, `start`, `current version`
- Test suites, but am lazy for it
- Implements launch agent with `start-service` and `stop-service`? Though I am fine just doing it in my nix config directly
- Macos bundle with `GoVim.app` for easier installation? But we'll need a nice app icon for it tho...

## 🚀 Installation

### Homebrew (Recommended)

```bash
brew tap y3owk1n/tap
brew install y3owk1n/tap/govim
```

### Nix Darwin

<details>
<summary>Click to expand Nix configuration</summary>

Add the following file to your overlay:

```nix
{
  stdenv,
  buildGoModule,
  fetchFromGitHub,
  installShellFiles,
  writableTmpDirAsHomeHook,
  lib,
  nix-update-script,
}:
buildGoModule (finalAttrs: {
  pname = "govim";
  version = "1.2.0";

  src = fetchFromGitHub {
    owner = "y3owk1n";
    repo = "govim";
    tag = "v${finalAttrs.version}";
    hash = "sha256-dzcpdVbl2eVU4L9x/qnXC5DOWbMCgt9/wEKroOedXoY=";
  };

  vendorHash = "sha256-x5NB18fP8ERIB5qeMAMyMnSoDEF2+g+NoJKrC+kIj+k=";

  ldflags = [
    "-s"
    "-w"
    "-X github.com/y3owk1n/govim/internal/cli.Version=${finalAttrs.version}"
  ];

  # Completions
  nativeBuildInputs = [
    installShellFiles
    writableTmpDirAsHomeHook
  ];
  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
    installShellCompletion --cmd govim \
      --bash <($out/bin/govim completion bash) \
      --fish <($out/bin/govim completion fish) \
      --zsh <($out/bin/govim completion zsh)
  '';

  passthru = {
    updateScript = nix-update-script { };
  };
})
```

Create a custom module:

```nix
{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.govim;
  configFile = pkgs.writeScript "govim.toml" cfg.config;
in

{
  options = {
    govim = with lib.types; {
      enable = lib.mkEnableOption "Govim keyboard navigation";
      package = lib.mkPackageOption pkgs "govim" { };
      config = lib.mkOption {
        type = types.lines;
        default = ''
          # Your config here
        '';
        description = "Config to use for {file} `govim.toml`.";
      };
    };
  };

  config = (
    lib.mkIf (cfg.enable) {
      environment.systemPackages = [ cfg.package ];

      launchd.user.agents.govim = {
        command =
          "${cfg.package}/bin/govim launch"
          + (lib.optionalString (cfg.config != "") " --config ${configFile}");
        serviceConfig = {
          KeepAlive = false;
          RunAtLoad = true;
        };
      };
    }
  );
}
```

Enable in your configuration:

```nix
{
  imports = [ ./path-to-govim-module.nix ];
  govim.enable = true;
}
```

</details>

### Manual Build

```bash
# Clone and build
git clone https://github.com/y3owk1n/govim.git
cd govim
just build

# Install
mv ./bin/govim /usr/local/bin/govim

# Run
govim launch
```

### ⚠️ Required Permissions

Grant Accessibility permissions:

1. Open **System Settings**
2. Navigate to **Privacy & Security** → **Accessibility**
3. Enable **GoVim**

## 🎯 Quick Start

After installation and granting permissions:

```bash
# Start GoVim
govim launch

# Try hint mode: Press Cmd+Shift+Space
# Try scroll mode: Press Cmd+Shift+J
```

## 🎮 Usage

### Hint Mode (Direct Click)

**Default hotkey:** `Cmd+Shift+Space`

1. Press the hotkey
2. Clickable elements display hint labels (e.g., "AA", "AB", "AC")
3. Type the label to click that element
4. Press `Esc` to cancel

### Hint Mode (Action Selection)

**Default hotkey:** `Cmd+Shift+A`

1. Press the hotkey
2. Type the hint label
3. Choose your action:
   - **Left click** (default)
   - **Right click** (context menu)
   - **Double click**
   - **Middle click**
4. Press `Esc` to cancel

### Scroll Mode

**Default hotkey:** `Cmd+Shift+J`

Vim-style navigation keys:

- `j` / `k` - Scroll down/up
- `h` / `l` - Scroll left/right
- `Ctrl+d` / `Ctrl+u` - Page down/up
- `gg` - Jump to top
- `G` - Jump to bottom
- `Esc` - Exit scroll mode

### Menu Bar

The ⌨️ icon in your menu bar provides quick access to:

- Quit GoVim

## ⚙️ Configuration

### Config File Location

GoVim looks for configuration in:

- **macOS Convention:** `~/Library/Application Support/govim/config.toml`
- **XDG Standard:** `~/.config/govim/config.toml`
- **Custom Path:** Use `--config` flag: `govim launch --config /path/to/config.toml`

### Default Configuration

See [`configs/default-config.toml`](configs/default-config.toml) for all available options.

### Common Configurations

#### Enable Menu Bar and Dock Hints

```toml
[general]
include_menubar_hints = true
include_dock_hints = true  # Also includes Mission Control
```

#### App Exclusion

Exclude specific apps where GoVim shouldn't activate:

```toml
[general]
excluded_apps = [
    "com.apple.Terminal",
    "com.googlecode.iterm2",
    "com.microsoft.rdc.macos",
    "com.vmware.fusion"
]
```

**Find an app's bundle ID:**

```bash
osascript -e 'id of app "AppName"'
```

#### Per-App Role Configuration

Customize clickable/scrollable roles for specific apps. This is useful when certain apps use non-standard UI elements.

**Example:** In Mail.app, many `AXStaticText` elements should be clickable:

```toml
[accessibility]

[[accessibility.app_configs]]
bundle_id = "com.apple.mail"
additional_clickable_roles = ["AXStaticText"]
```

Global roles apply to all apps, while per-app roles are merged with global roles for specific applications.

#### Reduce Hint Label Length

When many elements are clickable, hints use 3 characters (e.g., "AAA") to avoid conflicts. To keep hints at 2 characters:

**Option 1:** Reduce max hints displayed

```toml
[performance]
max_hints_displayed = 80  # Keeps hints at 2 characters with default 9-char set
```

**Option 2:** Add more hint characters

```toml
[hints]
hint_characters = "asdfghjklqwertyuiop"  # 19 chars = 361 two-char combinations
```

**Hint capacity reference:**

- 9 characters = 81 two-char hints
- 12 characters = 144 two-char hints
- 15 characters = 225 two-char hints

### Electron & Chrome Support

GoVim includes built-in support for Electron apps and Chromium browsers.

**Supported Electron apps** (automatic):

- VS Code
- Windsurf
- Cursor
- Slack
- Spotify
- Obsidian

**Supported Chromium browsers** (automatic):

- Chrome
- Brave
- Helium
- Arc

**Supported Firefox browsers** (automatic):

- Firefox
- Zen

> [!NOTE]
> You can still add your browsers to `additional_bundles` if they aren't detected automatically.
> Right now, chromium, firefox and electron apps are all under the same key, in future we might split them out.

**Configuration:**

```toml
[accessibility.electron_support]
enable = true  # Enable Electron support (default)

# Add additional Electron apps
additional_bundles = [
    "com.example.app",          # Exact bundle ID
    "com.company.prefix*"       # Wildcard for multiple apps
]
```

## 🖥️ CLI Usage

GoVim provides comprehensive CLI commands with IPC (Inter-Process Communication) for controlling the daemon.

### Launch Daemon

```bash
govim launch                              # Use default config location
govim launch --config /path/to/config.toml  # Use custom config
```

### Control Commands

These require GoVim to be running:

```bash
govim start     # Resume GoVim if paused
govim stop      # Pause GoVim (doesn't quit, just disables functionality)
govim status    # Show current status (running/paused, current mode, config path)
```

### Mode Activation

```bash
govim hints         # Activate hint mode (direct click)
govim hints_action  # Activate hint mode with action selection
govim scroll        # Activate scroll mode
govim idle          # Return to idle state
```

### Example Workflow

```bash
# Start GoVim
govim launch

# Check status in another terminal
govim status
# Output:
#   GoVim Status:
#     Status: running
#     Mode: idle
#     Config: /Users/you/Library/Application Support/govim/config.toml

# Activate hints via CLI instead of hotkey
govim hints

# Pause functionality temporarily
govim stop

# Resume
govim start
```

### IPC Architecture

The CLI uses Unix domain sockets (`/tmp/govim.sock`) for:

- Fast, reliable communication between CLI and daemon
- Multiple concurrent CLI commands
- Proper error handling when daemon isn't running

## 🔧 Troubleshooting

### Accessibility Permissions Not Working

1. Open **System Settings** → **Privacy & Security** → **Accessibility**
2. Remove GoVim from the list if present
3. Re-add GoVim
4. Restart GoVim

### Hints Not Appearing in Electron Apps

If hints don't appear in Electron app content (VS Code, Windsurf, etc.):

1. Verify Electron support is enabled (it's on by default):

   ```toml
   [accessibility.electron_support]
   enable = true
   ```

2. Check logs for these messages:
   - `"App requires Electron support"`
   - `"Enabled AXManualAccessibility"`

3. If your app isn't in the built-in list, add it manually:

   ```toml
   [accessibility.electron_support]
   additional_bundles = ["com.your.app"]
   ```

### Viewing Logs

**Log location:** `~/Library/Logs/govim/app.log`

**Enable debug logging:**

```toml
[logging]
log_level = "debug"
```

### GoVim Not Responding

```bash
# Check if daemon is running
govim status

# If not running, restart
govim launch

# If stuck, force quit and restart
pkill govim
govim launch
```

## 🏗️ Architecture

GoVim is built with:

- **Go** - Core application logic, configuration, and CLI
- **CGo/Objective-C** - macOS Accessibility API integration
- **Native macOS APIs** - Overlay rendering and hotkey management

### Project Structure

```
govim/
├── cmd/govim/           # Main entry point
├── internal/
│   ├── accessibility/   # Accessibility API wrappers
│   ├── hints/          # Hint generation and display logic
│   ├── scroll/         # Scroll mode implementation
│   ├── hotkeys/        # Global hotkey management
│   ├── config/         # TOML configuration parsing
│   ├── cli/            # CLI commands (Cobra-based)
│   ├── ipc/            # IPC server/client for daemon control
│   └── bridge/         # CGo/Objective-C bridges
├── configs/            # Default configuration files
└── scripts/            # Build and packaging scripts
```

## 💻 Development

### Prerequisites

- Go 1.21+
- macOS development tools
- Just (command runner)

### Building

```bash
just build                  # Development build (auto-detects version from git)
just release                # Optimized release build
just build-version v1.0.0   # Build with custom version
just test                   # Run tests
just test-race              # Run tests with race detection
just lint                   # Run linter
```

Version information is automatically injected at build time from git tags.

**Manual build with custom version:**

```bash
go build -ldflags="-s -w -X github.com/y3owk1n/govim/internal/cli.Version=v1.0.0" \
  -o bin/govim cmd/govim/main.go
```

### Testing

```bash
just test           # Run unit tests
just test-race      # Run tests with race detection
just lint           # Run golangci-lint
```

## 🤝 Contributing

Contributions are welcome! Here's how to contribute:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Make** your changes
4. **Test** your changes (`just test && just lint`)
5. **Commit** (`git commit -m 'Add amazing feature'`)
6. **Push** to your branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Contribution Guidelines

- Write clear commit messages
- Add tests for new features
- Update documentation as needed
- Follow existing code style
- Keep PRs focused on a single change

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

GoVim is inspired by these excellent projects:

- [Homerow](https://www.homerow.app/) - Modern keyboard navigation for macOS
- [Vimac](https://github.com/dexterleng/vimac) - Vim-style keyboard navigation
- [Shortcat](https://shortcat.app/) - Keyboard productivity tool
- [Vimium](https://github.com/philc/vimium) - Vim bindings for browsers

## 🌟 Star History

If you find GoVim useful, consider giving it a star on GitHub!

---

<div align="center">

Made with ❤️ by [y3owk1n](https://github.com/y3owk1n)

[Report Bug](https://github.com/y3owk1n/govim/issues) · [Request Feature](https://github.com/y3owk1n/govim/issues) · [Discussions](https://github.com/y3owk1n/govim/discussions)

</div>
