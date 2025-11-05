package accessibility

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/accessibility.h"
#include <stdlib.h>

*/
import "C"

import (
	"image"
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
	MaxDepth           int
	FilterFunc         func(*ElementInfo) bool
	IncludeOutOfBounds bool
	CheckOcclusion     bool
}

// DefaultTreeOptions returns default tree traversal options
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		MaxDepth:           10,
		FilterFunc:         nil,
		IncludeOutOfBounds: false,
		CheckOcclusion:     false,
	}
}

// BuildTree builds an accessibility tree from the given root element
func BuildTree(root *Element, opts TreeOptions) (*TreeNode, error) {
	if root == nil {
		return nil, nil
	}
	info, err := root.GetInfo()
	if err != nil {
		return nil, err
	}
	// Calculate window bounds for spatial filtering
	windowBounds := rectFromInfo(info)
	// Add padding to catch elements slightly outside
	windowBounds = expandRectangle(windowBounds, 0)

	node := &TreeNode{
		Element: root,
		Info:    info,
	}

	if opts.MaxDepth > 0 {
		buildTreeRecursive(node, 1, opts, windowBounds)
	}
	return node, nil
}

func buildTreeRecursive(parent *TreeNode, depth int, opts TreeOptions, windowBounds image.Rectangle) {
	if depth >= opts.MaxDepth {
		return
	}

	children, err := parent.Element.GetChildren()
	if err != nil || len(children) == 0 {
		return
	}

	parent.Children = make([]*TreeNode, 0, len(children))

	for _, child := range children {
		info, err := child.GetInfo()
		if err != nil {
			continue
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
