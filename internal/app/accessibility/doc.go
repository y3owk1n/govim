// Package accessibility provides accessibility service functionality for Neru.
//
// This package contains the core accessibility service implementation that
// interfaces with macOS accessibility APIs to provide system-level integration.
// The service is responsible for querying accessibility elements, managing
// accessibility permissions, and providing the foundation for hint-based
// navigation features.
//
// Key Responsibilities:
//   - Accessibility element querying and management
//   - Permission handling and validation
//   - System integration with macOS accessibility frameworks
//   - Providing reliable access to UI elements across applications
//
// The accessibility service acts as a bridge between Neru's hint-based navigation
// system and the underlying macOS accessibility infrastructure, ensuring
// consistent and reliable access to UI elements regardless of the application
// being used.
package accessibility
