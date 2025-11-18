package app

import (
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/overlay"
)

type overlayRenderer struct {
	mgr       *overlay.Manager
	hintStyle hints.StyleMode
	gridStyle grid.Style
}

func newOverlayRenderer(mgr *overlay.Manager, hs hints.StyleMode, gs grid.Style) *overlayRenderer {
	return &overlayRenderer{mgr: mgr, hintStyle: hs, gridStyle: gs}
}

func (r *overlayRenderer) drawHints(hs []*hints.Hint) error {
	return r.mgr.DrawHintsWithStyle(hs, r.hintStyle)
}

func (r *overlayRenderer) drawGrid(g *grid.Grid, input string) error {
	return r.mgr.DrawGrid(g, input, r.gridStyle)
}

func (r *overlayRenderer) showSubgrid(cell *grid.Cell) { r.mgr.ShowSubgrid(cell, r.gridStyle) }

func (r *overlayRenderer) updateGridMatches(prefix string) { r.mgr.UpdateGridMatches(prefix) }

func (r *overlayRenderer) setHideUnmatched(hide bool) { r.mgr.SetHideUnmatched(hide) }

func (r *overlayRenderer) show() { r.mgr.Show() }

func (r *overlayRenderer) clear() { r.mgr.Clear() }

func (r *overlayRenderer) resizeActive() { r.mgr.ResizeToActiveScreenSync() }

// drawActionHighlight draws an action highlight border around the active screen.
func (r *overlayRenderer) drawActionHighlight(x, y, width, height int) {
	r.mgr.DrawActionHighlight(x, y, width, height)
}

// drawScrollHighlight draws a scroll highlight border around the active screen.
func (r *overlayRenderer) drawScrollHighlight(x, y, width, height int) {
	r.mgr.DrawScrollHighlight(x, y, width, height)
}
