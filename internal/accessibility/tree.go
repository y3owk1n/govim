package accessibility

import (
	"context"
	"sync"
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
	MaxDepth         int
	IncludeInvisible bool
	FilterFunc       func(*ElementInfo) bool
	MaxConcurrent    int
}

// DefaultTreeOptions returns default tree traversal options
func DefaultTreeOptions() TreeOptions {
	return TreeOptions{
		MaxDepth:      10,
		MaxConcurrent: 10,
		FilterFunc:    nil,
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

	node := &TreeNode{
		Element: root,
		Info:    info,
	}

	if opts.MaxDepth > 0 {
		buildTreeRecursive(node, 1, opts)
	}

	return node, nil
}

func buildTreeRecursive(parent *TreeNode, depth int, opts TreeOptions) {
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
		buildTreeRecursive(childNode, depth+1, opts)
	}
}

// BuildTreeConcurrent builds an accessibility tree using concurrent traversal
func BuildTreeConcurrent(root *Element, opts TreeOptions) (*TreeNode, error) {
	if root == nil {
		return nil, nil
	}

	info, err := root.GetInfo()
	if err != nil {
		return nil, err
	}

	node := &TreeNode{
		Element: root,
		Info:    info,
	}

	if opts.MaxDepth > 0 {
		ctx := context.Background()
		sem := make(chan struct{}, opts.MaxConcurrent)
		var wg sync.WaitGroup

		buildTreeConcurrentRecursive(ctx, node, 1, opts, sem, &wg)
		wg.Wait()
	}

	return node, nil
}

func buildTreeConcurrentRecursive(ctx context.Context, parent *TreeNode, depth int, opts TreeOptions, sem chan struct{}, wg *sync.WaitGroup) {
	if depth >= opts.MaxDepth {
		return
	}

	children, err := parent.Element.GetChildren()
	if err != nil || len(children) == 0 {
		return
	}

	parent.Children = make([]*TreeNode, 0, len(children))
	var mu sync.Mutex

	for _, child := range children {
		wg.Add(1)
		go func(child *Element) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			info, err := child.GetInfo()
			if err != nil {
				return
			}

			// Apply filter if provided
			if opts.FilterFunc != nil && !opts.FilterFunc(info) {
				return
			}

			childNode := &TreeNode{
				Element:  child,
				Info:     info,
				Parent:   parent,
				Children: []*TreeNode{},
			}

			mu.Lock()
			parent.Children = append(parent.Children, childNode)
			mu.Unlock()

			buildTreeConcurrentRecursive(ctx, childNode, depth+1, opts, sem, wg)
		}(child)
	}
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
