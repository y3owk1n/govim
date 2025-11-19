package hints

// Context holds the state and context for hint mode operations.
type Context struct {
	SelectedHint *Hint
	InActionMode bool
}

// SetSelectedHint sets the currently selected hint.
func (c *Context) SetSelectedHint(hint *Hint) {
	c.SelectedHint = hint
}

// GetSelectedHint returns the currently selected hint.
func (c *Context) GetSelectedHint() *Hint {
	return c.SelectedHint
}

// SetInActionMode sets whether hint mode is in action mode.
func (c *Context) SetInActionMode(inActionMode bool) {
	c.InActionMode = inActionMode
}

// GetInActionMode returns whether hint mode is in action mode.
func (c *Context) GetInActionMode() bool {
	return c.InActionMode
}

// Reset resets the hints context to its initial state.
func (c *Context) Reset() {
	c.SelectedHint = nil
	c.InActionMode = false
}
