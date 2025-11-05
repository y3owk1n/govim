package accessibility

import (
	"image"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/y3owk1n/neru/internal/logger"
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

// Cache for element info to avoid repeated GetInfo() calls
var (
	elementInfoCache sync.Map // map[uintptr]*ElementInfo
	cacheTTL         = 500 * time.Millisecond

	// Dynamic concurrency based on CPU cores
	maxConcurrency      = runtime.NumCPU() * 2
	maxChildConcurrency = runtime.NumCPU()

	// Object pools to reduce allocations
	treeNodePool = sync.Pool{
		New: func() any {
			return &TreeNode{Children: make([]*TreeNode, 0, 8)}
		},
	}
	childResultPool = sync.Pool{
		New: func() any {
			return &childResult{}
		},
	}
)

type cachedInfo struct {
	info      *ElementInfo
	timestamp time.Time
}

type childResult struct {
	node  *TreeNode
	index int
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

	node := getTreeNode()
	node.Element = root
	node.Info = info
	node.Parent = nil

	if opts.MaxDepth > 0 {
		buildTreeRecursive(node, 1, opts, windowBounds)
	}
	return node, nil
}

func buildTreeRecursive(parent *TreeNode, depth int, opts TreeOptions, windowBounds image.Rectangle) {
	if depth >= opts.MaxDepth {
		return
	}
	children, err := parent.Element.GetChildrenWithOcclusionCheck(true, opts.CheckOcclusion)
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

	// Early exit for small sets - no need for parallelization overhead
	if len(indicesToCheck) <= 3 {
		parent.Children = make([]*TreeNode, 0, len(indicesToCheck))
		for _, idx := range indicesToCheck {
			if idx >= len(children) {
				continue
			}
			child := children[idx]
			info, err := getCachedInfo(child)
			if err != nil {
				continue
			}

			if !shouldIncludeElement(info, opts, windowBounds) {
				continue
			}

			childNode := getTreeNode()
			childNode.Element = child
			childNode.Info = info
			childNode.Parent = parent

			parent.Children = append(parent.Children, childNode)
			buildTreeRecursive(childNode, depth+1, opts, windowBounds)
		}
		return
	}

	// Parallel processing for larger sets
	resultsChan := make(chan *childResult, len(indicesToCheck))
	var wg sync.WaitGroup

	// Limit concurrency based on CPU cores
	semaphore := make(chan struct{}, maxConcurrency)

	// Process children concurrently
	for _, idx := range indicesToCheck {
		if idx >= len(children) {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			child := children[idx]

			// Try to get cached info first
			info, err := getCachedInfo(child)
			if err != nil {
				return
			}

			// Combined filtering check
			if !shouldIncludeElement(info, opts, windowBounds) {
				return
			}

			childNode := getTreeNode()
			childNode.Element = child
			childNode.Info = info
			childNode.Parent = parent

			result := getChildResult()
			result.node = childNode
			result.index = idx

			resultsChan <- result
		}(idx)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and maintain order
	results := make(map[int]*TreeNode, len(indicesToCheck))
	for result := range resultsChan {
		results[result.index] = result.node
		putChildResult(result)
	}

	// Convert map to slice in order
	parent.Children = make([]*TreeNode, 0, len(results))
	for _, idx := range indicesToCheck {
		if node, ok := results[idx]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	// Summary for this container
	if totalChildren > 50 {
		logger.Debug("Result: ",
			zap.Int("checked_count", len(indicesToCheck)),
			zap.Int("added_count", len(parent.Children)),
			zap.Int("skipped_count", totalChildren-len(indicesToCheck)),
		)
	}

	// Recursively process children
	// For deep trees, process these in parallel too
	if len(parent.Children) > 5 && depth < opts.MaxDepth {
		var childWg sync.WaitGroup
		childSem := make(chan struct{}, maxChildConcurrency)

		for _, childNode := range parent.Children {
			childWg.Add(1)
			go func(cn *TreeNode) {
				defer childWg.Done()
				childSem <- struct{}{}
				defer func() { <-childSem }()

				buildTreeRecursive(cn, depth+1, opts, windowBounds)
			}(childNode)
		}
		childWg.Wait()
	} else {
		// For shallow trees or few children, process sequentially
		for _, childNode := range parent.Children {
			buildTreeRecursive(childNode, depth+1, opts, windowBounds)
		}
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

// getTreeNode gets a TreeNode from the pool
func getTreeNode() *TreeNode {
	node := treeNodePool.Get().(*TreeNode)
	// Reset the node
	node.Element = nil
	node.Info = nil
	node.Parent = nil
	node.Children = node.Children[:0] // Keep capacity, reset length
	return node
}

// collectNodes collects all nodes in the tree
func (n *TreeNode) collectNodes(result *[]*TreeNode) {
	if n == nil {
		return
	}
	*result = append(*result, n)
	for _, child := range n.Children {
		child.collectNodes(result)
	}
}

// getChildResult gets a childResult from the pool
func getChildResult() *childResult {
	return childResultPool.Get().(*childResult)
}

// putChildResult returns a childResult to the pool
func putChildResult(result *childResult) {
	if result != nil {
		result.node = nil
		result.index = 0
		childResultPool.Put(result)
	}
}

// getCachedInfo retrieves element info from cache or fetches it
func getCachedInfo(elem *Element) (*ElementInfo, error) {
	// Use element pointer address as cache key
	key := uintptr(unsafe.Pointer(elem))

	// Check cache
	if cached, ok := elementInfoCache.Load(key); ok {
		cachedData := cached.(*cachedInfo)
		// Check if cache entry is still valid
		if time.Since(cachedData.timestamp) < cacheTTL {
			return cachedData.info, nil
		}
	}

	// Fetch fresh info
	info, err := elem.GetInfo()
	if err != nil {
		return nil, err
	}

	// Store in cache
	elementInfoCache.Store(key, &cachedInfo{
		info:      info,
		timestamp: time.Now(),
	})

	return info, nil
}

// ClearElementInfoCache clears the element info cache
// Call this when you want to force fresh data
func ClearElementInfoCache() {
	elementInfoCache = sync.Map{}
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

// elementCheckFunc is a function that checks if an element meets certain criteria
type elementCheckFunc func(*Element) bool

// roleFilterFunc is a function that checks if a role should be included
type roleFilterFunc func(role string) bool

// findElementsByPredicate is a generic function to find elements matching a predicate
// with role pre-filtering and parallel checking
func (n *TreeNode) findElementsByPredicate(
	roleFilterFn roleFilterFunc,
	checkFunc elementCheckFunc,
) []*TreeNode {
	// Collect all nodes first (fast, no blocking)
	allNodes := make([]*TreeNode, 0, 1000)
	n.collectNodes(&allNodes)

	// Early filter by role BEFORE calling expensive ax checks
	candidateNodes := make([]*TreeNode, 0, len(allNodes)/10)
	for _, node := range allNodes {
		if roleFilterFn == nil || roleFilterFn(node.Info.Role) {
			candidateNodes = append(candidateNodes, node)
		}
	}

	// Now check in parallel on candidates only
	resultChan := make(chan *TreeNode, len(candidateNodes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency*3)

	for _, node := range candidateNodes {
		wg.Add(1)
		go func(n *TreeNode) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if checkFunc(n.Element) {
				resultChan <- n
			}
		}(node)
	}

	// Close channel when all checks complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	result := make([]*TreeNode, 0, len(candidateNodes)/3)
	for node := range resultChan {
		result = append(result, node)
	}

	return result
}

// FindClickableElements finds all clickable elements in the tree
func (n *TreeNode) FindClickableElements() []*TreeNode {
	return n.findElementsByPredicate(
		func(role string) bool {
			clickableRolesMu.RLock()
			_, ok := clickableRoles[role]
			clickableRolesMu.RUnlock()
			return ok
		},
		func(e *Element) bool { return e.IsClickable() },
	)
}

// FindScrollableElements finds all scrollable elements in the tree
func (n *TreeNode) FindScrollableElements() []*TreeNode {
	return n.findElementsByPredicate(
		func(role string) bool {
			scrollableRolesMu.RLock()
			_, ok := scrollableRoles[role]
			scrollableRolesMu.RUnlock()
			return ok
		},
		func(e *Element) bool { return e.IsScrollable() },
	)
}

// FindElementsByRole finds all elements with a specific role (no additional checks)
func (n *TreeNode) FindElementsByRole(role string) []*TreeNode {
	return n.findElementsByPredicate(
		func(r string) bool { return r == role },
		func(e *Element) bool { return true }, // All candidates pass
	)
}

// FindElementsByCustomCheck finds elements using a custom check function
// This is useful for ad-hoc queries without role filtering
func (n *TreeNode) FindElementsByCustomCheck(checkFunc elementCheckFunc) []*TreeNode {
	return n.findElementsByPredicate(
		nil, // No role filtering - all roles pass
		checkFunc,
	)
}
