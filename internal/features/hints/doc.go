// Package hints provides accessibility tree-based hint generation and navigation.
//
// This package implements a hint-based navigation system that queries the
// accessibility tree to find clickable elements and overlays them with
// keyboard hints (similar to Vimium for browsers).
//
// Key Features:
//   - Hint Generation: Create unique keyboard hints for UI elements
//   - Element Detection: Query accessibility tree for clickable elements
//   - Visual Overlays: Render hints over UI elements
//   - Input Routing: Match user input to hints and trigger actions
//   - Style Customization: Support for different hint appearance modes
//
// The hints system works well for applications with proper accessibility
// support but may have limitations with Electron apps or other applications
// that don't expose complete accessibility information. For universal
// coverage, use the grid package instead.
//
// Hint styles supported:
//   - Vimium-style: Sequential character hints (aa, ab, ac...)
//   - Custom characters: User-defined hint character sets
package hints
