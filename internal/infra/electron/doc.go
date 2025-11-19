// Package electron provides enhanced accessibility support for Electron-based applications,
// enabling proper UI element detection in applications that don't properly expose their
// interface to the macOS accessibility API by default.
//
// This package implements a specialized solution for Electron, Chromium, and Firefox-based
// applications that suffer from accessibility limitations. Many modern applications built
// with these frameworks don't expose complete accessibility information, making standard
// accessibility-based navigation difficult or impossible.
//
// Key Features:
//   - Manual Accessibility Injection: Enable accessibility attributes for Electron apps
//   - Browser Detection: Automatic identification of Chromium/Firefox-based applications
//   - Bundle ID Management: Configurable support for custom Electron/Chromium applications
//   - Permission Handling: Manage accessibility permissions for enhanced support
//   - Performance Optimization: Efficient injection with minimal overhead
//
// The package works by enabling the AXManualAccessibility attribute for supported applications,
// which forces them to expose their UI elements to the macOS accessibility API. This technique
// is particularly effective for:
//   - Electron applications (VS Code, Slack, Discord, etc.)
//   - Chromium-based browsers (Chrome, Brave, Arc, etc.)
//   - Firefox-based browsers (Firefox, Zen, etc.)
//
// Implementation Details:
// The electron package uses the bridge infrastructure to communicate with macOS accessibility
// APIs and enable manual accessibility attributes for target applications. The process involves:
//  1. Detecting supported applications through bundle ID matching
//  2. Enabling AXManualAccessibility attribute for the application
//  3. Verifying that accessibility attributes are properly exposed
//  4. Providing feedback on success or failure
//
// Configuration Integration:
// The package integrates with the configuration system to allow users to:
//   - Enable or disable enhanced accessibility support
//   - Add custom bundle IDs for Electron/Chromium applications
//   - Configure behavior for different application types
//
// Important Considerations:
// Enhanced accessibility support comes with important trade-offs:
//   - Tiling Window Manager Conflicts: Enabling AXEnhancedUserInterface can interfere with
//     tiling window managers like yabai, Amethyst, or Rectangle
//   - System Resource Usage: May increase memory usage in target applications
//   - Application Stability: Rare cases of application instability have been reported
//
// Best Practices:
//   - Keep enhanced support disabled if using a tiling window manager
//   - Enable only for applications where standard hints don't work
//   - Test thoroughly with each target application
//   - Monitor system performance after enabling
//
// Error Handling:
// The package provides comprehensive error handling for common issues:
//   - Permission denied errors with guidance for resolution
//   - Application not found conditions with diagnostic information
//   - API communication failures with retry logic
//   - Compatibility issues with fallback strategies
//
// Integration Points:
// The electron package integrates with:
//   - Accessibility system for attribute management
//   - App watcher for application detection
//   - Configuration system for user preferences
//   - Main application for coordination and state management
package electron
