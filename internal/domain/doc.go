// Package domain defines the core domain models, constants, and shared types for the Neru application.
//
// This package contains the fundamental building blocks that define the Neru domain,
// including mode definitions, action types, and other shared constants and types that
// are used across multiple packages in the application.
//
// Key Components:
//   - Mode Definitions: Enumerated types for application modes (Idle, Hints, Grid, etc.)
//   - Action Types: Enumerated types for different action kinds (LeftClick, RightClick, etc.)
//   - Constants: Shared constants used throughout the application
//   - Core Types: Fundamental data structures that represent core concepts
//
// The domain package serves as the foundation for the entire application, ensuring
// consistency and type safety across all components. It deliberately has minimal
// dependencies to avoid circular imports and maintain clean architectural boundaries.
//
// Mode System:
// Neru uses a mode-based approach to manage different states:
//   - ModeIdle: No active navigation mode
//   - ModeHints: Hint-based navigation active
//   - ModeGrid: Grid-based navigation active
//   - ModeScroll: Scroll mode active
//   - ModeAction: Action mode active (accessed via Tab in other modes)
//
// Action System:
// The action system defines various mouse actions that can be performed:
//   - ActionLeftClick: Standard left mouse click
//   - ActionRightClick: Right mouse click (context menu)
//   - ActionMiddleClick: Middle mouse click
//   - ActionMouseDown: Press and hold mouse button
//   - ActionMouseUp: Release held mouse button
//
// This package is intentionally kept minimal and focused to serve as a stable
// contract between different parts of the application.
package domain
