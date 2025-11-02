package accessibility

import (
	"image"

	"github.com/y3owk1n/govim/internal/logger"
	"go.uber.org/zap"
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
}

// DefaultTreeOptions returns default tree traversal options
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		MaxDepth:           10,
		FilterFunc:         nil,
		IncludeOutOfBounds: false,
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

	totalChildren := len(children)

	// Smart sampling for large containers
	indicesToCheck := getIndicesToCheck(totalChildren, parent.Info.Role)

	// Print sampling stats for large containers
	if totalChildren > 50 {
		logger.Debug("Sampling",
			zap.String("role", parent.Info.Role),
			zap.Int("total_children", totalChildren),
			zap.Int("indices_to_check", len(indicesToCheck)),
			zap.Float64("percent", float64(len(indicesToCheck))/float64(totalChildren)*100),
		)
	}

	parent.Children = make([]*TreeNode, 0, len(indicesToCheck))

	checkedCount := 0
	addedCount := 0

	for _, idx := range indicesToCheck {
		if idx >= len(children) {
			continue
		}
		child := children[idx]
		info, err := child.GetInfo()
		if err != nil {
			continue
		}

		checkedCount++

		// Skip elements that are completely outside the window bounds
		// UNLESS IncludeOutOfBounds is true
		if !opts.IncludeOutOfBounds {
			elementRect := rectFromInfo(info)
			if !elementRect.Overlaps(windowBounds) {
				continue
			}
		}

		// Apply filter if provided
		if opts.FilterFunc != nil && !opts.FilterFunc(info) {
			continue
		}

		childNode := &TreeNode{
			Element:  child,
			Info:     info,
			Parent:   parent,
			Children: []*TreeNode{},
		}
		parent.Children = append(parent.Children, childNode)
		addedCount++
		buildTreeRecursive(childNode, depth+1, opts, windowBounds)
	}

	// Summary for this container
	if totalChildren > 50 {
		logger.Debug("Result: ",
			zap.Int("checked_count", checkedCount),
			zap.Int("added_count", addedCount),
			zap.Int("skipped_count", totalChildren-checkedCount),
		)
	}
}

// getIndicesToCheck returns which indices to check based on container type and size
func getIndicesToCheck(totalChildren int, role string) []int {
	// For non-list containers, check all children
	if role != "AXList" && role != "AXTable" && role != "AXOutline" {
		indices := make([]int, totalChildren)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// For small lists, check everything
	if totalChildren <= 50 {
		indices := make([]int, totalChildren)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// For very large lists (>1000), be MUCH more conservative
	if totalChildren > 1000 {
		indices := make([]int, 0, 50)

		// First 20 items
		for i := 0; i < 20 && i < totalChildren; i++ {
			indices = append(indices, i)
		}

		// Sample every 100th item in the middle (or every 5% of total, whichever is larger)
		step := max(100, totalChildren/20)
		for i := 20; i < totalChildren-20; i += step {
			indices = append(indices, i)
		}

		// Last 20 items
		start := max(totalChildren-20, 20)
		for i := start; i < totalChildren; i++ {
			indices = append(indices, i)
		}

		return indices
	}

	// For medium lists (50-1000), use the original strategy
	indices := make([]int, 0, 100)

	// First 30
	for i := 0; i < 30 && i < totalChildren; i++ {
		indices = append(indices, i)
	}

	// Sample middle (every 10th)
	for i := 30; i < totalChildren-20; i += 10 {
		indices = append(indices, i)
	}

	// Last 20
	start := max(totalChildren-20, 30)
	for i := start; i < totalChildren; i++ {
		indices = append(indices, i)
	}

	return indices
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
