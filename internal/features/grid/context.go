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

// SetGridInstanceValue sets the value of the grid instance pointer.
func (c *Context) SetGridInstanceValue(gridInstance *Grid) {
	*c.GridInstance = gridInstance
}

// GetGridInstance returns the grid instance.
func (c *Context) GetGridInstance() **Grid {
	return c.GridInstance
}

// SetGridOverlay sets the grid overlay.
func (c *Context) SetGridOverlay(gridOverlay **Overlay) {
	c.GridOverlay = gridOverlay
}

// SetGridOverlayValue sets the value of the grid overlay pointer.
func (c *Context) SetGridOverlayValue(gridOverlay *Overlay) {
	*c.GridOverlay = gridOverlay
}

// GetGridOverlay returns the grid overlay.
func (c *Context) GetGridOverlay() **Overlay {
	return c.GridOverlay
}

// SetInActionMode sets whether grid mode is in action mode.
func (c *Context) SetInActionMode(inActionMode bool) {
	c.InActionMode = inActionMode
}

// GetInActionMode returns whether grid mode is in action mode.
func (c *Context) GetInActionMode() bool {
	return c.InActionMode
}

// Reset resets the grid context to its initial state.
func (c *Context) Reset() {
	c.GridInstance = nil
	c.GridOverlay = nil
	c.InActionMode = false
}
