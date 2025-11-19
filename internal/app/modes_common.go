package app

import (
	"image"
	"math"

	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/features/grid"
	"github.com/y3owk1n/neru/internal/features/hints"
	"github.com/y3owk1n/neru/internal/infra/accessibility"
	"github.com/y3owk1n/neru/internal/infra/bridge"
	"github.com/y3owk1n/neru/internal/ui/overlay"
	"go.uber.org/zap"
)

const unknownAction = "unknown"

// HintsContext holds the state and context for hint mode operations.
type HintsContext struct {
	selectedHint *hints.Hint
	inActionMode bool
}

// GridContext holds the state and context for grid mode operations.
type GridContext struct {
	gridInstance **grid.Grid
	gridOverlay  **grid.Overlay
	inActionMode bool
}

// handleActiveKey dispatches key events by current mode.
func (a *App) handleKeyPress(key string) {
	// If in idle mode, check if we should handle scroll keys
	if a.state.CurrentMode() == ModeIdle {
		// Handle escape to exit standalone scroll
		if key == "\x1b" || key == "escape" {
			a.logger.Info("Exiting standalone scroll mode")
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			if a.eventTap != nil {
				a.eventTap.Disable()
			}
			a.state.SetScrollingActive(false)
			a.state.SetIdleScrollLastKey("") // Reset scroll state
			return
		}
		// Try to handle scroll keys with generic handler using persistent state
		// If it's not a scroll key, it will just be ignored
		lastKey := a.state.IdleScrollLastKey()
		a.handleGenericScrollKey(key, &lastKey)
		a.state.SetIdleScrollLastKey(lastKey)
		return
	}

	// Handle Tab key to toggle between overlay mode and action mode
	if key == "\t" { // Tab key
		a.handleTabKey()
		return
	}

	// Handle Escape key to exit action mode or current mode
	if key == "\x1b" || key == "escape" {
		a.handleEscapeKey()
		return
	}

	// Explicitly dispatch by current mode
	a.handleModeSpecificKey(key)
}

// handleTabKey handles the tab key to toggle between overlay mode and action mode.
func (a *App) handleTabKey() {
	switch a.state.CurrentMode() {
	case ModeHints:
		if a.hintsCtx.inActionMode {
			// Switch back to overlay mode
			a.hintsCtx.inActionMode = false
			if overlay.Get() != nil {
				overlay.Get().Clear()
				overlay.Get().Hide()
			}
			// Re-activate hint mode while preserving action mode state
			a.activateHintModeInternal(true)
			a.logger.Info("Switched back to hints overlay mode")
			a.overlaySwitch(overlay.ModeHints)
		} else {
			// Switch to action mode
			a.hintsCtx.inActionMode = true
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			a.drawHintsActionHighlight()
			a.overlayManager.Show()
			a.logger.Info("Switched to hints action mode")
			a.overlaySwitch(overlay.ModeAction)
		}
	case ModeGrid:
		if a.gridCtx.inActionMode {
			// Switch back to overlay mode
			a.gridCtx.inActionMode = false
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			// Re-setup grid to show grid again with proper refresh
			err := a.setupGrid()
			if err != nil {
				a.logger.Error("Failed to re-setup grid", zap.Error(err))
			}
			a.logger.Info("Switched back to grid overlay mode")
			a.overlaySwitch(overlay.ModeGrid)
		} else {
			// Switch to action mode
			a.gridCtx.inActionMode = true
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			a.drawGridActionHighlight()
			a.overlayManager.Show()
			a.logger.Info("Switched to grid action mode")
			a.overlaySwitch(overlay.ModeAction)
		}
	case ModeIdle:
		// Nothing to do in idle mode
		return
	}
}

// handleEscapeKey handles the escape key to exit action mode or current mode.
func (a *App) handleEscapeKey() {
	switch a.state.CurrentMode() {
	case ModeHints:
		if a.hintsCtx.inActionMode {
			// Exit action mode completely, go back to idle mode
			a.hintsCtx.inActionMode = false
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			a.exitMode()
			a.logger.Info("Exited hints action mode completely")
			a.overlaySwitch(overlay.ModeIdle)
			return
		}
		// Fall through to exit mode
	case ModeGrid:
		if a.gridCtx.inActionMode {
			// Exit action mode completely, go back to idle mode
			a.gridCtx.inActionMode = false
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			a.exitMode()
			a.logger.Info("Exited grid action mode completely")
			a.overlaySwitch(overlay.ModeIdle)
			return
		}
		// Fall through to exit mode
	case ModeIdle:
		// Nothing to do in idle mode
		return
	}
	a.exitMode()
	a.setModeIdle()
}

