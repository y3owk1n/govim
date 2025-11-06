package accessibility

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/accessibility.h"
#include <stdlib.h>

*/
import "C"

import (
	"image"
	"time"
)

// TreeNode represents a node in the accessibility tree
type TreeNode struct {
	Element  *Element
	Info     *ElementInfo
	Children []*TreeNode
	Parent   *TreeNode
}

// TreeOptions configures tree traversal
type TreeOptions struct {
	FilterFunc         func(*ElementInfo) bool
	IncludeOutOfBounds bool
	Cache              *InfoCache
}

// DefaultTreeOptions returns default tree traversal options
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		FilterFunc:         nil,
		IncludeOutOfBounds: false,
		Cache:              NewInfoCache(5 * time.Second),
	}
}

// BuildTree builds an accessibility tree from the given root element
func BuildTree(root *Element, opts TreeOptions) (*TreeNode, error) {
	if root == nil {
		return nil, nil
	}

	// Try to get from cache first
	info := opts.Cache.Get(root)
	if info == nil {
		var err error
		info, err = root.GetInfo()
		if err != nil {
			return nil, err
		}
		opts.Cache.Set(root, info)
	}

	// Calculate window bounds for spatial filtering
	windowBounds := rectFromInfo(info)
	// Add padding to catch elements slightly outside
	windowBounds = expandRectangle(windowBounds, 0)

	node := &TreeNode{
		Element: root,
		Info:    info,
	}

	buildTreeRecursive(node, 1, opts, windowBounds)

	return node, nil
}

// Roles that typically don't contain interactive elements
var nonInteractiveRoles = map[string]bool{
	"AXStaticText": true,
	"AXImage":      true,
	"AXHeading":    true,
}

// Roles that are themselves interactive (leaf nodes)
var interactiveLeafRoles = map[string]bool{
	"AXButton":             true,
	"AXComboBox":           true,
	"AXCheckBox":           true,
	"AXRadioButton":        true,
	"AXLink":               true,
	"AXPopUpButton":        true,
	"AXTextField":          true,
	"AXSlider":             true,
	"AXTabButton":          true,
	"AXSwitch":             true,
	"AXDisclosureTriangle": true,
	"AXTextArea":           true,
	"AXMenuButton":         true,
	"AXMenuItem":           true,
}

func buildTreeRecursive(parent *TreeNode, depth int, opts TreeOptions, windowBounds image.Rectangle) {
	// Early exit for roles that can't have interactive children
	if nonInteractiveRoles[parent.Info.Role] {
		return
	}

	// Don't traverse deeper into interactive leaf elements
	if interactiveLeafRoles[parent.Info.Role] {
		return
	}

	children, err := parent.Element.GetChildren()
	if err != nil || len(children) == 0 {
		return
	}

	parent.Children = make([]*TreeNode, 0, len(children))

	for _, child := range children {
		// Try cache first
		info := opts.Cache.Get(child)
		if info == nil {
			info, err = child.GetInfo()
			if err != nil {
				continue
			}
			opts.Cache.Set(child, info)
		}

		if !shouldIncludeElement(info, opts, windowBounds) {
			continue
		}

		childNode := &TreeNode{
			Element:  child,
			Info:     info,
			Parent:   parent,
			Children: []*TreeNode{},
		}

		parent.Children = append(parent.Children, childNode)
		buildTreeRecursive(childNode, depth+1, opts, windowBounds)
	}
}

// shouldIncludeElement combines all filtering logic into one function
func shouldIncludeElement(info *ElementInfo, opts TreeOptions, windowBounds image.Rectangle) bool {
	// Skip elements that are completely outside the window bounds
	// UNLESS IncludeOutOfBounds is true
	if !opts.IncludeOutOfBounds {
		elementRect := rectFromInfo(info)
		if !elementRect.Overlaps(windowBounds) {
			return false
		}
	}

	// Apply filter if provided
	if opts.FilterFunc != nil && !opts.FilterFunc(info) {
		return false
	}

	return true
}

// FindClickableElements finds all clickable elements in the tree
func (n *TreeNode) FindClickableElements() []*TreeNode {
	var result []*TreeNode
	n.walkTree(func(node *TreeNode) bool {
		if node.Element.IsClickable() {
			result = append(result, node)
		}
		return true
	})
	return result
}

// FindScrollableElements finds all scrollable elements in the tree
func (n *TreeNode) FindScrollableElements() []*TreeNode {
	var result []*TreeNode
	n.walkTree(func(node *TreeNode) bool {
		if node.Element.IsScrollable() {
			result = append(result, node)
		}
		return true
	})
	return result
}

// walkTree walks the tree and calls the visitor function for each node
func (n *TreeNode) walkTree(visitor func(*TreeNode) bool) {
	if !visitor(n) {
		return
	}

	for _, child := range n.Children {
		child.walkTree(visitor)
	}
}
