// Package components provides component wrapper types for Neru's feature modules.
//
// This package contains wrapper structs that encapsulate the various feature
// components used throughout the Neru application. These wrappers provide a
// unified interface for accessing and managing the different navigation modes
// and their associated functionality.
//
// Component Wrappers:
//   - HintsComponent: Encapsulates all hints-based navigation functionality
//   - GridComponent: Encapsulates all grid-based navigation functionality
//   - ScrollComponent: Encapsulates all scroll-based navigation functionality
//   - ActionComponent: Encapsulates all action-based functionality
//
// Each component wrapper bundles together the related managers, routers, contexts,
// and overlays needed for that specific feature, providing a clean separation
// of concerns and making it easier to manage the different aspects of Neru's
// navigation system.
package components
