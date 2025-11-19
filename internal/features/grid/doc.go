// Package grid provides grid-based navigation functionality for Neru.
//
// This package implements a comprehensive visual grid overlay system that allows users
// to navigate and click anywhere on the screen using keyboard input. Unlike accessibility
// tree-based approaches, the grid system works universally across all applications by
// dividing the screen into cells, making it the most reliable navigation method in Neru.
//
// Key Features:
//   - Grid Generation: Create customizable grid layouts with optimal cell sizing
//   - Subgrid Navigation: Refined targeting within selected cells for precision
//   - Input Matching: Match user input to grid cells with efficient algorithms
//   - Visual Rendering: Draw grid overlays with customizable styles and appearance
//   - Manager Coordination: Handle grid lifecycle and state management
//   - Performance Optimization: Efficient rendering and input processing
//
// The grid system works by:
//  1. Dividing the screen into a coordinate-based grid of cells
//  2. Rendering visual overlays with unique character identifiers
//  3. Matching user input to select specific grid cells
//  4. Optionally activating a subgrid for precision targeting
//  5. Executing actions at the selected screen coordinates
//
// Grid Architecture:
// The package implements a two-level grid system:
//   - Main Grid: Coarse navigation covering the entire screen
//   - Subgrid: Fine-grained precision targeting within selected cells
//
// Configuration Options:
//   - characters: Customizable character set for main grid labels
//   - sublayer_keys: Character set for subgrid labels
//   - Visual appearance: Colors, fonts, border styles, opacity
//   - Behavior settings: subgrid_enabled, live_match_update, hide_unmatched
//
// Advantages:
// The grid approach is more reliable than accessibility-based hints because it doesn't
// depend on applications exposing proper accessibility information, making it work
// everywhere including:
//   - Electron apps that don't expose complete accessibility trees
//   - Menubar items and system UI elements
//   - Mission Control and workspace management interfaces
//   - Dock items and application switchers
//   - Full-screen applications and games
//
// Performance Characteristics:
// The grid system is designed for optimal performance with:
//   - Efficient cell calculation algorithms
//   - Minimal rendering overhead
//   - Fast input processing
//   - Intelligent caching of grid calculations
//
// Integration Points:
// The grid package integrates with:
//   - UI overlay system for visual rendering
//   - App state management for mode transitions
//   - Configuration system for customization
//   - Input handling for character matching
package grid
