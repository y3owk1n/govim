package accessibility

import (
	"fmt"
	"image"
)

func rectFromInfo(info *ElementInfo) image.Rectangle {
	return image.Rect(
		info.Position.X,
		info.Position.Y,
		info.Position.X+info.Size.X,
		info.Position.Y+info.Size.Y,
	)
}

func expandRectangle(rect image.Rectangle, padding int) image.Rectangle {
	return image.Rect(
		rect.Min.X-padding,
		rect.Min.Y-padding,
		rect.Max.X+padding,
		rect.Max.Y+padding,
	)
}

// GetClickableElements returns all clickable elements in the frontmost window
func GetClickableElements() ([]*TreeNode, error) {
	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	windowInfo, err := window.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get window info: %w", err)
	}

	opts := DefaultTreeOptions()
	// Increase depth for Electron/web apps which have deeply nested content
	opts.MaxDepth = 25
	visibleBounds := expandRectangle(rectFromInfo(windowInfo), 200)
	opts.FilterFunc = func(info *ElementInfo) bool {
		// Filter out very small elements
		if info.Size.X < 10 || info.Size.Y < 10 {
			return false
		}

		// Skip elements that are completely outside the visible bounds (with padding)
		elementRect := rectFromInfo(info)
		return elementRect.Overlaps(visibleBounds)
	}

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return tree.FindClickableElements(), nil
}

// GetScrollableElements returns all scrollable elements in the frontmost window
func GetScrollableElements() ([]*TreeNode, error) {
	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	windowInfo, err := window.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get window info: %w", err)
	}

	opts := DefaultTreeOptions()
	opts.MaxDepth = 5
	visibleBounds := expandRectangle(rectFromInfo(windowInfo), 200)
	opts.FilterFunc = func(info *ElementInfo) bool {
		// Allow only elements overlapping the visible window bounds
		elementRect := rectFromInfo(info)
		return elementRect.Overlaps(visibleBounds)
	}

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return tree.FindScrollableElements(), nil
}

// GetMenuBarClickableElements returns clickable elements from the focused app's menu bar
func GetMenuBarClickableElements() ([]*TreeNode, error) {
	app := GetFocusedApplication()
	if app == nil {
		return []*TreeNode{}, nil
	}
	defer app.Release()

	menubar := app.GetMenuBar()
	if menubar == nil {
		return []*TreeNode{}, nil
	}
	defer menubar.Release()

	opts := DefaultTreeOptions()
	opts.MaxDepth = 8
	// Filter out tiny elements
	opts.FilterFunc = func(info *ElementInfo) bool {
		if info.Size.X < 6 || info.Size.Y < 6 {
			return false
		}
		return true
	}

	tree, err := BuildTree(menubar, opts)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return []*TreeNode{}, nil
	}
	return tree.FindClickableElements(), nil
}

// GetDockClickableElements returns clickable elements from the Dock
func GetDockClickableElements() ([]*TreeNode, error) {
	dock := GetApplicationByBundleID("com.apple.dock")
	if dock == nil {
		return []*TreeNode{}, nil
	}
	defer dock.Release()

	opts := DefaultTreeOptions()
	opts.MaxDepth = 8
	// Filter out tiny elements
	opts.FilterFunc = func(info *ElementInfo) bool {
		if info.Size.X < 6 || info.Size.Y < 6 {
			return false
		}
		return true
	}

	tree, err := BuildTree(dock, opts)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return []*TreeNode{}, nil
	}
	return tree.FindClickableElements(), nil
}