// handleModeSpecificKey handles mode-specific key processing.
func (a *App) handleModeSpecificKey(key string) {
	switch a.state.CurrentMode() {
	case ModeHints:
		// If in action mode, handle action keys
		if a.hintsCtx.inActionMode {
			a.handleHintsActionKey(key)
			// After handling the action, we stay in action mode
			// The user can press Tab to go back to overlay mode or perform more actions
			return
		}

		// Route hint-specific keys via hints router
		res := a.hintsRouter.RouteKey(key, a.hintsCtx.selectedHint != nil)
		if res.Exit {
			a.exitMode()
			return
		}

		// Hint input processed by router; if exact match, perform action
		if res.ExactHint != nil {
			hint := res.ExactHint
			info, err := hint.Element.Element.GetInfo()
			if err != nil {
				a.logger.Error("Failed to get element info", zap.Error(err))
				a.exitMode()
				return
			}
			center := image.Point{
				X: info.Position.X + info.Size.X/2,
				Y: info.Position.Y + info.Size.Y/2,
			}

			a.logger.Info("Found element", zap.String("label", a.hintManager.GetInput()))
			accessibility.MoveMouseToPoint(center)

			if a.hintManager != nil {
				a.hintManager.Reset()
			}
			a.hintsCtx.selectedHint = nil

			a.activateHintMode()

			return
		}
	case ModeGrid:
		// If in action mode, handle action keys
		if a.gridCtx.inActionMode {
			a.handleGridActionKey(key)
			// After handling the action, we stay in action mode
			// The user can press Tab to go back to overlay mode or perform more actions
			return
		}

		res := a.gridRouter.RouteKey(key)
		if res.Exit {
			a.exitMode()
			return
		}

		// Complete coordinate entered - perform action
		if res.Complete {
			targetPoint := res.TargetPoint

			// Convert from window-local coordinates to absolute screen coordinates
			// The grid was generated with normalized bounds (0,0 origin) but clicks need absolute coords
			screenBounds := bridge.GetActiveScreenBounds()
			absolutePoint := image.Point{
				X: targetPoint.X + screenBounds.Min.X,
				Y: targetPoint.Y + screenBounds.Min.Y,
			}

			a.logger.Info(
				"Grid move mouse",
				zap.Int("x", absolutePoint.X),
				zap.Int("y", absolutePoint.Y),
			)
			accessibility.MoveMouseToPoint(absolutePoint)

			// No need to exit grid mode, just let it going
			// a.exitMode()

			return
		}
	case ModeIdle:
		// Nothing to do in idle mode
		return
	}
}

// exitMode exits the current mode.
func (a *App) exitMode() {
	if a.state.CurrentMode() == ModeIdle {
		return
	}

	a.logger.Info("Exiting current mode", zap.String("mode", a.getCurrModeString()))

	// Mode-specific cleanup
	a.performModeSpecificCleanup()

	// Common cleanup
	a.performCommonCleanup()

	// Cursor restoration
	a.handleCursorRestoration()
}

// performModeSpecificCleanup handles mode-specific cleanup logic.
func (a *App) performModeSpecificCleanup() {
	switch a.state.CurrentMode() {
	case ModeHints:
		a.cleanupHintsMode()
	case ModeGrid:
		a.cleanupGridMode()
	default:
		a.cleanupDefaultMode()
	}
}

// cleanupDefaultMode handles cleanup for default/unknown modes.
func (a *App) cleanupDefaultMode() {
	// No domain-specific cleanup for other modes yet
	// But still clear and hide action overlay
	if overlay.Get() != nil {
		overlay.Get().Clear()
		overlay.Get().Hide()
	}
}

// performCommonCleanup handles common cleanup logic for all modes.
func (a *App) performCommonCleanup() {
	// Clear scroll overlay
	a.overlayManager.Clear()

	// Disable event tap when leaving active modes
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	// Update mode after all cleanup is done
	a.state.SetMode(ModeIdle)
	a.logger.Debug("Mode transition complete",
		zap.String("to", "idle"))
	a.overlayManager.SwitchTo(overlay.ModeIdle)

	// If a hotkey refresh was deferred while in an active mode, perform it now
	if a.state.HotkeyRefreshPending() {
		a.state.SetHotkeyRefreshPending(false)
		go a.refreshHotkeysForAppOrCurrent("")
	}
}

