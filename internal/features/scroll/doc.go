// Package scroll provides Vim-style scrolling functionality for the Neru application.
//
// This package implements a comprehensive scrolling system that allows users to navigate
// content using familiar Vim-style keybindings. The scroll functionality can be used
// both as a standalone mode and within other navigation modes (hints/grid) for precise
// content navigation.
//
// Key Features:
//   - Vim-style Navigation: Implements standard Vim keybindings for scrolling
//   - j/k: Scroll down/up by configured step amount
//   - h/l: Scroll left/right by configured step amount
//   - Ctrl+d/Ctrl+u: Scroll half-page down/up
//   - gg/G: Jump to top/bottom of scrollable area
//   - Visual Feedback: Provides visual highlighting of the current scroll area
//   - Configurable Steps: All scroll distances are configurable via the config system
//   - Mode Integration: Seamlessly integrates with hints and grid modes
//   - Context Awareness: Automatically detects and scrolls the appropriate element
//
// The scroll system works by identifying the currently focused or mouse-over element
// that supports scrolling and then applying the appropriate scroll actions. When used
// standalone, it scrolls at the current cursor position. When used within hints or
// grid modes, it scrolls at the selected element location.
//
// Configuration Options:
//   - scroll_step: Distance for j/k scrolling (default: 50 pixels)
//   - scroll_step_half: Distance for Ctrl+d/Ctrl+u (default: 500 pixels)
//   - scroll_step_full: Distance for gg/G (default: 1000000 pixels)
//   - highlight_scroll_area: Whether to visually highlight the scroll area
//   - highlight_color: Color of the scroll area highlight
//   - highlight_width: Width of the scroll area highlight border
package scroll
