package app

import (
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/overlay"
)

// OverlayRenderer handles rendering operations for the application overlays.
type OverlayRenderer struct {
	mgr       *overlay.Manager
	hintStyle hints.StyleMode
	gridStyle grid.Style
}

// newOverlayRenderer creates a new overlay renderer.
func newOverlayRenderer(mgr *overlay.Manager, hs hints.StyleMode, gs grid.Style) *OverlayRenderer {
	return &OverlayRenderer{mgr: mgr, hintStyle: hs, gridStyle: gs}
}

// DrawHints draws hints with the configured style.
func (r *OverlayRenderer) DrawHints(hs []*hints.Hint) error {
	return r.mgr.DrawHintsWithStyle(hs, r.hintStyle)
}

// DrawGrid draws a grid with the configured style.
func (r *OverlayRenderer) DrawGrid(g *grid.Grid, input string) error {
	return r.mgr.DrawGrid(g, input, r.gridStyle)
}

// ShowSubgrid shows a subgrid for the specified cell.
func (r *OverlayRenderer) ShowSubgrid(cell *grid.Cell) { r.mgr.ShowSubgrid(cell, r.gridStyle) }

// UpdateGridMatches updates the grid matches with the specified prefix.
func (r *OverlayRenderer) UpdateGridMatches(prefix string) { r.mgr.UpdateGridMatches(prefix) }

// SetHideUnmatched sets whether to hide unmatched cells.
func (r *OverlayRenderer) SetHideUnmatched(hide bool) { r.mgr.SetHideUnmatched(hide) }

// Show shows the overlay.
func (r *OverlayRenderer) Show() { r.mgr.Show() }

// Clear clears the overlay.
func (r *OverlayRenderer) Clear() { r.mgr.Clear() }

// ResizeActive resizes the overlay to the active screen.
func (r *OverlayRenderer) ResizeActive() { r.mgr.ResizeToActiveScreenSync() }

// DrawActionHighlight draws an action highlight border around the active screen.
func (r *OverlayRenderer) DrawActionHighlight(x, y, width, height int) {
	r.mgr.DrawActionHighlight(x, y, width, height)
}

// DrawScrollHighlight draws a scroll highlight border around the active screen.
func (r *OverlayRenderer) DrawScrollHighlight(x, y, width, height int) {
	r.mgr.DrawScrollHighlight(x, y, width, height)
}
