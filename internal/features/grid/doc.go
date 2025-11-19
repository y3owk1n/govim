// Package grid provides grid-based navigation functionality for Neru.
//
// This package implements a visual grid overlay system that allows users
// to navigate and click anywhere on the screen using keyboard input.
// Unlike accessibility tree-based approaches, the grid system works
// universally across all applications by dividing the screen into cells.
//
// Key Features:
//   - Grid Generation: Create customizable grid layouts
//   - Subgrid Navigation: Refined targeting within selected cells
//   - Input Matching: Match user input to grid cells
//   - Visual Rendering: Draw grid overlays with customizable styles
//   - Manager Coordination: Handle grid lifecycle and state
//
// The grid approach is more reliable than accessibility-based hints
// because it doesn't depend on applications exposing proper accessibility
// information, making it work everywhere including Electron apps, menubar
// items, Mission Control, and the Dock.
package grid
