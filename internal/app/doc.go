// Package app contains the main application logic for Neru.
//
// This package orchestrates all components of the Neru application, serving as the
// central coordinator that manages state transitions, mode switching, and integration
// between various subsystems. The App struct is the heart of the application.
//
// Key Components:
//   - Mode Management: Handles switching between idle, hints, grid, scroll, and action modes
//   - Hotkey System: Registers and handles global hotkeys via eventtap
//   - IPC Handler: Processes commands from the CLI client through Unix domain sockets
//   - Overlay Management: Controls visual overlays for hints, grid, scroll, and action modes
//   - Accessibility Integration: Coordinates with accessibility APIs for hint-based navigation
//   - Lifecycle Management: Manages application startup, shutdown, and state persistence
//
// The App struct acts as the central coordinator, managing state transitions between
// different modes (Idle, Hints, Grid, Scroll, Action) and coordinating between various
// subsystems including:
//   - Hotkey registration and event handling
//   - Overlay rendering and management
//   - IPC command processing
//   - Configuration management
//   - Error handling and logging
//
// Mode Transitions:
// The application follows a state machine approach with well-defined mode transitions:
//   - Idle → Hints/Grid/Scroll/Action: Activation via hotkey or IPC command
//   - Hints/Grid/Scroll/Action → Idle: Completion, cancellation, or error
//   - Hints/Grid → Action: Tab key switches to action mode while maintaining position
//   - Action → Hints/Grid: Action completion returns to previous navigation mode
//
// The package uses dependency injection patterns for testability, with interfaces
// defined for external dependencies like hotkey services, event taps, and IPC servers.
// This design enables easier unit testing and mocking of external systems.
//
// Lifecycle Management:
// The application lifecycle is carefully managed to ensure proper resource cleanup,
// graceful shutdown, and consistent state management across restarts. Key lifecycle
// events include initialization, configuration loading, permission checking, and
// cleanup procedures.
package app
