# Installation Guide

This guide covers all installation methods for Neru on macOS.

## Prerequisites

- macOS 11.0 or later
- Accessibility permissions (granted after installation)

---

## Method 1: Homebrew (Recommended)

The easiest way to install Neru:

```bash
brew tap y3owk1n/tap
brew install --cask y3owk1n/tap/neru
```

**To update:**

```bash
brew upgrade --cask neru
```

**To uninstall:**

```bash
brew uninstall --cask neru
```

---

## Method 2: Nix

### Overlay Configuration

Add this overlay file to your Nix configuration:

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
  version = "1.5.0"; # change this to the latest version
in
stdenv.mkDerivation {
  pname = "neru";
  inherit version;

  src = fetchzip {
    url = "https://github.com/y3owk1n/neru/releases/download/v${version}/neru-darwin-arm64.zip";
    sha256 = "sha256-TcurBgkj3/vmOd40rBJzoVr2+rcOspR0QZ1zv9sN6O4="; # run rebuild and update the latest sha256
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
  nativeInstallCheckInputs = [ versionCheckHook ];

  passthru.updateScript = gitUpdater {
    url = "https://github.com/y3owk1n/neru.git";
    rev-prefix = "v";
  };

  meta = {
    mainProgram = "neru";
  };
}
```

### Module Configuration

Create a custom module for Neru:

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
      enable = lib.mkEnableOption "Neru keyboard navigation";
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
          "${cfg.package}/Applications/Neru.app/Contents/MacOS/Neru launch"
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

### Enable in Configuration

```nix
{
  imports = [ ./path-to-neru-module.nix ];

  neru.enable = true;
  neru.config = ''
    [hotkeys]
    "Cmd+Shift+Space" = "hints left_click"

    [general]
    excluded_apps = ["com.apple.Terminal"]
  '';
}
```

---

## Method 3: From Source

### Requirements

- Go 1.25 or later
- Xcode Command Line Tools
- [Just](https://github.com/casey/just) (command runner)

### Build Steps

```bash
# Clone repository
git clone https://github.com/y3owk1n/neru.git
cd neru

# Build CLI only
just release

# Or build app bundle
just bundle

# Move to installation location
# For CLI:
mv ./bin/neru /usr/local/bin/neru

# For app bundle:
mv ./build/Neru.app /Applications/Neru.app
```

### Manual Build (without Just)

```bash
# Build with version info
go build \
  -ldflags="-s -w -X github.com/y3owk1n/neru/internal/cli.Version=$(git describe --tags --always)" \
  -o bin/neru \
  ./cmd/neru
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for more build options.

---

## Post-Installation

### 1. Grant Accessibility Permissions

**Required for Neru to function:**

1. Open **System Settings**
2. Navigate to **Privacy & Security → Accessibility**
3. Click the lock icon to make changes
4. Click **+** and add Neru
5. Ensure the checkbox is enabled

### 2. Start Neru

**For app bundle:**

```bash
open -a Neru
```

**For CLI:**

```bash
neru launch
```

**With custom config:**

```bash
neru launch --config /path/to/config.toml
```

### 3. Verify Installation

```bash
# Check version
neru --version

# Check status
neru status
```

Expected output:

```
Neru Status:
  Status: running
  Mode: idle
  Config: /Users/you/.config/neru/config.toml
```

---

## Configuration Setup

After installation, Neru looks for configuration in:

1. **~/.config/neru/config.toml** (XDG standard - recommended)
2. **~/Library/Application Support/neru/config.toml** (macOS convention)
3. Custom path via `--config` flag

**Start with default config:**

```bash
# Create config directory
mkdir -p ~/.config/neru

# Copy default config
curl -o ~/.config/neru/config.toml \
  https://raw.githubusercontent.com/y3owk1n/neru/main/configs/default-config.toml
```

See [CONFIGURATION.md](CONFIGURATION.md) for detailed configuration options.

---

## Shell Completions

Neru provides shell completions for bash, zsh, and fish.

### Bash

```bash
neru completion bash > /usr/local/etc/bash_completion.d/neru
```

### Zsh

```bash
neru completion zsh > "${fpath[1]}/_neru"
```

### Fish

```bash
neru completion fish > ~/.config/fish/completions/neru.fish
```

---

## Troubleshooting

### "Neru wants to control this computer using accessibility features"

This is normal. Click **OK** and grant permissions in System Settings.

### Command not found: neru

If using the CLI build, ensure the binary is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="/usr/local/bin:$PATH"
```

### Permission denied

Make the binary executable:

```bash
chmod +x /usr/local/bin/neru
```

### App won't open (macOS quarantine)

macOS may quarantine apps from unidentified developers:

```bash
xattr -cr /Applications/Neru.app
```

Then try opening again.

### Nix build fails

Ensure you're on an Apple Silicon Mac (arm64). For Intel Macs, change the URL to:

```nix
url = "https://github.com/y3owk1n/neru/releases/download/v${version}/neru-darwin-amd64.zip";
```

---

## Uninstallation

### Homebrew

```bash
brew uninstall --cask neru
```

### Manual

```bash
# Remove app bundle
rm -rf /Applications/Neru.app

# Remove CLI
rm /usr/local/bin/neru

# Remove configuration
rm -rf ~/.config/neru
rm -rf ~/Library/Application\ Support/neru

# Remove logs
rm -rf ~/Library/Logs/neru
```

### Nix

Remove the module from your configuration and rebuild.

---

## Next Steps

- Read [CONFIGURATION.md](CONFIGURATION.md) to customize Neru
- Check [CLI.md](CLI.md) for command-line usage
- Review [TROUBLESHOOTING.md](TROUBLESHOOTING.md) if you encounter issues
