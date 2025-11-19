package components

import (
	"github.com/y3owk1n/neru/internal/features/action"
	"github.com/y3owk1n/neru/internal/features/grid"
	"github.com/y3owk1n/neru/internal/features/hints"
	"github.com/y3owk1n/neru/internal/features/scroll"
)

// HintsComponent encapsulates all hints-related functionality.
type HintsComponent struct {
	Generator *hints.Generator
	Overlay   *hints.Overlay
	Manager   *hints.Manager
	Router    *hints.Router
	Context   *hints.Context
	Style     hints.StyleMode
}

// GridComponent encapsulates all grid-related functionality.
type GridComponent struct {
	Manager *grid.Manager
	Router  *grid.Router
	Context *grid.Context
	Style   grid.Style
}

// ScrollComponent encapsulates all scroll-related functionality.
type ScrollComponent struct {
	Controller *scroll.Controller
	Overlay    *scroll.Overlay
	Context    *scroll.Context
}

// ActionComponent encapsulates all action-related functionality.
type ActionComponent struct {
	Overlay *action.Overlay
}
