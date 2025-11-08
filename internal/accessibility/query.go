package accessibility

import (
	"fmt"
	"image"
	"sync"
	"time"
)

var (
	globalCache *InfoCache
	cacheOnce   sync.Once
)

func rectFromInfo(info *ElementInfo) image.Rectangle {
	return image.Rect(
		info.Position.X,
		info.Position.Y,
		info.Position.X+info.Size.X,
		info.Position.Y+info.Size.Y,
	)
}

// NOTE: This is a debugging function that prints the entire accessibility tree structure.
// Print the entire tree structure for debugging
func PrintTree(node *TreeNode, depth int) {
	if node == nil || node.Info == nil {
		return
	}
	indent := ""
	for range depth {
		indent += "  "
	}
	fmt.Printf("%sRole: %s, Title: %s, Size: %dx%d\n",
		indent, node.Info.Role, node.Info.Title, node.Info.Size.X, node.Info.Size.Y)

	for _, child := range node.Children {
		PrintTree(child, depth+1)
	}
}

// GetClickableElements returns all clickable elements in the frontmost window
func GetClickableElements() ([]*TreeNode, error) {
	cacheOnce.Do(func() {
		globalCache = NewInfoCache(5 * time.Second)
	})

	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	opts := DefaultTreeOptions()
	opts.Cache = globalCache

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return tree.FindClickableElements(), nil
}

// GetScrollableElements returns all scrollable elements in the frontmost window
func GetScrollableElements() ([]*TreeNode, error) {
	cacheOnce.Do(func() {
		globalCache = NewInfoCache(5 * time.Second)
	})

	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	opts := DefaultTreeOptions()
	opts.Cache = globalCache

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return tree.FindScrollableElements(), nil
}

// GetMenuBarClickableElements returns clickable elements from the focused app's menu bar
func GetMenuBarClickableElements() ([]*TreeNode, error) {
	cacheOnce.Do(func() {
		globalCache = NewInfoCache(5 * time.Second)
	})

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
	opts.Cache = globalCache

	tree, err := BuildTree(menubar, opts)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return []*TreeNode{}, nil
	}
	return tree.FindClickableElements(), nil
}

func GetClickableElementsFromBundleID(bundleID string) ([]*TreeNode, error) {
	cacheOnce.Do(func() {
		globalCache = NewInfoCache(5 * time.Second)
	})

	app := GetApplicationByBundleID(bundleID)
	if app == nil {
		return []*TreeNode{}, nil
	}
	defer app.Release()

	opts := DefaultTreeOptions()
	opts.Cache = globalCache
	opts.IncludeOutOfBounds = true

	tree, err := BuildTree(app, opts)
	if err != nil {
		return nil, err
	}
	if tree == nil {
		return []*TreeNode{}, nil
	}
	return tree.FindClickableElements(), nil
}
