# GoVim Build System

# Default build
default: build

# Build the application
build:
    @echo "Building GoVim..."
    go build -o bin/govim cmd/govim/main.go
    @echo "✓ Build complete: bin/govim"

# Build with optimizations for release
release:
    @echo "Building release version..."
    go build -ldflags="-s -w" -o bin/govim cmd/govim/main.go
    @echo "✓ Release build complete: bin/govim"


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
    cp bin/govim /usr/local/bin/
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
    @echo "Running GoVim..."
    go run cmd/govim/main.go

# Run with verbose logging
run-debug:
    @echo "Running GoVim (debug mode)..."
    go run cmd/govim/main.go --log-level=debug

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
    @echo "GoVim Build Commands:"
    @echo ""
    @echo "  just build       - Build the application"
    @echo "  just release     - Build optimized release version"
    @echo "  just bundle      - Create macOS .app bundle"
    @echo "  just test        - Run tests"
    @echo "  just test-race   - Run tests with race detection"
    @echo "  just bench       - Run benchmarks"
    @echo "  just install     - Install to /usr/local/bin"
    @echo "  just clean       - Clean build artifacts"
    @echo "  just run         - Run the application"
    @echo "  just run-debug   - Run with debug logging"
    @echo "  just fmt         - Format code"
    @echo "  just lint        - Lint code"
    @echo "  just dist        - Create distribution package"
    @echo "  just deps        - Download dependencies"
    @echo "  just verify      - Verify dependencies"
