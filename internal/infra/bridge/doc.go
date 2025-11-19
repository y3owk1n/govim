// Package bridge provides comprehensive Go bindings for Objective-C APIs used in the Neru application,
// enabling seamless integration with macOS system frameworks for accessibility, hotkeys, and overlay functionality.
//
// This package serves as the critical bridge between Go and macOS native APIs, implementing CGo wrappers
// around Objective-C code to provide access to system-level functionality that is not available through
// standard Go libraries. It is the foundation for all macOS-specific integration in Neru.
//
// Key Components:
//   - Accessibility Bridge: Interface with macOS accessibility APIs for element querying and actions
//   - Hotkey Bridge: Integration with Carbon Event Manager for global hotkey registration
//   - Overlay Bridge: Native overlay rendering using macOS windowing and graphics APIs
//   - Event Tap Bridge: Low-level input event capture and processing
//   - Application Bridge: Interface with macOS application management APIs
//
// The bridge package implements a clean separation between Go application logic and Objective-C
// system APIs. This design provides several benefits:
//   - Type safety through Go interfaces
//   - Memory management with proper CGo conventions
//   - Error handling with Go-style error returns
//   - Testability through mockable interfaces
//
// Core Functionality Areas:
//
// Accessibility Integration:
//   - Element tree traversal and querying
//   - Mouse action execution (click, scroll, drag)
//   - Attribute reading and property extraction
//   - Permission management and system prompts
//
// Hotkey Management:
//   - Global hotkey registration and unregistration
//   - Modifier key combination handling
//   - Event callback dispatching
//   - Conflict detection and resolution
//
// Overlay Rendering:
//   - Window creation and management
//   - Visual element drawing and updating
//   - Transparency and layering controls
//   - Performance-optimized rendering
//
// Event Processing:
//   - Low-level event tap installation
//   - Event filtering and routing
//   - Input monitoring and capture
//   - System event interception
//
// Memory Management:
// The package carefully manages memory between Go and Objective-C environments:
//   - Proper CGo memory allocation and deallocation
//   - Reference counting for Objective-C objects
//   - Garbage collection coordination
//   - Leak prevention through systematic cleanup
//
// Error Handling:
// The bridge implements robust error handling for system-level operations:
//   - Conversion of Objective-C errors to Go errors
//   - Graceful degradation on API failures
//   - Detailed error information for debugging
//   - Recovery mechanisms for transient failures
//
// Performance Considerations:
//   - Minimal overhead in Go/Objective-C transitions
//   - Efficient data marshaling between environments
//   - Caching of expensive operations where appropriate
//   - Asynchronous processing for non-blocking operations
//
// Thread Safety:
// The package ensures thread safety for concurrent access:
//   - Proper locking for shared resources
//   - Serial dispatch queues for Objective-C operations
//   - Safe callback mechanisms across language boundaries
//   - Atomic operations for state management
package bridge
