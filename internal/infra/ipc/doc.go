// Package ipc provides comprehensive inter-process communication functionality for the Neru application,
// allowing external commands to control the daemon via Unix domain sockets with structured messaging.
//
// This package implements a robust IPC system that enables communication between the Neru CLI and
// the daemon process. It serves as the primary mechanism for external control, status reporting,
// and configuration management, ensuring reliable and secure communication between components.
//
// Key Features:
//   - Unix Domain Socket Communication: Fast, secure local IPC
//   - JSON Message Protocol: Structured request/response messaging
//   - Concurrent Connection Handling: Support multiple simultaneous clients
//   - Error Code System: Machine-readable error codes for programmatic handling
//   - Request Correlation: Per-connection request tracking for debugging
//   - Timeout Management: Automatic cleanup of stale connections
//
// The IPC system works by:
//  1. Creating a Unix domain socket server in the daemon
//  2. Accepting connections from CLI clients
//  3. Parsing JSON requests from clients
//  4. Routing requests to appropriate handlers
//  5. Formatting and sending JSON responses
//  6. Managing connection lifecycle and cleanup
//
// Message Format:
// The IPC protocol uses JSON for all communication with a consistent structure:
//
// Request Format:
//
//	{
//	  "action": "hints",           // Command to execute
//	  "params": {"key": "value"},  // Optional parameters
//	  "args": ["arg1", "arg2"],    // Optional positional arguments
//	  "req_id": "unique-id"        // Optional request identifier for tracking
//	}
//
// Response Format:
//
//	{
//	  "success": true,             // Whether the operation succeeded
//	  "message": "Hints activated",// Human-readable response message
//	  "code": "OK",                // Machine-readable error code
//	  "data": {"optional": "payload"} // Optional response data
//	}
//
// Supported Actions:
// The IPC system supports all major Neru functionality:
//   - Mode Control: hints, grid, scroll, action
//   - System Control: start, stop, idle, launch
//   - Status Reporting: status, config, version
//   - Configuration: config inspection
//
// Error Handling:
// The package provides comprehensive error handling with structured responses:
//   - Machine-readable error codes for programmatic handling
//   - Human-readable error messages for users
//   - Contextual error information for debugging
//   - Graceful degradation on partial failures
//
// Common Error Codes:
//   - ERR_DAEMON_NOT_RUNNING: Daemon is not running or not responding
//   - ERR_MODE_DISABLED: Requested mode is disabled in configuration
//   - ERR_UNKNOWN_COMMAND: Invalid command provided
//   - ERR_INVALID_PARAMS: Malformed or invalid parameters
//   - ERR_PERMISSION_DENIED: Insufficient permissions for operation
//
// Performance Considerations:
// The IPC system is designed for efficiency and responsiveness:
//   - Minimal serialization overhead with JSON
//   - Efficient connection handling with connection pooling
//   - Proper resource cleanup to prevent leaks
//   - Asynchronous processing where appropriate
//
// Security Considerations:
// The IPC system implements several security measures:
//   - Unix domain sockets limit communication to local machine
//   - File permissions control socket access
//   - Input validation prevents injection attacks
//   - Rate limiting prevents abuse
//
// Thread Safety:
// The IPC system ensures thread-safe operation:
//   - Serialized request processing per connection
//   - Safe concurrent handling of multiple connections
//   - Proper locking for shared server state
//   - Atomic operations for critical sections
//
// Integration Points:
// The IPC package integrates with:
//   - Main application for command execution
//   - CLI for client-side communication
//   - Error handling for structured error responses
//   - Logging system for request tracking and debugging
package ipc
