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

> [!NOTE]
> This is a personal project maintained on a best-effort basis. PRs are more likely to be reviewed than feature requests or issues, unless I am facing the same problem.

## ✨ Features

- 🎯 **Hint Mode** - Click any UI element using keyboard shortcuts
- 📜 **Scroll Mode** - Vim-style scrolling in any application
- 🌐 **Universal** - Works with native macOS apps and Electron apps
- ⚡ **Performance** - Built with native macOS APIs for instant response
- 🛠️ **Customizable** - Configure hints, hotkeys, and behaviors via TOML
- 🎨 **Minimal UI** - Non-intrusive hints that don't get in your way

## 🚀 Installation

### Homebrew

```bash
brew tap y3owk1n/tap
brew install y3owk1n/tap/govim
```

### Nix Darwin

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
  version = "1.0.2";

  src = fetchFromGitHub {
    owner = "y3owk1n";
    repo = "govim";
    tag = "v${finalAttrs.version}";
    hash = "sha256-0iuZsYxND490TWueNxsGNE/0k7g3xVOUNOQD/xkIFs0=";
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

Then, you can add the following to as a custom module package:

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

Then somewhere in your nix config, you can add the following to enable the module:

```nix
{
  imports = [
    ./path-to-govim-module.nix
  ];

  govim = {
    enable = true;
  };
}
```

### Manual Build

```bash
# Clone the repository
git clone https://github.com/y3owk1n/govim.git
cd govim

# Build
just build

# Move the binary to a location in your $PATH
mv ./bin/govim /usr/local/bin/govim # Or anywhere else

# Run
govim launch # or without path `./bin/govim launch`

```

### Required Permissions

⚠️ Grant Accessibility permissions in:
`System Settings` → `Privacy & Security` → `Accessibility`

## 🎮 Usage

### Hint Mode (Direct Click)

1. Press `Cmd+Shift+Space` (default)
2. Clickable elements show hint labels
3. Type the label to click (e.g., "aa", "ab")
4. Press `Esc` to exit

### Hint Mode (Action)

1. Press `Cmd+Shift+A` (default)
2. Clickable elements show hint labels
3. Type the label to click (e.g., "aa", "ab")
4. Choose the action to perform (e.g., "left click", "right click", "double click", "middle click")
5. Press `Esc` to exit

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

### Include hints on the macOS menu bar and Dock

```toml
[hints]
menubar = true
dock = true
```

### Per-App Role Configuration

GoVim supports global and per-app accessibility role configurations:

- **Global roles** (`clickable_roles` and `scrollable_roles`) are used for all applications
- **Per-app roles** allow you to add additional roles for specific applications
- The final roles used = global roles + app-specific additional roles (merged)

To find an app's bundle ID:

```bash
osascript -e 'id of app "AppName"'
```

For example, in `Mail.app`, lots of element are `AXStaticText` and they should be clickable. In this case, we don't want to add to the global one, as it will causes lots of unclickable hints, especially in browser space.

```toml
[[accessibility.app_configs]]
bundle_id = "com.apple.mail"
additional_clickable_roles = ["AXStaticText"]
```

### Electron & Chrome Support

> [!NOTE]
> Electron, firefox and chrome support is a mess, and I'm not even sure if I am doing it correctly or not.
> Feel free to help out if you know better about this...

GoVim includes built-in support for Electron apps (VS Code, Windsurf, Slack, etc.) and Chromium browsers.

- `accessibility.electron_support.enable` toggles automatic enabling of Electron accessibility hooks for legacy apps.
- `accessibility.electron_support.additional_bundles` accepts exact bundle IDs or `prefix*` wildcards for extra Electron apps that require manual accessibility.

Built-in enabled apps:

- `org.mozilla.firefox` - Firefox
- `com.microsoft.VSCode` - VS Code
- `com.exafunction.windsurf` - Windsurf
- `com.todesktop.230313mzl4w4u92` - Cursor
- `com.tinyspeck.slackmacgap` - Slack
- `com.spotify.client` - Spotify
- `md.obsidian` - Obsidian
- For other apps, you can add them to `accessibility.electron_support.additional_bundles` in your config, or just submit a PR to add into the source

Note that you don't have to add chromium bundles here. To support it, you need to:

1. Go to `chrome://accessibility`
2. Turn on `Native accessibility API support`
3. Turn on `Web accessibility`

> [!NOTE]
> Not sure if there's a better way to do this... feel free to help out!

## CLI Usage

GoVim provides a comprehensive CLI with IPC (Inter-Process Communication) for controlling the running daemon.

### Launch Commands

Start the GoVim daemon:

```bash
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

### Examples

```bash
# Start GoVim
govim launch

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Inspired by:

- [Homerow](https://www.homerow.app/)
- [Vimac](https://github.com/dexterleng/vimac)
- [Shortcat](https://shortcat.app/)
- [Vimium](https://github.com/philc/vimium)

## 🤝 Contributing

Contributions welcome! Here's how:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `just test` and `just lint` to verify
5. Submit a pull request

---

<div align="center">
Made with ❤️ by <a href="https://github.com/y3owk1n">y3owk1n</a>
</div>
