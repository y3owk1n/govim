// Package ui implements the user interface components for the Neru application.
//
// This package provides the visual interface layer for Neru, managing all aspects
// of on-screen rendering and user interaction. It serves as the bridge between
// the application's core logic and the visual elements displayed to users.
//
// Key Components:
//   - Overlay Manager: Central coordinator for all visual overlays
//   - Renderer: Handles the actual drawing of visual elements
//   - Layout System: Manages positioning and sizing of UI elements
//   - Style System: Controls visual appearance and theming
//
// The UI package is designed to be lightweight and focused, delegating complex
// logic to feature-specific packages while providing a consistent interface
// for visual presentation. It integrates with macOS-native rendering APIs
// through the bridge package to ensure optimal performance and native look.
//
// Overlay System:
// Neru's UI is primarily based on overlays - temporary visual elements that
// appear on top of other applications. The overlay system supports:
//   - Hint Overlays: Character labels for clickable elements
//   - Grid Overlays: Coordinate-based navigation grid
//   - Action Overlays: Visual feedback for action mode
//   - Scroll Overlays: Visual indication of scroll areas
//
// Rendering Approach:
// The package uses a layered rendering approach where different types of
// overlays can coexist without conflict. Each overlay type has its own
// rendering logic while sharing common infrastructure for positioning
// and display management.
//
// The UI package is designed to be highly responsive and efficient,
// ensuring that visual feedback is immediate and doesn't interfere
// with the user's workflow.
package ui
