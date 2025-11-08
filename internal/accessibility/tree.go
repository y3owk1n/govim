package accessibility

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/accessibility.h"
#include <stdlib.h>

*/
import "C"

import (
	"image"
	"sync"
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
	ParallelThreshold  int
	MaxParallelDepth   int
}

// DefaultTreeOptions returns default tree traversal options
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		FilterFunc:         nil,
		IncludeOutOfBounds: false,
		Cache:              NewInfoCache(5 * time.Second),
		ParallelThreshold:  8, // Only parallelize if there are more than 8 elements
		MaxParallelDepth:   4, // Don't parallelize deeper than 4 levels
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

	// Decide whether to parallelize
	shouldParallelize := depth <= opts.MaxParallelDepth &&
		len(children) >= opts.ParallelThreshold

	if shouldParallelize {
		buildChildrenParallel(parent, children, depth, opts, windowBounds)
	} else {
		buildChildrenSequential(parent, children, depth, opts, windowBounds)
	}
}

func buildChildrenSequential(parent *TreeNode, children []*Element, depth int, opts TreeOptions, windowBounds image.Rectangle) {
	parent.Children = make([]*TreeNode, 0, len(children))

	for _, child := range children {
		// Try cache first
		info := opts.Cache.Get(child)
		if info == nil {
			var err error
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

func buildChildrenParallel(parent *TreeNode, children []*Element, depth int, opts TreeOptions, windowBounds image.Rectangle) {
	// Pre-allocate result slice
	type childResult struct {
		node  *TreeNode
		index int
	}

	results := make(chan childResult, len(children))
	var wg sync.WaitGroup

	// Process children in parallel
	for i, child := range children {
		wg.Add(1)
		go func(idx int, elem *Element) {
			defer wg.Done()

			// Try cache first (cache must be thread-safe!)
			info := opts.Cache.Get(elem)
			if info == nil {
				var err error
				info, err = elem.GetInfo()
				if err != nil {
					return
				}
				opts.Cache.Set(elem, info)
			}

			if !shouldIncludeElement(info, opts, windowBounds) {
				return
			}

			childNode := &TreeNode{
				Element:  elem,
				Info:     info,
				Parent:   parent,
				Children: []*TreeNode{},
			}

			// Recursively build (this may spawn more goroutines at deeper levels)
			buildTreeRecursive(childNode, depth+1, opts, windowBounds)

			results <- childResult{node: childNode, index: idx}
		}(i, child)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results while maintaining order
	collected := make([]*TreeNode, len(children))
	validCount := 0

	for result := range results {
		collected[result.index] = result.node
		validCount++
	}

	// Build final children slice in original order, skipping nils
	parent.Children = make([]*TreeNode, 0, validCount)
	for _, node := range collected {
		if node != nil {
			parent.Children = append(parent.Children, node)
		}
	}
}

// shouldIncludeElement combines all filtering logic into one function
func shouldIncludeElement(info *ElementInfo, opts TreeOptions, windowBounds image.Rectangle) bool {
	if !opts.IncludeOutOfBounds {
		elementRect := rectFromInfo(info)

		// Filter out zero-sized interactive elements (they're broken/invalid)
		if elementRect.Dx() == 0 || elementRect.Dy() == 0 {
			if interactiveLeafRoles[info.Role] {
				return false
			}
		}

		// For non-zero sized elements, check if they overlap with window bounds
		if elementRect.Dx() > 0 && elementRect.Dy() > 0 {
			if !elementRect.Overlaps(windowBounds) {
				return false
			}
		}
	}

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
