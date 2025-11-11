# CLI Usage

Neru provides a comprehensive command-line interface for controlling the daemon and automating workflows.

## Overview

The Neru CLI communicates with the daemon via IPC (Inter-Process Communication) using Unix domain sockets at `/tmp/neru.sock`. This enables:

- Fast, reliable communication
- Multiple concurrent CLI commands
- Proper error handling
- Scriptable workflows

---

## Table of Contents

- [Daemon Management](#daemon-management)
- [Hint Commands](#hint-commands)
- [Status and Info](#status-and-info)
- [Shell Completions](#shell-completions)
- [Scripting Examples](#scripting-examples)
- [IPC Details](#ipc-details)

---

## Daemon Management

### Launch

Start the Neru daemon:

```bash
# Use default config location
neru launch

# Use custom config
neru launch --config /path/to/config.toml
```

**Flags:**

- `--config` - Path to config file

### Start

Resume a paused daemon (does not launch if not running):

```bash
neru start
```

### Stop

Pause the daemon (disables functionality but doesn't quit):

```bash
neru stop
```

**Use cases:**

- Temporarily disable Neru without quitting
- Prevent accidental hint activation during presentations
- Quick toggle via script

### Idle

Return to idle state (cancel any active mode):

```bash
neru idle
```

**Use cases:**

- Cancel hint / grid mode programmatically
- Reset state in scripts
- Force exit from stuck modes

---

## Hint / Grid Commands

### General Syntax

```bash
neru hints [action]
neru grid [action]
```

**Available actions:**

- `left_click` - Show hints and left click
- `right_click` - Show hints and right click
- `double_click` - Show hints and double click
- `triple_click` - Show hints and triple click
- `middle_click` - Show hints and middle click (useful for opening links in new tabs)
- `move_mouse` - Show hints and move cursor (no click)
- `mouse_down` - Show hints and hold mouse button (for dragging)
- `mouse_up` - Release mouse button (typically called after mouse_down)
- `scroll` - Enter scroll mode
- `context_menu` - Show hints and open context menu (action selection)

### Examples

```bash
# Left click (most common)
neru hints left_click
neru grid left_click

# Right click for context menu
neru hints right_click
neru grid right_click

# Double click (e.g., open files in Finder)
neru hints double_click
neru grid double_click

# Middle click (e.g., open link in new tab)
neru hints middle_click
neru grid middle_click

# Move mouse without clicking
neru hints move_mouse
neru grid move_mouse

# Enter scroll mode
neru hints scroll
neru grid scroll

# Context menu (choose action after selecting hint)
neru hints context_menu
neru grid context_menu
```

### Context Menu Actions

When using `context_menu`, after selecting a hint you can choose:

- `l` - Left click (with cursor restore if enabled)
- `r` - Right click (with cursor restore if enabled)
- `d` - Double click (with cursor restore if enabled)
- `t` - Triple click (with cursor restore if enabled)
- `m` - Middle click (with cursor restore if enabled)
- `g` - Go to position (move cursor)
- `h` - Hold mouse (start drag)
- `H` - Release mouse (end drag)

---

## Status and Info

### Status

Check daemon status and current mode:

```bash
neru status
```

**Example output:**

```
Neru Status:
  Status: running
  Mode: idle
  Config: /Users/you/.config/neru/config.toml
```

**Possible statuses:**

- `running` - Daemon active and responsive
- `paused` - Daemon paused via `neru stop`

**Possible modes:**

- `idle` - No active hint mode
- `hints` - Hint mode active
- `scroll` - Scroll mode active

### Version

```bash
neru --version
```

### Help

```bash
# General help
neru --help

# Command-specific help
neru hints --help
neru grid --help
neru launch --help
```

---

## Shell Completions

Generate shell completions for your shell:

### Bash

```bash
# Generate completion
neru completion bash > /usr/local/etc/bash_completion.d/neru

# Add to ~/.bashrc
source /usr/local/etc/bash_completion.d/neru
```

### Zsh

```bash
# Generate completion
neru completion zsh > "${fpath[1]}/_neru"

# Reload completions
rm -f ~/.zcompdump
compinit
```

### Fish

```bash
# Generate completion
neru completion fish > ~/.config/fish/completions/neru.fish
```

---

## Scripting Examples

### Toggle Neru

```bash
#!/bin/bash
# toggle-neru.sh - Toggle Neru on/off

STATUS=$(neru status | grep "Status:" | awk '{print $2}')

if [ "$STATUS" = "running" ]; then
    echo "Pausing Neru..."
    neru stop
else
    echo "Resuming Neru..."
    neru start
fi
```

### Hotkey Integration (skhd)

Instead of Neru's built-in hotkeys, use external hotkey managers:

```bash
# ~/.config/skhd/skhdrc

# Neru hotkeys
ctrl - f : neru hints left_click
ctrl - g : neru hints context_menu
ctrl - s : neru hints scroll
ctrl - r : neru hints right_click
ctrl - d : neru hints double_click

# Toggle Neru
ctrl + alt - n : ~/scripts/toggle-neru.sh
```

### Check if Running

```bash
#!/bin/bash
# check-neru.sh

if neru status &>/dev/null; then
    echo "Neru is running"
    exit 0
else
    echo "Neru is not running"
    exit 1
fi
```

### Alfred Workflow

Create an Alfred workflow for quick access:

```bash
# Alfred Script Filter
neru hints left_click
```

Trigger: `nerul` (Neru Left click)
Action: Run script above

---

## IPC Details

### Socket Location

Unix domain socket: `/tmp/neru.sock`

### Communication Protocol

The CLI and daemon communicate via JSON messages over the Unix socket:

**Request format:**

```json
{
  "command": "hints",
  "action": "left_click"
}
```

**Response format:**

```json
{
  "success": true,
  "message": "Hints activated"
}
```

### Error Handling

When the daemon is not running:

```bash
$ neru hints left_click
Error: Failed to connect to Neru daemon
Is Neru running? Try: neru launch
```

When a mode is disabled:

```bash
$ neru hints left_click
Error: hints mode is disabled
```

### Rate Limiting

There's no artificial rate limiting on IPC commands. You can send commands as fast as your scripts need.

### Concurrency

Multiple CLI commands can run concurrently. The daemon handles them sequentially in the order received.

### Log Monitoring

```bash
# Real-time log monitoring
tail -f ~/Library/Logs/neru/app.log

# Search for errors
grep ERROR ~/Library/Logs/neru/app.log

# Watch for specific events
tail -f ~/Library/Logs/neru/app.log | grep "hints"
```

---

## Troubleshooting

### Command hangs

If a command hangs, the daemon may be stuck:

```bash
# Force quit daemon
pkill -9 neru

# Restart
neru launch
```

### Socket permission errors

If you get permission errors on `/tmp/neru.sock`:

```bash
# Remove stale socket
rm -f /tmp/neru.sock

# Restart daemon
neru launch
```

### Commands not working

Verify daemon is running:

```bash
neru status
```

If not running:

```bash
neru launch
```

---

## Next Steps

- See [CONFIGURATION.md](CONFIGURATION.md) to customize hotkeys and behavior
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
- Review [DEVELOPMENT.md](DEVELOPMENT.md) to contribute
