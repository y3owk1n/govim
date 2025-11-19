package grid

// Context holds the state and context for grid mode operations.
type Context struct {
	GridInstance **Grid
	GridOverlay  **Overlay
	InActionMode bool
}
