// Package accessibility provides comprehensive macOS accessibility API integration for Neru.
//
// This package implements a complete wrapper around macOS accessibility APIs, providing
// functionality for querying UI element trees, performing mouse actions, detecting
// clickable elements, and managing accessibility permissions. It serves as the foundation
// for the hints-based navigation system.
//
// Key Features:
//   - Element Tree Traversal: Navigate the accessibility hierarchy with efficient algorithms
//   - Click Actions: Perform left, right, middle clicks at specific screen coordinates
//   - Scroll Operations: Programmatic scrolling with precise control and smooth animation
//   - Role-based Filtering: Identify clickable elements by accessibility role with customization
//   - Caching: Improve performance through intelligent element caching and invalidation
//   - Permission Management: Handle accessibility permissions and system prompts
//   - Enhanced Browser Support: Specialized support for Chromium and Firefox applications
//
// The package provides both low-level API bindings and high-level abstractions for common
// accessibility operations. It handles the complexity of macOS accessibility APIs while
// providing a clean, Go-friendly interface.
//
// Core Functionality:
//   - Element Querying: Find UI elements by role, attribute, or position
//   - Action Execution: Perform mouse clicks, scrolls, and other interactions
//   - Tree Navigation: Traverse parent/child/sibling relationships in the accessibility tree
//   - Attribute Reading: Extract text, position, size, and other element properties
//   - State Monitoring: Track element changes and application lifecycle events
//
// Browser Support:
// The package includes specialized support for modern web browsers:
//   - Chromium-based apps (Chrome, Brave, Arc, VS Code, Slack, etc.)
//   - Firefox-based apps (Firefox, Zen, etc.)
//   - Electron apps with enhanced accessibility attributes
//
// This support is implemented through manual accessibility attribute injection,
// which enables proper element detection in applications that don't expose complete
// accessibility information by default.
//
// Performance Considerations:
// The accessibility package implements several performance optimizations:
//   - Intelligent caching of element trees to avoid repeated queries
//   - Lazy evaluation of expensive operations
//   - Efficient filtering algorithms for element detection
//   - Memory management to prevent leaks
//
// Error Handling:
// The package provides comprehensive error handling for common accessibility issues:
//   - Permission denied errors with guidance for resolution
//   - Element not found conditions with fallback strategies
//   - API communication failures with retry logic
//   - Invalid element states with graceful degradation
//
// Integration Points:
// The accessibility package integrates with:
//   - Hints system for element detection and labeling
//   - App watcher for application lifecycle events
//   - Electron package for enhanced browser support
//   - Configuration system for role filtering and behavior customization
package accessibility
