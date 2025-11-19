// Package state provides centralized state management for the Neru application.
//
// This package implements a structured approach to application state management,
// organizing state into focused, cohesive structures rather than scattering it
// across the main App struct. This improves maintainability, testability, and
// clarity of the application's state management.
//
// Key Features:
//   - Structured State Organization: State is grouped into logical components
//   - Type Safety: Strong typing ensures state consistency
//   - Encapsulation: State access is controlled through well-defined interfaces
//   - Testability: Centralized state makes unit testing easier
//   - Separation of Concerns: Different state aspects are managed independently
//
// State Components:
//   - App State: Core application state including current mode and status
//   - Cursor State: Cursor position and screen information for restoration
//   - Mode State: State specific to different navigation modes
//   - Configuration State: Current configuration values
//
// The state package is designed to be the single source of truth for application
// state, with other packages accessing state through well-defined interfaces
// rather than direct manipulation. This approach reduces coupling and makes
// the application more maintainable.
//
// State Management Patterns:
//   - Immutable Updates: State changes create new state instances rather than
//     modifying existing ones where appropriate
//   - Controlled Access: State is accessed through getter methods rather than
//     direct field access
//   - Validation: State changes are validated to ensure consistency
//
// This package plays a crucial role in maintaining the integrity and consistency
// of the application's state throughout its lifecycle.
package state
