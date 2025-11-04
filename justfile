# Neru Build System
# Version information (can be overridden)

VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
GIT_COMMIT := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`
BUILD_DATE := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

# Ldflags for version injection

LDFLAGS := "-s -w -X github.com/y3owk1n/neru/internal/cli.Version=" + VERSION + " -X github.com/y3owk1n/neru/internal/cli.GitCommit=" + GIT_COMMIT + " -X github.com/y3owk1n/neru/internal/cli.BuildDate=" + BUILD_DATE

# Default build
default: build

# Build the application (development)
build:
    @echo "Building Neru..."
    @echo "Version: {{ VERSION }}"
    go build -ldflags="{{ LDFLAGS }}" -o bin/neru cmd/neru/main.go
    @echo "✓ Build complete: bin/neru"

# Build with optimizations for release
release:
    @echo "Building release version..."
    @echo "Version: {{ VERSION }}"
    @echo "Commit: {{ GIT_COMMIT }}"
    @echo "Date: {{ BUILD_DATE }}"
    go build -ldflags="{{ LDFLAGS }}" -trimpath -o bin/neru cmd/neru/main.go
    @echo "✓ Release build complete: bin/neru"

# Build with custom version
build-version VERSION_OVERRIDE:
    @echo "Building Neru with custom version..."
    go build -ldflags="-s -w -X github.com/y3owk1n/neru/internal/cli.Version={{ VERSION_OVERRIDE }} -X github.com/y3owk1n/neru/internal/cli.GitCommit={{ GIT_COMMIT }} -X github.com/y3owk1n/neru/internal/cli.BuildDate={{ BUILD_DATE }}" -trimpath -o bin/neru cmd/neru/main.go
    @echo "✓ Build complete: bin/neru (version: {{ VERSION_OVERRIDE }})"

# Run tests
test:
    @echo "Running tests..."
    go test -v ./...

# Run with race detection
test-race:
    @echo "Running tests with race detection..."
    go test -race -v ./...

# Run benchmarks
bench:
    @echo "Running benchmarks..."
    go test -bench=. -benchmem ./...

# Install locally
install: build
    @echo "Installing to /usr/local/bin..."
    cp bin/neru /usr/local/bin/
    @echo "✓ Installed successfully"

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf build/
    rm -rf *.app
    @echo "✓ Clean complete"

# Run the application (debug mode)
run:
    @echo "Running Neru..."
    go run cmd/neru/main.go

# Run with verbose logging
run-debug:
    @echo "Running Neru (debug mode)..."
    go run cmd/neru/main.go --log-level=debug

# Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...
    @echo "✓ Format complete"

# Lint code
lint:
    @echo "Linting code..."
    golangci-lint run
    @echo "✓ Lint complete"

# Generate code (if needed)
generate:
    @echo "Generating code..."
    go generate ./...
    @echo "✓ Generate complete"

# Download dependencies
deps:
    @echo "Downloading dependencies..."
    go mod download
    go mod tidy
    @echo "✓ Dependencies updated"

# Verify dependencies
verify:
    @echo "Verifying dependencies..."
    go mod verify
    @echo "✓ Dependencies verified"

# Show help
help:
    @echo "Neru Build Commands:"
    @echo ""
    @echo "  just build                    - Build the application with version info"
    @echo "  just release                  - Build optimized release version"
    @echo "  just build-version VERSION    - Build with custom version string"
    @echo "  just test                     - Run tests"
    @echo "  just test-race                - Run tests with race detection"
    @echo "  just bench                    - Run benchmarks"
    @echo "  just install                  - Install to /usr/local/bin"
    @echo "  just clean                    - Clean build artifacts"
    @echo "  just run                      - Run the application"
    @echo "  just run-debug                - Run with debug logging"
    @echo "  just fmt                      - Format code"
    @echo "  just lint                     - Lint code"
    @echo "  just deps                     - Download dependencies"
    @echo "  just verify                   - Verify dependencies"
    @echo ""
    @echo "Examples:"
    @echo "  just build-version v1.0.0     - Build with version v1.0.0"
    @echo "  just build-version local-dev  - Build with custom version tag"
