// Package logger provides comprehensive structured logging functionality for the Neru application,
// using the zap logging library with file rotation support to ensure reliable and efficient logging.
//
// This package implements a robust logging system that provides detailed insights into Neru's
// operation while maintaining high performance and minimal overhead. It serves as the foundation
// for debugging, monitoring, and auditing all aspects of the application.
//
// Key Features:
//   - Structured Logging: JSON-formatted logs for easy parsing and analysis
//   - Log Rotation: Automatic file rotation based on size and time limits
//   - Multiple Levels: Debug, info, warn, error severity levels
//   - Contextual Information: Rich context with structured fields
//   - Performance Optimized: High-performance logging with minimal overhead
//   - File and Console Output: Simultaneous logging to files and stdout/stderr
//
// The logging system works by:
//  1. Creating structured log entries with consistent formatting
//  2. Applying appropriate severity levels to different message types
//  3. Adding contextual information to aid debugging
//  4. Writing logs to both files and console output
//  5. Managing log file rotation to prevent disk space issues
//
// Log Levels:
// The package supports standard log levels with specific usage guidelines:
//   - Debug: Detailed information for diagnosing problems (development)
//   - Info: General information about application operation (production)
//   - Warn: Warning conditions that may indicate problems
//   - Error: Error conditions that prevent normal operation
//
// Structured Logging:
// All logs use structured JSON format with consistent fields:
//   - timestamp: ISO8601 formatted timestamp
//   - level: Log severity level
//   - logger: Logger name (typically package name)
//   - caller: File and line number of log call
//   - msg: Human-readable log message
//   - fields: Additional structured context data
//
// Configuration Integration:
// The package integrates with the configuration system to:
//   - Control log level verbosity
//   - Configure log file locations and naming
//   - Set rotation policies (size, backups, age)
//   - Enable or disable structured logging
//   - Toggle file logging on/off
//
// Performance Considerations:
// The logging system is designed for high performance:
//   - Minimal allocation overhead with pre-structured fields
//   - Efficient JSON serialization with zap
//   - Asynchronous logging where appropriate
//   - Conditional logging to avoid expensive operations
//
// File Rotation:
// The package implements comprehensive log rotation:
//   - Size-based rotation (default: 10MB per file)
//   - Backup retention (default: 5 backup files)
//   - Age-based cleanup (default: 30 days)
//   - Proper handling of rotation during active logging
//
// Error Handling:
// The logging system handles errors gracefully:
//   - Failures in file writing don't crash the application
//   - Fallback to console output when file logging fails
//   - Detailed error reporting for log system issues
//   - Recovery from disk space or permission issues
//
// Thread Safety:
// The logger ensures thread-safe operation:
//   - Concurrent access to logger instances
//   - Safe writing to log files from multiple goroutines
//   - Atomic operations for critical logging functions
//   - Proper synchronization for shared resources
//
// Integration Points:
// The logger package integrates with:
//   - Configuration system for runtime settings
//   - All other packages for application-wide logging
//   - Error handling for structured error reporting
//   - System monitoring for operational insights
package logger
