package errors

import (
	"errors"
	"fmt"
)

// Common sentinel errors that can be used for error checking with errors.Is.
var (
	// ErrNotRunning indicates that the Neru daemon is not running.
	ErrNotRunning = errors.New("neru is not running")

	// ErrAlreadyRunning indicates that the Neru daemon is already running.
	ErrAlreadyRunning = errors.New("neru is already running")

	// ErrAccessibilityPermission indicates missing accessibility permissions.
	ErrAccessibilityPermission = errors.New("accessibility permissions not granted")

	// ErrInvalidConfig indicates configuration validation failure.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrModeDisabled indicates an attempt to use a disabled mode.
	ErrModeDisabled = errors.New("mode is disabled")

	// ErrInvalidInput indicates invalid user input.
	ErrInvalidInput = errors.New("invalid input")

	// ErrActionFailed indicates an action execution failure.
	ErrActionFailed = errors.New("action failed")

	// ErrUnknownCommand indicates an unknown IPC command.
	ErrUnknownCommand = errors.New("unknown command")

	// ErrTimeout indicates an operation timeout.
	ErrTimeout = errors.New("operation timed out")
)

// ConfigError represents a configuration-related error with additional context.
type ConfigError struct {
	Field   string // Configuration field that caused the error
	Value   string // The problematic value (if applicable)
	Message string // Error message
	Err     error  // Underlying error (if any)
}

// NewConfigError creates a new ConfigError with the specified field and message.
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}

// NewConfigErrorWithValue creates a new ConfigError with field, value, and message.
func NewConfigErrorWithValue(field, value, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// WrapConfigError wraps an existing error as a ConfigError.
func WrapConfigError(field string, err error) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: err.Error(),
		Err:     err,
	}
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		if e.Value != "" {
			return fmt.Sprintf(
				"config error in field '%s' (value: '%s'): %s",
				e.Field,
				e.Value,
				e.Message,
			)
		}
		return fmt.Sprintf("config error in field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// IPCError represents an IPC-related error.
type IPCError struct {
	Operation string // The IPC operation (e.g., "send", "connect", "decode")
	Message   string // Error message
	Err       error  // Underlying error
}

// NewIPCError creates a new IPCError.
func NewIPCError(operation, message string) *IPCError {
	return &IPCError{
		Operation: operation,
		Message:   message,
	}
}

// WrapIPCError wraps an existing error as an IPCError.
func WrapIPCError(operation string, err error) *IPCError {
	return &IPCError{
		Operation: operation,
		Message:   err.Error(),
		Err:       err,
	}
}

func (e *IPCError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("IPC %s error: %s", e.Operation, e.Message)
	}
	return fmt.Sprintf("IPC error: %s", e.Message)
}

func (e *IPCError) Unwrap() error {
	return e.Err
}

// AccessibilityError represents an accessibility-related error.
type AccessibilityError struct {
	Operation string // The accessibility operation (e.g., "click", "scroll", "query")
	Message   string // Error message
	Err       error  // Underlying error
}

// NewAccessibilityError creates a new AccessibilityError.
func NewAccessibilityError(operation, message string) *AccessibilityError {
	return &AccessibilityError{
		Operation: operation,
		Message:   message,
	}
}

// WrapAccessibilityError wraps an existing error as an AccessibilityError.
func WrapAccessibilityError(operation string, err error) *AccessibilityError {
	return &AccessibilityError{
		Operation: operation,
		Message:   err.Error(),
		Err:       err,
	}
}

func (e *AccessibilityError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("accessibility %s error: %s", e.Operation, e.Message)
	}
	return fmt.Sprintf("accessibility error: %s", e.Message)
}

func (e *AccessibilityError) Unwrap() error {
	return e.Err
}

// HotkeyError represents a hotkey-related error.
type HotkeyError struct {
	Hotkey  string // The hotkey string (e.g., "cmd+shift+space")
	Message string // Error message
	Err     error  // Underlying error
}

// NewHotkeyError creates a new HotkeyError.
func NewHotkeyError(hotkey, message string) *HotkeyError {
	return &HotkeyError{
		Hotkey:  hotkey,
		Message: message,
	}
}

// WrapHotkeyError wraps an existing error as a HotkeyError.
func WrapHotkeyError(hotkey string, err error) *HotkeyError {
	return &HotkeyError{
		Hotkey:  hotkey,
		Message: err.Error(),
		Err:     err,
	}
}

func (e *HotkeyError) Error() string {
	if e.Hotkey != "" {
		return fmt.Sprintf("hotkey '%s' error: %s", e.Hotkey, e.Message)
	}
	return fmt.Sprintf("hotkey error: %s", e.Message)
}

func (e *HotkeyError) Unwrap() error {
	return e.Err
}
