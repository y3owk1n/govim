package overlay

/*
#cgo CFLAGS: -x objective-c
#include "../bridge/overlay.h"
*/
import "C"

import (
	"github.com/y3owk1n/neru/internal/action"
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
	"sync"
	"unsafe"
)

type Mode string

const (
	ModeIdle   Mode = "idle"
	ModeHints  Mode = "hints"
	ModeGrid   Mode = "grid"
	ModeAction Mode = "action"
	ModeScroll Mode = "scroll"
)

type StateChange struct {
	Prev Mode
	Next Mode
}

type Manager struct {
	window        C.OverlayWindow
	logger        *zap.Logger
	mu            sync.Mutex
	mode          Mode
	subs          map[uint64]func(StateChange)
	nextID        uint64
	hintOverlay   *hints.Overlay
	gridOverlay   *grid.GridOverlay
	actionOverlay *action.Overlay
	scrollOverlay *scroll.Overlay
}

var (
	mgr  *Manager
	once sync.Once
)

func Init(logger *zap.Logger) *Manager {
	once.Do(func() {
		w := C.createOverlayWindow()
		mgr = &Manager{
			window: w,
			logger: logger,
			mode:   ModeIdle,
			subs:   make(map[uint64]func(StateChange)),
		}
	})
	return mgr
}

func Get() *Manager {
	return mgr
}

func (m *Manager) GetWindowPtr() unsafe.Pointer { return unsafe.Pointer(m.window) }

func (m *Manager) Show()                     { C.showOverlayWindow(m.window) }
func (m *Manager) Hide()                     { C.hideOverlayWindow(m.window) }
func (m *Manager) Clear()                    { C.clearOverlay(m.window) }
func (m *Manager) ResizeToActiveScreenSync() { C.resizeOverlayToActiveScreen(m.window) }

func (m *Manager) SwitchTo(next Mode) {
	m.mu.Lock()
	prev := m.mode
	m.mode = next
	m.mu.Unlock()
	if m.logger != nil {
		m.logger.Debug("Overlay mode switch", zap.String("prev", string(prev)), zap.String("next", string(next)))
	}
	m.publish(StateChange{Prev: prev, Next: next})
}

func (m *Manager) Subscribe(fn func(StateChange)) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	id := m.nextID
	m.subs[id] = fn
	return id
}

func (m *Manager) Unsubscribe(id uint64) {
	m.mu.Lock()
	delete(m.subs, id)
	m.mu.Unlock()
}

func (m *Manager) publish(ev StateChange) {
	m.mu.Lock()
	subs := make([]func(StateChange), 0, len(m.subs))
	for _, s := range m.subs {
		subs = append(subs, s)
	}
	m.mu.Unlock()
	for _, s := range subs {
		s(ev)
	}
}

func (m *Manager) Destroy() {
	if m.window != nil {
		C.destroyOverlayWindow(m.window)
		m.window = nil
	}
}

// Wiring overlay renderers
func (m *Manager) UseHintOverlay(o *hints.Overlay)    { m.hintOverlay = o }
func (m *Manager) UseGridOverlay(o *grid.GridOverlay) { m.gridOverlay = o }
func (m *Manager) UseActionOverlay(o *action.Overlay) { m.actionOverlay = o }
func (m *Manager) UseScrollOverlay(o *scroll.Overlay) { m.scrollOverlay = o }

// Centralized draw methods
func (m *Manager) DrawHintsWithStyle(hs []*hints.Hint, style hints.StyleMode) error {
	if m.hintOverlay == nil {
		return nil
	}
	return m.hintOverlay.DrawHintsWithStyle(hs, style)
}

func (m *Manager) DrawActionHighlight(x, y, w, h int) {
	if m.actionOverlay == nil {
		return
	}
	m.actionOverlay.DrawActionHighlight(x, y, w, h)
}

func (m *Manager) DrawScrollHighlight(x, y, w, h int) {
	if m.scrollOverlay == nil {
		return
	}
	m.scrollOverlay.DrawScrollHighlight(x, y, w, h)
}

func (m *Manager) DrawGrid(g *grid.Grid, input string, style grid.GridStyle) error {
	if m.gridOverlay == nil {
		return nil
	}
	return m.gridOverlay.Draw(g, input, style)
}

func (m *Manager) UpdateGridMatches(prefix string) {
	if m.gridOverlay == nil {
		return
	}
	m.gridOverlay.UpdateMatches(prefix)
}

func (m *Manager) ShowSubgrid(cell *grid.Cell, style grid.GridStyle) {
	if m.gridOverlay == nil {
		return
	}
	m.gridOverlay.ShowSubgrid(cell, style)
}

func (m *Manager) SetHideUnmatched(hide bool) {
	if m.gridOverlay == nil {
		return
	}
	m.gridOverlay.SetHideUnmatched(hide)
}
