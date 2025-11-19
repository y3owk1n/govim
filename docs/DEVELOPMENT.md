# Development Guide

Contributing to Neru: build instructions, architecture overview, and contribution guidelines.

---

## Table of Contents

- [Development Setup](#development-setup)
- [Building](#building)
- [Testing](#testing)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [Release Process](#release-process)

---

## Development Setup

### Prerequisites

- **Go 1.25+** - [Install Go](https://golang.org/dl/)
- **Xcode Command Line Tools** - `xcode-select --install`
- **Just** - Command runner ([Install](https://github.com/casey/just))

  ```bash
  brew install just
  ```

- **golangci-lint** - Linter ([Install](https://golangci-lint.run/usage/install/))

  ```bash
  brew install golangci-lint
  ```

### Clone Repository

```bash
git clone https://github.com/y3owk1n/neru.git
cd neru
```

### Verify Setup

```bash
# Check Go version
go version  # Should be 1.25+

# Check tools
just --version
golangci-lint --version

# List available commands
just --list
```

---

## Building

Neru uses [Just](https://github.com/casey/just) as a command runner (alternative to Make).

### Quick Start

```bash
# Development build
just build

# Run
./bin/neru launch
```

### Build Commands

```bash
# Development build (auto-detects version from git tags)
just build

# Release build (optimized, stripped)
just release

# Build app bundle (macOS .app)
just bundle

# Build with custom version
just build-version v1.0.0

# Clean build artifacts
just clean
```

### Manual Build

Without Just:

```bash
# Basic build
go build -o bin/neru ./cmd/neru

# With version info
VERSION=$(git describe --tags --always --dirty)
go build \
  -ldflags="-s -w -X github.com/y3owk1n/neru/internal/cli.Version=$VERSION" \
  -o bin/neru \
  ./cmd/neru

# For release (optimized)
go build \
  -ldflags="-s -w -X github.com/y3owk1n/neru/internal/cli.Version=$VERSION" \
  -trimpath \
  -o bin/neru \
  ./cmd/neru
```

### Build Flags Explained

- `-ldflags="-s -w"` - Strip debug info and symbol table (smaller binary)
- `-trimpath` - Remove file system paths from binary
- `-X pkg.Var=value` - Set string variable at build time (version injection)

---

## Testing

### Run Tests

```bash
# All tests
just test

# With race detection
just test-race

# Specific package
go test ./internal/config/...

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...
```

### Run Linter

```bash
# Run all linters
just lint

# Auto-fix issues
golangci-lint run --fix
```

### Test During Development

```bash
# Watch mode (requires entr or similar)
find . -name "*.go" | entr -r just test

# Quick iteration
just build && ./bin/neru launch --config test-config.toml
```

---

## Architecture

### Project Structure

```
neru/
├── cmd/neru/              # Main entry point
│   └── main.go
├── internal/              # Internal packages
│   ├── app/               # Main application orchestration
│   ├── cli/               # CLI commands (Cobra-based)
│   ├── config/            # TOML configuration parsing
│   ├── domain/            # Core domain models and constants
│   ├── errors/            # Error definitions
│   ├── features/          # Feature implementations
│   │   ├── action/        # Action mode implementation
│   │   ├── grid/          # Grid mode implementation
│   │   ├── hints/         # Hints mode implementation
│   │   └── scroll/        # Scroll mode implementation
│   ├── infra/             # Infrastructure components
│   │   ├── accessibility/ # Accessibility API wrappers
│   │   ├── appwatcher/    # App watcher with callbacks
│   │   ├── bridge/        # CGo/Objective-C bridges
│   │   ├── electron/      # Electron/Chromium/Firefox manager
│   │   ├── eventtap/      # Event tap management
│   │   ├── hotkeys/       # Global hotkey management
│   │   ├── ipc/           # IPC server/client for daemon control
│   │   └── logger/        # Logging infrastructure
│   └── ui/                # UI components
│       └── overlay/       # Overlay manager
├── configs/               # Default configuration files
├── demos/                 # Demonstration files
├── docs/                  # Documentation
├── resources/             # Resource files
├── scripts/               # Build and packaging scripts
└── justfile               # Build commands
```

### Key Technologies

- **Go** - Core application logic, CLI, configuration
- **CGo + Objective-C** - macOS Accessibility API integration
- **Cobra** - CLI framework
- **TOML** - Configuration format
- **Unix Sockets** - IPC communication

### Key Packages

#### `internal/features/hints`

Grid-based or Accessibility hint generation and overlay rendering. This is the core of Neru's navigation system.

**Responsibilities:**

- Generate grid of hint labels
- Overlay hints on screen
- Handle hint selection
- Dispatch click actions

**Why Accessibility based?**

- More familiar to vimium users
- Able to get accurate click coordinates from AX API

**Why grid-based?**

- Works everywhere (no accessibility tree dependency)
- Fast and reliable
- Simple to maintain

#### `internal/infra/accessibility`

CGo wrappers around macOS Accessibility APIs.

**Responsibilities:**

- Query UI elements
- Determine clickable/scrollable elements
- Perform clicks and scrolls
- Enable AX support for Electron/Chromium

#### `internal/infra/hotkeys`

Global hotkey registration and management.

**Responsibilities:**

- Register system-wide hotkeys
- Dispatch hotkey events to hint/scroll modes
- Handle modifier key combinations

#### `internal/infra/ipc`

Unix socket-based IPC for CLI ↔ daemon communication.

**Responsibilities:**

- Server: Accept connections from CLI
- Client: Send commands to daemon
- Protocol: JSON messages
- Error handling and timeouts
- Structured response codes (`code` field) for reliable scripting
- Per-connection correlation IDs (`req_id`) for log tracing

#### `internal/config`

TOML configuration parsing and validation.

**Responsibilities:**

- Load config from multiple locations
- Parse TOML into structs
- Validate configuration
- Provide defaults

#### `internal/features/grid`

Grid-based navigation implementation.

**Responsibilities:**

- Generate coordinate-based grid overlay
- Handle grid cell selection
- Manage subgrid precision layer
- Dispatch click actions

#### `internal/features/scroll`

Scroll mode implementation.

**Responsibilities:**

- Handle Vim-style scrolling keys (j/k, gg/G, Ctrl+D/U)
- Provide visual scroll feedback
- Manage scroll state

#### `internal/features/action`

Action mode implementation.

**Responsibilities:**

- Handle different click types (left, right, middle, etc.)
- Provide visual action feedback
- Manage action state

#### Cursor Position Restoration

**Overview:**
- Stores the initial cursor coordinates and screen bounds when entering a mode.
- Restores the cursor on exit if enabled via `general.restore_cursor_position`.
- Adjusts for screen resolution/origin changes by mapping the original position proportionally to the current active screen.

**Key points:**
- Config flag: `general.restore_cursor_position` (default `true`).
- Entry points: Mode activation functions capture once per activation.
- Exit path: Mode exit functions restore and clear state.

**Usage example (config API):**
```go
cfg := config.Global()
if cfg != nil && cfg.General.RestoreCursorPosition {
    // restoration is enabled
}
```

#### `internal/cli`

Cobra-based CLI commands.

**Responsibilities:**

- Parse command-line arguments
- Dispatch to appropriate handlers
- Format output
- Error messages

---

## Contributing

### Contribution Workflow

1. **Fork the repository**
2. **Create a feature branch**

   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make your changes**
   - Write clean, documented code
   - Follow existing code style
   - Add tests for new features
4. **Test thoroughly**

   ```bash
   just test && just lint
   ```

5. **Commit with conventional commit**

   ```bash
   git commit -m "feat: description of what it does"
   git commit -m "fix(scope): description of what it does"
   ```

6. **Push to your branch**

   ```bash
   git push origin feature/amazing-feature
   ```

7. **Open a Pull Request**
   - Describe what the PR does
   - Reference any related issues
   - Include screenshots/demos if applicable

### Code Style

- **Follow Go conventions** - Use `gofmt`, `goimports`
- **Keep it simple** - Prefer clarity over cleverness
- **Document exports** - Add godoc comments for public functions/types
- **Handle errors** - Don't ignore errors; wrap them with context
- **Use meaningful names** - `clickableElement` not `ce`

**Example:**

```go
// Good
func (h *HintManager) GenerateHints(elements []Element) ([]Hint, error) {
    if len(elements) == 0 {
        return nil, fmt.Errorf("no elements to generate hints for")
    }
    // ...
}

// Bad
func (h *HintManager) gen(e []Element) []Hint {
    // ...
}
```

### Testing Guidelines

- **Write tests for new features**
- **Test edge cases** - Empty inputs, nil values, boundary conditions
- **Use table-driven tests** where appropriate
- **Mock external dependencies** - Don't rely on system state

**Example table-driven test:**

```go
func TestParseHotkey(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Hotkey
        wantErr bool
    }{
        {"simple", "Cmd+Space", Hotkey{Mod: Cmd, Key: Space}, false},
        {"invalid", "Cmd-Space", Hotkey{}, true},
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseHotkey(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseHotkey() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseHotkey() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- **Update README.md** if changing user-facing behavior
- **Update docs/** for significant features
- **Add godoc comments** for exported functions
- **Include examples** in documentation

### Commit Messages

Use clear, descriptive commit messages:

**Good:**

```
Add support for custom hint characters

Users can now configure which characters are used for hint labels
via the hint_characters config option.
```

**Bad:**

```
fix bug
```

---

## Release Process

Releases are being handled via [Release Please](https://github.com/googleapis/release-please) automatically.

### Version Numbering

Neru uses semantic versioning: `vMAJOR.MINOR.PATCH`

- **MAJOR** - Breaking changes
- **MINOR** - New features (backward compatible)
- **PATCH** - Bug fixes

### Creating a Release

Creating a release is just as easy as merging the release please PR, and it will build and publish the binaries on github.

> [!NOTE]
> Homebrew version bump is in a separate repo, it will be updated separately.

---

## Development Tips

### Quick Iteration

```bash
# One-liner: build and run
just build && ./bin/neru launch --config test-config.toml

# Watch for changes (requires entr)
ls **/*.go | entr -r sh -c 'just build && ./bin/neru launch'
```

### Debugging

```bash
# Enable debug logging
[logging]
log_level = "debug"

# Run verbose output
./bin/neru launch

# Watch logs in real-time
tail -f ~/Library/Logs/neru/app.log
```

### Useful Go Commands

```bash
# Format code
gofmt -w .

# Organize imports
goimports -w .

# Check for suspicious constructs
go vet ./...

# List dependencies
go list -m all

# Update dependencies
go get -u ./...
go mod tidy
```

### Profiling

```bash
# CPU profile
go test -cpuprofile cpu.prof ./internal/hints
go tool pprof cpu.prof

# Memory profile
go test -memprofile mem.prof ./internal/hints
go tool pprof mem.prof
```

---

## Need Help?

- **Read existing code** - The codebase is well-structured
- **Check issues** - Someone may have asked the same question
- **Ask in discussions** - Open a discussion for questions
- **Open a draft PR** - Get early feedback on your approach

---

## Project Philosophy

### Keep It Simple

Neru intentionally avoids:

- GUI settings (config files are superior)
- Feature bloat (focus on core navigation)

When adding features, ask:

1. Does this align with keyboard-driven productivity?
2. Is this the simplest way to achieve the goal?
3. Will this complicate maintenance?

### Community-Driven

Neru thrives on community contributions:

- **PRs over issues** - Code speaks louder than feature requests
- **Best-effort maintenance** - No promises of 24/7 support
- **Collective ownership** - Everyone can improve Neru

Your contributions shape Neru's future!

---

## Resources

- **Go Documentation:** <https://golang.org/doc/>
- **macOS Accessibility API:** <https://developer.apple.com/documentation/applicationservices/ax_ui_element_ref>
- **TOML Spec:** <https://toml.io/>
- **Cobra CLI Framework:** <https://github.com/spf13/cobra>
- **Just Command Runner:** <https://github.com/casey/just>
