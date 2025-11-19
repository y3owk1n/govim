// Package eventtap provides comprehensive functionality for capturing and handling global keyboard events
// on macOS using the Carbon Event Manager API, enabling Neru to respond to system-wide input.
//
// This package implements a robust event tapping system that allows Neru to monitor and intercept
// keyboard events at the system level, regardless of which application is currently active. It
// serves as the foundation for global hotkey functionality and input processing in Neru.
//
// Key Features:
//   - Global Event Monitoring: Capture keyboard events system-wide
//   - Event Filtering: Selective processing of specific event types
//   - Callback Dispatching: Route events to appropriate handlers
//   - Performance Optimization: Efficient event processing with minimal overhead
//   - Error Resilience: Graceful handling of system event interruptions
//
// The event tap system works by installing a low-level event handler in the macOS event stream,
// allowing Neru to intercept and process keyboard events before they reach applications. This
// enables features like:
//   - Global hotkey activation for hints, grid, and scroll modes
//   - Input capture during active navigation modes
//   - Modifier key state tracking
//   - System-level input monitoring
//
// Event Processing Pipeline:
// The package implements a multi-stage event processing pipeline:
//  1. Event Capture: Low-level interception of keyboard events
//  2. Event Filtering: Determine which events require processing
//  3. Event Translation: Convert system events to internal representations
//  4. Callback Dispatch: Route events to registered handlers
//  5. Event Continuation: Allow or block event propagation
//
// Supported Event Types:
//   - Key Down Events: Capture when keys are pressed
//   - Key Up Events: Capture when keys are released
//   - Modifier Changes: Track shift, control, alt, and command key states
//   - System Hotkeys: Handle special system key combinations
//
// Performance Considerations:
// The event tap system is designed for maximum efficiency:
//   - Minimal processing overhead in the event stream
//   - Asynchronous callback execution to avoid blocking
//   - Intelligent event filtering to reduce unnecessary processing
//   - Proper resource management to prevent leaks
//
// Security and Permissions:
// Event tapping requires explicit user permission on macOS:
//   - Accessibility permissions must be granted for event monitoring
//   - System prompts users for permission on first use
//   - Graceful degradation when permissions are denied
//   - Clear error messaging for permission issues
//
// Error Handling:
// The package provides comprehensive error handling:
//   - Permission denied errors with user guidance
//   - System event stream interruptions with recovery
//   - Callback execution failures with isolation
//   - Resource allocation failures with cleanup
//
// Thread Safety:
// The event tap system ensures thread-safe operation:
//   - Serialized event processing to maintain order
//   - Safe callback execution across goroutines
//   - Proper locking for shared state management
//   - Atomic operations for critical sections
//
// Integration Points:
// The event tap package integrates with:
//   - Hotkey system for global hotkey registration
//   - Main application for mode management
//   - Input processing for active navigation modes
//   - Error handling for system-level issues
package eventtap
