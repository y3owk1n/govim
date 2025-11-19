// Package hints provides accessibility tree-based hint generation and navigation.
//
// This package implements a comprehensive hint-based navigation system that queries the
// accessibility tree to find clickable elements and overlays them with keyboard hints
// (similar to Vimium for browsers). It serves as one of the two primary navigation
// methods in Neru, complementing the grid-based approach.
//
// Key Features:
//   - Hint Generation: Create unique keyboard hints for UI elements using configurable character sets
//   - Element Detection: Query accessibility tree for clickable elements with role-based filtering
//   - Visual Overlays: Render customizable hints over UI elements with configurable appearance
//   - Input Routing: Match user input to hints and trigger appropriate actions
//   - Style Customization: Support for different hint appearance modes and visual customization
//   - Performance Optimization: Intelligent caching and efficient tree traversal algorithms
//
// The hints system works by:
//  1. Querying the accessibility tree of the active application
//  2. Identifying clickable elements based on accessibility roles and attributes
//  3. Generating unique character sequences for each clickable element
//  4. Rendering visual overlays with these hints at element locations
//  5. Matching user input to select and activate elements
//
// Supported Accessibility Roles:
// The package includes comprehensive support for standard accessibility roles including
// AXButton, AXLink, AXTextField, AXCheckBox, and many others. It also supports app-specific
// role configurations through the configuration system.
//
// Configuration Options:
//   - hint_characters: Customizable character set for hint generation
//   - Visual appearance: Colors, fonts, border styles, opacity
//   - include_menubar_hints: Whether to show hints in menubar items
//   - include_dock_hints: Whether to show hints in Dock items
//   - clickable_roles: Configurable list of accessibility roles considered clickable
//   - [hints.additional_ax_support]: Enhanced support for Electron/Chromium/Firefox apps
//
// Limitations:
// The hints system works well for applications with proper accessibility support but
// may have limitations with Electron apps or other applications that don't expose
// complete accessibility information. For universal coverage, use the grid package instead.
//
// Hint Generation Algorithms:
//   - Vimium-style: Sequential character hints (aa, ab, ac...)
//   - Custom characters: User-defined hint character sets
//   - Optimized distribution: Efficient character distribution for faster typing
//
// Integration Points:
// The hints package integrates with:
//   - Accessibility infrastructure for element detection
//   - UI overlay system for visual rendering
//   - App state management for mode transitions
//   - Configuration system for customization
package hints
