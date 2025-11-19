package state

import (
	"sync"
)

// Mode represents the current application mode.
type Mode int

const (
	// ModeIdle is the default mode when no operation is active.
	ModeIdle Mode = iota
	// ModeHints is active during hint-based navigation.
	ModeHints
	// ModeGrid is active during grid-based navigation.
	ModeGrid
)

// String returns the string representation of the mode.
func (m Mode) String() string {
	switch m {
	case ModeIdle:
		return "idle"
	case ModeHints:
		return "hints"
	case ModeGrid:
		return "grid"
	default:
		return "unknown"
	}
}

// AppState manages the core application state including enabled status,
// current mode, and various operational flags.
type AppState struct {
	mu sync.RWMutex

	// Core state
	enabled     bool
	currentMode Mode

	// Operational flags
	hotkeysRegistered       bool
	screenChangeProcessing  bool
	gridOverlayNeedsRefresh bool
	hintOverlayNeedsRefresh bool
	hotkeyRefreshPending    bool

	// Scroll state
	idleScrollLastKey string
	isScrollingActive bool
}

// NewAppState creates a new AppState with default values.
func NewAppState() *AppState {
	return &AppState{
		enabled:     true,
		currentMode: ModeIdle,
	}
}

// IsEnabled returns whether the application is enabled.
func (s *AppState) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetEnabled sets the enabled state of the application.
func (s *AppState) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// Enable enables the application.
func (s *AppState) Enable() {
	s.SetEnabled(true)
}

// Disable disables the application.
func (s *AppState) Disable() {
	s.SetEnabled(false)
}

// CurrentMode returns the current application mode.
func (s *AppState) CurrentMode() Mode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentMode
}

// SetMode sets the current application mode.
func (s *AppState) SetMode(mode Mode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentMode = mode
}

// HotkeysRegistered returns whether hotkeys are currently registered.
func (s *AppState) HotkeysRegistered() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hotkeysRegistered
}

// SetHotkeysRegistered sets the hotkeys registered flag.
func (s *AppState) SetHotkeysRegistered(registered bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hotkeysRegistered = registered
}

// ScreenChangeProcessing returns whether a screen change is being processed.
func (s *AppState) ScreenChangeProcessing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.screenChangeProcessing
}

// SetScreenChangeProcessing sets the screen change processing flag.
func (s *AppState) SetScreenChangeProcessing(processing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.screenChangeProcessing = processing
}

// GridOverlayNeedsRefresh returns whether the grid overlay needs refresh.
func (s *AppState) GridOverlayNeedsRefresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.gridOverlayNeedsRefresh
}

// SetGridOverlayNeedsRefresh sets the grid overlay refresh flag.
func (s *AppState) SetGridOverlayNeedsRefresh(needs bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gridOverlayNeedsRefresh = needs
}

// HintOverlayNeedsRefresh returns whether the hint overlay needs refresh.
func (s *AppState) HintOverlayNeedsRefresh() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hintOverlayNeedsRefresh
}

// SetHintOverlayNeedsRefresh sets the hint overlay refresh flag.
func (s *AppState) SetHintOverlayNeedsRefresh(needs bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hintOverlayNeedsRefresh = needs
}

// HotkeyRefreshPending returns whether a hotkey refresh is pending.
func (s *AppState) HotkeyRefreshPending() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hotkeyRefreshPending
}

// SetHotkeyRefreshPending sets the hotkey refresh pending flag.
func (s *AppState) SetHotkeyRefreshPending(pending bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hotkeyRefreshPending = pending
}

// IdleScrollLastKey returns the last key pressed during idle scroll.
func (s *AppState) IdleScrollLastKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.idleScrollLastKey
}

// SetIdleScrollLastKey sets the last key pressed during idle scroll.
func (s *AppState) SetIdleScrollLastKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idleScrollLastKey = key
}

// IsScrollingActive returns whether scrolling is currently active.
func (s *AppState) IsScrollingActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isScrollingActive
}

// SetScrollingActive sets the scrolling active flag.
func (s *AppState) SetScrollingActive(active bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isScrollingActive = active
}
