package accessibility

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Query represents a query for UI elements
type Query struct {
	Role            string
	Title           string
	TitleContains   string
	Enabled         *bool
	Clickable       *bool
	Scrollable      *bool
	MinWidth        int
	MinHeight       int
	MaxResults      int
	Timeout         time.Duration
}

// QueryBuilder helps build queries
type QueryBuilder struct {
	query Query
}

// NewQuery creates a new query builder
func NewQuery() *QueryBuilder {
	return &QueryBuilder{
		query: Query{
			MaxResults: 100,
			Timeout:    5 * time.Second,
		},
	}
}

// WithRole sets the role filter
func (qb *QueryBuilder) WithRole(role string) *QueryBuilder {
	qb.query.Role = role
	return qb
}

// WithTitle sets the exact title filter
func (qb *QueryBuilder) WithTitle(title string) *QueryBuilder {
	qb.query.Title = title
	return qb
}

// WithTitleContains sets the title contains filter
func (qb *QueryBuilder) WithTitleContains(text string) *QueryBuilder {
	qb.query.TitleContains = text
	return qb
}

// WithEnabled sets the enabled filter
func (qb *QueryBuilder) WithEnabled(enabled bool) *QueryBuilder {
	qb.query.Enabled = &enabled
	return qb
}

// WithClickable sets the clickable filter
func (qb *QueryBuilder) WithClickable(clickable bool) *QueryBuilder {
	qb.query.Clickable = &clickable
	return qb
}

// WithScrollable sets the scrollable filter
func (qb *QueryBuilder) WithScrollable(scrollable bool) *QueryBuilder {
	qb.query.Scrollable = &scrollable
	return qb
}

// WithMinSize sets minimum size filter
func (qb *QueryBuilder) WithMinSize(width, height int) *QueryBuilder {
	qb.query.MinWidth = width
	qb.query.MinHeight = height
	return qb
}

// WithMaxResults sets the maximum number of results
func (qb *QueryBuilder) WithMaxResults(max int) *QueryBuilder {
	qb.query.MaxResults = max
	return qb
}

// WithTimeout sets the query timeout
func (qb *QueryBuilder) WithTimeout(timeout time.Duration) *QueryBuilder {
	qb.query.Timeout = timeout
	return qb
}

// Build returns the built query
func (qb *QueryBuilder) Build() Query {
	return qb.query
}

// Execute executes the query on the given tree
func (q Query) Execute(tree *TreeNode) ([]*TreeNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), q.Timeout)
	defer cancel()

	results := make([]*TreeNode, 0)
	resultsChan := make(chan *TreeNode, 100)
	done := make(chan struct{})

	// Start collection goroutine
	go func() {
		for node := range resultsChan {
			results = append(results, node)
			if q.MaxResults > 0 && len(results) >= q.MaxResults {
				break
			}
		}
		close(done)
	}()

	// Execute query
	q.executeRecursive(ctx, tree, resultsChan)
	close(resultsChan)

	// Wait for collection to finish
	select {
	case <-done:
	case <-ctx.Done():
		return nil, fmt.Errorf("query timeout")
	}

	return results, nil
}

func (q Query) executeRecursive(ctx context.Context, node *TreeNode, results chan<- *TreeNode) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if q.matches(node) {
		select {
		case results <- node:
		case <-ctx.Done():
			return
		}
	}

	for _, child := range node.Children {
		q.executeRecursive(ctx, child, results)
	}
}

func (q Query) matches(node *TreeNode) bool {
	info := node.Info

	// Role filter
	if q.Role != "" && info.Role != q.Role {
		return false
	}

	// Title filter
	if q.Title != "" && info.Title != q.Title {
		return false
	}

	// Title contains filter
	if q.TitleContains != "" && !strings.Contains(strings.ToLower(info.Title), strings.ToLower(q.TitleContains)) {
		return false
	}

	// Enabled filter
	if q.Enabled != nil && info.IsEnabled != *q.Enabled {
		return false
	}

	// Clickable filter
	if q.Clickable != nil && node.Element.IsClickable() != *q.Clickable {
		return false
	}

	// Scrollable filter
	if q.Scrollable != nil && node.Element.IsScrollable() != *q.Scrollable {
		return false
	}

	// Size filters
	if q.MinWidth > 0 && info.Size.X < q.MinWidth {
		return false
	}
	if q.MinHeight > 0 && info.Size.Y < q.MinHeight {
		return false
	}

	return true
}

// GetClickableElements returns all clickable elements in the frontmost window
func GetClickableElements() ([]*TreeNode, error) {
	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	opts := DefaultTreeOptions()
	opts.MaxDepth = 15
	opts.FilterFunc = func(info *ElementInfo) bool {
		// Filter out very small elements
		return info.Size.X >= 10 && info.Size.Y >= 10
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

	opts := DefaultTreeOptions()
	opts.MaxDepth = 15

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return tree.FindScrollableElements(), nil
}

// SearchElements searches for elements matching the query
func SearchElements(query Query) ([]*TreeNode, error) {
	window := GetFrontmostWindow()
	if window == nil {
		return nil, fmt.Errorf("no frontmost window")
	}
	defer window.Release()

	opts := DefaultTreeOptions()
	opts.MaxDepth = 15

	tree, err := BuildTree(window, opts)
	if err != nil {
		return nil, err
	}

	return query.Execute(tree)
}

// FuzzySearch performs a fuzzy search for elements by title
func FuzzySearch(searchText string, maxResults int) ([]*TreeNode, error) {
	query := NewQuery().
		WithTitleContains(searchText).
		WithEnabled(true).
		WithMaxResults(maxResults).
		Build()

	return SearchElements(query)
}
