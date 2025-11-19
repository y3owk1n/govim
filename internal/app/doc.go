// Package app contains the main application logic for Neru.
//
// This package orchestrates all components of the Neru application including:
//   - Hotkey management and event handling
//   - Hints and grid mode coordination
//   - IPC server for external control
//   - Application lifecycle management
//   - Overlay rendering and management
//
// The App struct acts as the central coordinator, managing state transitions
// between different modes (Idle, Hints, Grid) and coordinating between
// various subsystems.
//
// Key Components:
//   - Mode Management: Handles switching between idle, hints, and grid modes
//   - Hotkey System: Registers and handles global hotkeys via eventtap
//   - IPC Handler: Processes commands from the CLI client
//   - Overlay Management: Controls visual overlays for hints and grid
//   - Accessibility: Integrates with macOS accessibility APIs
//
// The package uses dependency injection patterns for testability, with
// interfaces defined for external dependencies like hotkey services,
// event taps, and IPC servers.
package app
