// Package modes implements the different operational modes for Neru.
//
// This package contains the core logic for managing and transitioning between
// different navigation modes in Neru, including Idle, Hints, Grid, Scroll, and
// Action modes. Each mode represents a distinct state with specific behaviors
// and user interactions.
//
// Mode Descriptions:
//   - Idle: Default state where Neru waits for activation
//   - Hints: Label-based navigation using accessibility information
//   - Grid: Coordinate-based navigation using visual grid overlays
//   - Scroll: Vim-style scrolling navigation
//   - Action: Contextual action execution (click, drag, etc.)
//
// The modes package provides:
//   - State transition management between different modes
//   - Mode-specific handlers and behaviors
//   - Validation logic for mode transitions
//   - Common functionality shared across modes
//
// Mode transitions follow a state machine approach with well-defined rules
// to ensure predictable behavior and proper resource cleanup when switching
// between modes.
package modes
