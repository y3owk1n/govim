package scroll

// Context holds the state and context for scroll mode operations.
type Context struct {
	// LastKey tracks the last key pressed during scroll operations
	// This is used for multi-key operations like "gg" for top
	LastKey string

	// IsActive indicates whether scroll mode is currently active
	IsActive bool
}

// SetLastKey sets the last key pressed during scroll operations.
func (c *Context) SetLastKey(key string) {
	c.LastKey = key
}

// SetIsActive sets whether scroll mode is currently active.
func (c *Context) SetIsActive(active bool) {
	c.IsActive = active
}

// Reset resets the scroll context to its initial state.
func (c *Context) Reset() {
	c.LastKey = ""
	c.IsActive = false
}
