// Package action provides action mode functionality for the Neru application.
//
// This package implements the action mode system that allows users to perform various
// mouse actions at the current cursor position. Action mode is typically accessed by
// pressing Tab while in hints or grid mode, providing a way to choose different click
// types without having to re-select an element.
//
// Key Features:
//   - Multiple Action Types: Supports various mouse action types
//   - Left click: Standard mouse click
//   - Right click: Context menu activation
//   - Middle click: Middle mouse button click (useful for opening links in new tabs)
//   - Mouse down: Press and hold mouse button
//   - Mouse up: Release held mouse button
//   - Visual Feedback: Provides visual highlighting to indicate current action mode
//   - Configurable Keys: All action keys are user-configurable via the config system
//   - Mode Integration: Seamlessly integrates with hints and grid navigation modes
//
// The action system works by maintaining the current cursor position and applying
// the selected action type when triggered. This allows users to:
//  1. Navigate to a location using hints or grid
//  2. Switch to action mode (Tab key)
//  3. Select the desired action type (l/r/m/i/u keys)
//  4. Execute the action at the current cursor position
//
// Configuration Options:
//   - highlight_color: Color of the action mode highlight
//   - highlight_width: Width of the action mode highlight border
//   - left_click_key: Key for left click action (default: "l")
//   - right_click_key: Key for right click action (default: "r")
//   - middle_click_key: Key for middle click action (default: "m")
//   - mouse_down_key: Key for mouse down action (default: "i")
//   - mouse_up_key: Key for mouse up action (default: "u")
//
// Action mode is particularly useful for complex interactions like drag-and-drop
// operations, where you need to press and hold the mouse button at one location
// and release it at another.
package action
