// Package errors provides custom error types and utilities for the Neru application.
//
// This package centralizes error handling patterns and provides domain-specific error types
// with better context and error wrapping capabilities. It follows Go's best practices for
// error handling while adding structure and consistency to error management across the application.
//
// Key Features:
//   - Domain-Specific Errors: Custom error types for common Neru-specific error conditions
//   - Error Wrapping: Utilities for wrapping errors with additional context
//   - Error Codes: Machine-readable error codes for programmatic error handling
//   - Structured Errors: Errors with consistent structure for logging and debugging
//
// Error Categories:
//   - Configuration Errors: Issues with configuration loading or validation
//   - Accessibility Errors: Problems with macOS accessibility permissions or APIs
//   - IPC Errors: Issues with inter-process communication
//   - Mode Errors: Problems related to mode activation or management
//   - System Errors: Low-level system or OS-related issues
//
// Error Handling Patterns:
//   - Contextual Errors: Errors include relevant context information
//   - Wrapping: Errors can be wrapped to provide additional context
//   - Codes: Each error type has a unique code for programmatic identification
//   - Logging: Errors are designed to be informative in logs
//
// The package provides both generic error utilities and specific error types for
// common error conditions in the Neru application. This approach ensures consistent
// error handling while maintaining the flexibility to handle specific error cases
// appropriately.
//
// Example Usage:
//
//	err := errors.New("failed to load config", errors.ConfigError)
//	if err != nil {
//	    log.Error("Config load failed", zap.Error(err))
//	    return err
//	}
//
// Error codes follow a consistent naming convention and are documented to enable
// reliable programmatic error handling in both the application and external tools.
package errors
