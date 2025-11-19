// Package coordinates provides coordinate conversion and transformation utilities for Neru.
//
// This package contains essential functions for converting between different coordinate
// systems used in Neru, particularly for mapping screen coordinates to window coordinates
// and vice versa. These conversions are crucial for accurate positioning of overlays
// and proper interaction with UI elements across different applications and display
// configurations.
//
// Key Functions:
//   - Screen to window coordinate conversion
//   - Window to screen coordinate conversion
//   - Coordinate normalization and scaling
//   - Display bounds and geometry calculations
//
// The coordinate system utilities ensure that Neru's overlays and interactions are
// accurately positioned regardless of:
//   - Multiple monitor setups
//   - Different screen resolutions and DPI settings
//   - Window positioning and scaling
//   - Application-specific coordinate systems
//
// These functions are used throughout Neru's UI rendering and interaction systems
// to maintain spatial accuracy across all navigation modes.
package coordinates
