// Package appwatcher provides comprehensive functionality for monitoring application lifecycle events
// on macOS, enabling Neru to respond to application launches, terminations, activations, and deactivations.
//
// This package implements a robust application watcher system that tracks the lifecycle of macOS
// applications in real-time. It serves as a critical component for maintaining context-aware
// behavior in Neru, allowing the application to adapt its functionality based on the currently
// active application.
//
// Key Features:
//   - Real-time Monitoring: Track application launches, terminations, activations, and deactivations
//   - Bundle ID Tracking: Identify applications by their unique bundle identifiers
//   - Event Callbacks: Register handlers for specific application lifecycle events
//   - Performance Optimization: Efficient event processing with minimal system overhead
//   - Error Resilience: Graceful handling of system events and edge cases
//
// The app watcher system works by registering for macOS notifications about application
// lifecycle events and dispatching them to registered callbacks. This enables Neru to:
//   - Adjust behavior based on the active application
//   - Apply application-specific configurations
//   - Handle excluded applications appropriately
//   - Trigger accessibility support for specific app types
//
// Event Types:
//   - App Launch: Triggered when an application starts
//   - App Terminate: Triggered when an application quits
//   - App Activate: Triggered when an application becomes active (foreground)
//   - App Deactivate: Triggered when an application loses focus (background)
//
// Integration Points:
// The app watcher integrates with several key components of Neru:
//   - Configuration system for app-specific settings and exclusions
//   - Accessibility system for browser detection and enhanced support
//   - Hotkey system for application-aware hotkey management
//   - Main application for state updates and mode management
//
// Performance Considerations:
// The package is designed to be lightweight and efficient:
//   - Minimal overhead on system resources
//   - Efficient event queuing and processing
//   - Smart filtering to avoid unnecessary callbacks
//   - Proper cleanup of system resources
//
// Error Handling:
// The app watcher includes comprehensive error handling:
//   - Graceful degradation when system notifications fail
//   - Recovery from unexpected application states
//   - Logging of unusual events for debugging
//   - Fallback mechanisms for critical functionality
//
// Usage Patterns:
// The package is typically used to:
//   - Monitor for excluded applications and disable Neru accordingly
//   - Detect browser applications and enable enhanced accessibility support
//   - Track application context for mode-specific behavior
//   - Maintain accurate state of active applications
package appwatcher
