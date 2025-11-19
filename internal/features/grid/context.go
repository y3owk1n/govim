package grid

// Context holds the state and context for grid mode operations.
type Context struct {
	GridInstance **Grid
	GridOverlay  **Overlay
	InActionMode bool
}

// SetGridInstance sets the grid instance.
func (c *Context) SetGridInstance(gridInstance **Grid) {
	c.GridInstance = gridInstance
}

// SetGridOverlay sets the grid overlay.
func (c *Context) SetGridOverlay(gridOverlay **Overlay) {
	c.GridOverlay = gridOverlay
}

// SetInActionMode sets whether grid mode is in action mode.
func (c *Context) SetInActionMode(inActionMode bool) {
	c.InActionMode = inActionMode
}

// Reset resets the grid context to its initial state.
func (c *Context) Reset() {
	c.GridInstance = nil
	c.GridOverlay = nil
	c.InActionMode = false
}
