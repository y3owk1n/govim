// Package hotkeys provides comprehensive functionality for registering and handling global hotkeys
// in the Neru application using the Carbon Event Manager API, enabling system-wide keyboard shortcuts.
//
// This package implements a complete hotkey management system that allows users to define and
// trigger global keyboard shortcuts for various Neru functions. It serves as the primary interface
// between user input and application functionality, translating keyboard combinations into actions.
//
// Key Features:
//   - Global Hotkey Registration: Register system-wide keyboard shortcuts
//   - Hotkey Unregistration: Clean removal of hotkey bindings
//   - Event Dispatching: Route hotkey events to appropriate handlers
//   - Modifier Support: Full support for shift, control, alt, and command keys
//   - Conflict Detection: Identify and handle hotkey conflicts
//   - Configuration Integration: Load hotkey mappings from user configuration
//
// The hotkey system works by:
//  1. Parsing user-defined hotkey configurations
//  2. Registering hotkeys with the macOS Carbon Event Manager
//  3. Monitoring for hotkey activation events
//  4. Dispatching events to registered callbacks
//  5. Managing hotkey lifecycle and cleanup
//
// Hotkey Format:
// Hotkeys are defined using a string format that specifies modifiers and keys:
//   - Modifiers: Cmd, Ctrl, Alt/Option, Shift (combined with +)
//   - Keys: Standard keyboard keys (Space, A-Z, 0-9, Function keys, etc.)
//   - Examples: "Cmd+Shift+Space", "Ctrl+D", "Alt+Tab"
//
// Supported Actions:
// Hotkeys can trigger various application actions:
//   - Mode Activation: hints, grid, scroll
//   - Action Execution: left_click, right_click, middle_click, mouse_down, mouse_up
//   - System Control: start, stop, idle
//   - Shell Commands: Execute external commands with "exec" prefix
//
// Configuration Integration:
// The package integrates with the configuration system to:
//   - Load hotkey mappings from config files
//   - Validate hotkey syntax and format
//   - Apply user-defined hotkey customizations
//   - Handle hotkey enable/disable states
//
// Performance Considerations:
// The hotkey system is designed for efficiency:
//   - Minimal overhead in the system event stream
//   - Efficient hotkey lookup and matching
//   - Proper resource cleanup to prevent leaks
//   - Asynchronous callback execution
//
// Security and Permissions:
// Hotkey registration requires accessibility permissions:
//   - Users must grant accessibility permissions for global hotkeys
//   - Clear error messaging for permission issues
//   - Graceful degradation when permissions are denied
//   - System prompts for first-time setup
//
// Error Handling:
// The package provides comprehensive error handling:
//   - Invalid hotkey format detection and reporting
//   - Registration failure with detailed error information
//   - Conflict resolution with user guidance
//   - System-level error recovery and retry logic
//
// Thread Safety:
// The hotkey system ensures thread-safe operation:
//   - Serialized hotkey registration and unregistration
//   - Safe callback execution across goroutines
//   - Proper locking for shared hotkey state
//   - Atomic operations for critical sections
//
// Integration Points:
// The hotkey package integrates with:
//   - Event tap system for low-level event monitoring
//   - Main application for action execution
//   - Configuration system for user preferences
//   - Error handling for system-level issues
package hotkeys
