// Package accessibility provides macOS accessibility API integration for Neru.
//
// This package wraps macOS accessibility APIs to provide functionality for:
//   - Querying UI element trees
//   - Performing mouse clicks and scrolls
//   - Detecting clickable elements
//   - Managing accessibility permissions
//
// The package handles both standard macOS accessibility queries and
// specialized support for Chromium-based applications and Firefox.
//
// Key Features:
//   - Element Tree Traversal: Navigate the accessibility hierarchy
//   - Click Actions: Perform left, right, middle clicks at specific points
//   - Scroll Operations: Programmatic scrolling with precise control
//   - Role-based Filtering: Identify clickable elements by accessibility role
//   - Caching: Improve performance through intelligent element caching
package accessibility