// handleCursorRestoration handles cursor position restoration on exit.
func (a *App) handleCursorRestoration() {
	shouldRestore := a.shouldRestoreCursorOnExit()
	if shouldRestore {
		currentBounds := bridge.GetActiveScreenBounds()
		target := computeRestoredPosition(
			a.cursor.GetInitialPosition(),
			a.cursor.GetInitialScreenBounds(),
			currentBounds,
		)
		accessibility.MoveMouseToPoint(target)
	}
	a.cursor.Reset()
	a.state.SetScrollingActive(false)
}

func getModeString(mode Mode) string {
	switch mode {
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

func getActionString(action domain.Action) string {
	switch action {
	case domain.ActionLeftClick:
		return "left_click"
	case domain.ActionRightClick:
		return "right_click"
	case domain.ActionMouseUp:
		return "mouse_up"
	case domain.ActionMouseDown:
		return "mouse_down"
	case domain.ActionMiddleClick:
		return "middle_click"
	case domain.ActionMoveMouse:
		return "move_mouse"
	case domain.ActionScroll:
		return "scroll"
	default:
		return "unknown"
	}
}

// getCurrModeString returns the current mode as a string.
func (a *App) getCurrModeString() string {
	return getModeString(a.state.CurrentMode())
}

func (a *App) captureInitialCursorPosition() {
	if a.cursor.IsCaptured() {
		return
	}
	pos := accessibility.GetCurrentCursorPosition()
	bounds := bridge.GetActiveScreenBounds()
	a.cursor.Capture(pos, bounds)
}

func (a *App) shouldRestoreCursorOnExit() bool {
	if a.config == nil {
		return false
	}
	if !a.config.General.RestoreCursorPosition {
		return false
	}
	if !a.cursor.IsCaptured() {
		return false
	}
	if a.state.IsScrollingActive() {
		return false
	}
	if a.cursor.ShouldRestore() {
		// Will skip this restoration
		return false
	}
	return true
}

func computeRestoredPosition(initPos image.Point, from, to image.Rectangle) image.Point {
	if from == to {
		return initPos
	}
	if from.Dx() == 0 || from.Dy() == 0 || to.Dx() == 0 || to.Dy() == 0 {
		return initPos
	}

	rx := float64(initPos.X-from.Min.X) / float64(from.Dx())
	ry := float64(initPos.Y-from.Min.Y) / float64(from.Dy())

	if rx < 0 {
		rx = 0
	}
	if rx > 1 {
		rx = 1
	}
	if ry < 0 {
		ry = 0
	}
	if ry > 1 {
		ry = 1
	}

	nx := to.Min.X + int(math.Round(rx*float64(to.Dx())))
	ny := to.Min.Y + int(math.Round(ry*float64(to.Dy())))

	if nx < to.Min.X {
		nx = to.Min.X
	}
	if nx > to.Max.X {
		nx = to.Max.X
	}
	if ny < to.Min.Y {
		ny = to.Min.Y
	}
	if ny > to.Max.Y {
		ny = to.Max.Y
	}
	return image.Point{X: nx, Y: ny}
}

// handleActionKey handles action keys for both hints and grid modes.
func (a *App) handleActionKey(key string, mode string) {
	// Get the current cursor position
	cursorPos := accessibility.GetCurrentCursorPosition()

	// Map action keys to actions using configurable keys
	var act string
	switch key {
	case a.config.Action.LeftClickKey:
		a.logger.Info(mode + " action: Left click")
		act = string(domain.ActionNameLeftClick)
	case a.config.Action.RightClickKey:
		a.logger.Info(mode + " action: Right click")
		act = string(domain.ActionNameRightClick)
	case a.config.Action.MiddleClickKey:
		a.logger.Info(mode + " action: Middle click")
		act = string(domain.ActionNameMiddleClick)
	case a.config.Action.MouseDownKey:
		a.logger.Info(mode + " action: Mouse down")
		act = string(domain.ActionNameMouseDown)
	case a.config.Action.MouseUpKey:
		a.logger.Info(mode + " action: Mouse up")
		act = string(domain.ActionNameMouseUp)
	default:
		a.logger.Debug("Unknown "+mode+" action key", zap.String("key", key))
		return
	}
	err := performActionAtPoint(act, cursorPos)
	if err != nil {
		a.logger.Error("Failed to perform action", zap.Error(err))
	}
}
