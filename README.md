# Neru (練る)

> Master your keyboard, refine your workflow

<div align="center">

[![License](https://img.shields.io/github/license/y3owk1n/neru)](LICENSE)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
![Go Version](https://img.shields.io/github/go-mod/go-version/y3owk1n/neru)
[![Latest Release](https://img.shields.io/github/v/release/y3owk1n/neru)](https://github.com/y3owk1n/neru/releases)

**Navigate macOS without touching your mouse - keyboard-driven productivity at its finest 🖱️⌨️**

[Installation](#-installation) •
[Quick Start](#-quick-start) •
[Features](#-features) •
[Configuration](#%EF%B8%8F-configuration) •
[CLI Usage](#%EF%B8%8F-cli-usage)

</div>

## About Neru

**Neru** (練る) - A Japanese word meaning "to refine, polish, and master through practice" - embodies the philosophy of this tool: mastering keyboard-driven navigation to refine your macOS workflow.

Neru is a free, open-source alternative to Homerow, Shortcat, and Vimac—and it always will be. Built with Objective-C and Go, it's a working MVP that's proven itself capable of replacing Homerow in daily use for myself.

This project thrives on community contributions. I'm happy to merge pull requests that align with the project's goals, and I hope Neru stays current through collective effort rather than solo maintenance, and slowly dies off with hundreds of issues.

**No GUI settings panel.** Neru embraces configuration files over UI settings because:

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
- 🚫 **App Exclusion** - Exclude specific apps where Neru shouldn't activate
- 🎨 **Minimal UI** - Non-intrusive hints that stay out of your way
- 💬 **IPC Control** - Control the daemon via CLI commands

### Consideration (NOT roadmap)

- Consider zIndex for hints on different layers? There might be multiple layers when we are considering `Menubar`, `Dock` and `Notification bar`
- Better UI representation for action menu (maybe auto edge detection like tooltip in browser, that will place itself around the element based on the space available around it)
- Find a way to auto deduplicate hints that are targeting the same point
- Scroll areas are hardcoded from 1-9, do we ever need more than that and make it dynamic upon query? We can still use <Tab> to cycle next tho
- Homerow supports `continuous clicks`, is that something important?
- Maybe forces to certain input source? I normally just use English, but maybe it's useful for others
- Electron apps does not show hints on the first activation. Maybe there's some manual work need to be done after setting up electron support?
- Cant actually make scroll works in Electron apps, what is the best approach? There are tons of random `AXGroup` where some is scrollable and some is not...
- Sort of working in Mision Control, but it still shows hints from the frontmost app. How can we know that we are in mission control and ignore the frontmost app?
- Firefox seems fine but still the first activation shows very minimal hint after setting it up, same as electron apps
- Research if there's a good way to implemet visual hint mode where we can select text? Doing this with accessibility seems a little hard and vague
- Add more actions to the menubar like `status`, `stop`, `start`, `current version`
- Test suites, but am lazy for it
- Implements launch agent with `start-service` and `stop-service`? Though I am fine just doing it in my nix config directly

## 🚀 Installation

### Homebrew (Recommended)

```bash
brew tap y3owk1n/tap
brew install --cask y3owk1n/tap/neru
```

### Nix Darwin

<details>
<summary>Click to expand Nix configuration</summary>

Add the following file to your overlay:

```nix
{
  fetchzip,
  gitUpdater,
  installShellFiles,
  stdenv,
  versionCheckHook,
  lib,
}:

let
  appName = "Neru.app";
  version = "1.5.0";
in
stdenv.mkDerivation {
  pname = "neru";

  inherit version;

  src = fetchzip {
    url = "https://github.com/y3owk1n/neru/releases/download/v${version}/neru-darwin-arm64.zip";
    sha256 = "sha256-TcurBgkj3/vmOd40rBJzoVr2+rcOspR0QZ1zv9sN6O4=";
    stripRoot = false;
  };

  nativeBuildInputs = [ installShellFiles ];

  installPhase = ''
    runHook preInstall
    mkdir -p $out/Applications
    mv ${appName} $out/Applications
    cp -R bin $out
    mkdir -p $out/share
    runHook postInstall
  '';

  postInstall = ''
    if ${lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) "true"}; then
      installShellCompletion --cmd neru \
        --bash <($out/bin/neru completion bash) \
        --fish <($out/bin/neru completion fish) \
        --zsh <($out/bin/neru completion zsh)
    fi
  '';

  doInstallCheck = true;
  nativeInstallCheckInputs = [
    versionCheckHook
  ];

  passthru.updateScript = gitUpdater {
    url = "https://github.com/y3owk1n/neru.git";
    rev-prefix = "v";
  };

  meta = {
    mainProgram = "neru";
  };
}
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
  cfg = config.neru;
  configFile = pkgs.writeScript "neru.toml" cfg.config;
in

{
  options = {
    neru = with lib.types; {
      enable = lib.mkEnableOption "neru keyboard navigation";
      package = lib.mkPackageOption pkgs "neru" { };
      config = lib.mkOption {
        type = types.lines;
        default = ''
          # Your config here
        '';
        description = "Config to use for {file} `neru.toml`.";
      };
    };
  };

  config = (
    lib.mkIf (cfg.enable) {
      environment.systemPackages = [ cfg.package ];

      launchd.user.agents.neru = {
        command =
          "${cfg.package}/bin/neru launch"
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
  imports = [ ./path-to-neru-module.nix ];
  neru.enable = true;
}
```

</details>

### Manual Build

```bash
# Clone and build
git clone https://github.com/y3owk1n/neru.git
cd neru

just release # if you prefer cli only
just bundle # if you prefer app bundle

# You can then move to where you want
mv ./bin/neru /usr/local/bin/neru # if you prefer cli only (bin path depends on your setup)
mv ./build/Neru.app /Applications # if you prefer app bundle

# Run
open -a Neru # if you prefer app bundle
neru launch # if you prefer cli only
```

### ⚠️ Required Permissions

Grant Accessibility permissions:

1. Open **System Settings**
2. Navigate to **Privacy & Security** → **Accessibility**
3. Enable **Neru**

## 🎯 Quick Start

After installation and granting permissions:

```bash
# Start Neru
open -a Neru # if you prefer app bundle
neru launch # if you prefer cli only

# Try hint mode: Press Cmd+Shift+Space
# Try scroll mode: Press Cmd+Shift+J
```

## 🎮 Usage

### Hint Mode (Direct Click)

**Default hotkey:** `Cmd+Shift+Space`

1. Press the hotkey
2. Clickable elements display hint labels (e.g., "AA", "AB", "AC") - You can also use `delete` just like in vimium if you type something wrong, instead of needing to reactivate the hint again.
3. Type the label to left click that element
4. Restore cursor to previous position (if enabled)

Press `esc` anytime to quit the hint selection.

### Hint Mode (Action Selection)

**Default hotkey:** `Cmd+Shift+A`

1. Press the hotkey
2. Type the hint label
3. Choose your action:
   - **Left click** (default)
   - **Right click** (context menu)
   - **Double click**
   - **Middle click**
   - **Go to a position** (move mouse)
4. Restore cursor to previous position (if enabled)

Press `esc` anytime to quit the hint selection.

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

- Quit Neru

## ⚙️ Configuration

### Config File Location

Neru looks for configuration in:

- **macOS Convention:** `~/Library/Application Support/neru/config.toml`
- **XDG Standard:** `~/.config/neru/config.toml` (Prefer this more so that we can put this in our dotfile)
- **Custom Path:** Use `--config` flag: `neru launch --config /path/to/config.toml` (Useful for nix users)

> [!NOTE]
> If theres no config file, Neru will use the defaults.

### Default Configuration

See [`configs/default-config.toml`](configs/default-config.toml) for all available options.

### Common Configurations

#### Change the default keybindings

```toml
[hotkeys]
# all hotkeys can be disabled by either setting the key to "" or just commenting it out
activate_hint_mode = "Ctrl+F"
activate_hint_mode_with_actions = "Ctrl+G"
activate_scroll_mode = "Ctrl+S"
```

You shoul be also able to just clear the keybind and bind it with something like skhd or any similar tools, since we exposes commands in the cli through IPC. For example in skhd:

```bash
ctrl - f : neru hints
ctrl - g : neru hints_action
ctrl - s : neru scroll
```

Read more about [CLI Usage](#%EF%B8%8F-cli-usage).

#### Enable Menu Bar Dock and Notification Center Hints

```toml
[general]
include_menubar_hints = true
# Do nothing if `include_menubar_hints` is false
additional_menubar_hints_targets = [
  "com.apple.TextInputMenuAgent", # TextInput menu bar
  "com.apple.controlcenter",      # Control Center
  "com.apple.systemuiserver",     # SystemUIServer (e.g. Siri in menubar)
]

include_dock_hints = true  # Also includes Mission Control
include_nc_hints = true # Notification popups
```

#### Restore cursor after click action

```toml
[general]
restore_pos_after_left_click = true
restore_pos_after_right_click = true
restore_pos_after_middle_click = true
restore_pos_after_double_click = true
```

#### App Exclusion

Exclude specific apps where Neru shouldn't activate. This will also unregister all the binded hotkeys when the app is focused so that in specific app, you can use the hotkey to do something else.

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

### Electron & Chrome Support

Neru includes built-in support for Electron apps and Chromium browsers. (Note that this is off by default)

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
enable = true  # Enable Electron support (off by default)

# Add additional Electron apps
additional_bundles = [
    "com.example.app",          # Exact bundle ID
    "com.company.prefix*"       # Wildcard for multiple apps
]
```

### Look and feel for hints

You can configure how you want those hints

```toml
[hints]
# Font size for hint labels
# Valid range: 6–72
font_size = 12

# Font family (leave empty for system default)
font_family = ""

# Background color (hex format)
background_color = "#FFD700"

# Text color (hex format)
text_color = "#000000"

# Matched text color - color for characters that have been typed (hex format)
matched_text_color = "#0066CC"

# Border radius (pixels)
# Non-negative integer
border_radius = 4

# Padding (pixels)
# Non-negative integer
padding = 4

# Border width (pixels)
# Non-negative integer
border_width = 1

# Border color (hex format)
border_color = "#000000"

# Opacity (0.0 to 1.0)
# Controls hint translucency; 1.0 is fully opaque
opacity = 0.95

# Action overlay colors (used when selecting click type)
# Background color (hex format)
action_background_color = "#66CCFF"

# Text color (hex format)
action_text_color = "#000000"

# Matched text color (hex format)
action_matched_text_color = "#003366"

# Border color (hex format)
action_border_color = "#000000"

# Opacity (0.0 to 1.0)
action_opacity = 0.95
```

### Scroll configuration

Note that this is a simple implementation for scrolling, not exactly working fine yet, help would be appreciated (especially in electron and chromium applications)

```toml
[scroll]
# Base scroll amount for j/k keys in pixels
# Minimum: 1. Increase for faster per-press scrolling.
scroll_speed = 50

# Highlight the active scroll area with a border
# When true, draws a border around the detected scroll container.
highlight_scroll_area = true

# Highlight border color (hex format)
highlight_color = "#FF0000"

# Highlight border width in pixels
highlight_width = 2

# Estimated page height in pixels (used for calculating Ctrl+D/U scroll distance)
# Increase if your app windows are taller; decrease for short panes.
page_height = 1200

# Half-page scroll multiplier for Ctrl+D/U (0.5 = 600px with default page_height)
# Valid range: (0, 1]
half_page_multiplier = 0.5

# Full-page scroll multiplier (not currently used)
# Valid range: (0, 1]
full_page_multiplier = 0.9

# Number of scroll events to send for gg/G commands
# Higher values create longer smooth scrolls to edges.
scroll_to_edge_iterations = 20

# Pixels to scroll per iteration for gg/G (total = iterations * delta)
# Increase for faster edge scrolling; tune with iterations above.
scroll_to_edge_delta = 5000
```

## 🖥️ CLI Usage

Neru provides comprehensive CLI commands with IPC (Inter-Process Communication) for controlling the daemon.

### Launch Daemon

```bash
neru launch                              # Use default config location
neru launch --config /path/to/config.toml  # Use custom config
```

### Control Commands

These require Neru to be running:

```bash
neru start     # Resume Neru if paused
neru stop      # Pause Neru (doesn't quit, just disables functionality)
neru status    # Show current status (running/paused, current mode, config path)
```

### Mode Activation

```bash
neru hints         # Activate hint mode (direct click)
neru hints_action  # Activate hint mode with action selection
neru scroll        # Activate scroll mode
neru idle          # Return to idle state
```

### Example Workflow

```bash
# Start Neru
neru launch

# Check status in another terminal
neru status
# Output:
#   Neru Status:
#     Status: running
#     Mode: idle
#     Config: /Users/you/Library/Application Support/neru/config.toml

# Activate hints via CLI instead of hotkey
neru hints

# Pause functionality temporarily
neru stop

# Resume
neru start
```

### IPC Architecture

The CLI uses Unix domain sockets (`/tmp/neru.sock`) for:

- Fast, reliable communication between CLI and daemon
- Multiple concurrent CLI commands
- Proper error handling when daemon isn't running

## 🔧 Troubleshooting

### Accessibility Permissions Not Working

1. Open **System Settings** → **Privacy & Security** → **Accessibility**
2. Remove Neru from the list if present
3. Re-add Neru
4. Restart Neru

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

**Log location:** `~/Library/Logs/neru/app.log`

**Enable debug logging:**

```toml
[logging]
log_level = "debug"
```

### Neru Not Responding

```bash
# Check if daemon is running
neru status

# If not running, restart
neru launch

# If stuck, force quit and restart
pkill neru
neru launch
```

## 🏗️ Architecture

Neru is built with:

- **Go** - Core application logic, configuration, and CLI
- **CGo/Objective-C** - macOS Accessibility API integration
- **Native macOS APIs** - Overlay rendering and hotkey management

### Project Structure

```
neru/
├── cmd/neru/           # Main entry point
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

- Go 1.25+
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
go build -ldflags="-s -w -X github.com/y3owk1n/neru/internal/cli.Version=v1.0.0" \
  -o bin/neru cmd/neru/main.go
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

Neru is inspired by these excellent projects:

- [Homerow](https://www.homerow.app/) - Modern keyboard navigation for macOS
- [Vimac](https://github.com/dexterleng/vimac) - Vim-style keyboard navigation
- [Shortcat](https://shortcat.app/) - Keyboard productivity tool
- [Vimium](https://github.com/philc/vimium) - Vim bindings for browsers

## 🌟 Star History

If you find Neru useful, consider giving it a star on GitHub!

---

<div align="center">

Made with ❤️ by [y3owk1n](https://github.com/y3owk1n)

[Report Bug](https://github.com/y3owk1n/neru/issues) · [Request Feature](https://github.com/y3owk1n/neru/issues) · [Discussions](https://github.com/y3owk1n/neru/discussions)

</div>
