package main

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
)

type HintsContext struct {
	selectedHint  *hints.Hint
	canScroll     bool
	lastScrollKey string
}

// activateMode activates a mode with a given action (for hints mode)
func (a *App) activateMode(mode Mode, action Action) {
	if mode == ModeIdle {
		// Explicit idle transition
		a.exitMode()
		return
	}
	if mode == ModeHints {
		a.activateHintMode(action)
		return
	}
	// Unknown or unsupported mode
	a.logger.Warn("Unknown mode", zap.String("mode", getModeString(mode)))
}

func (a *App) activateHintMode(action Action) {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring hint mode activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	actionString := getActionString(action)
	a.logger.Info("Activating hint mode", zap.String("action", actionString))

	a.exitMode() // Exit current mode first

	if actionString == "unknown" {
		a.logger.Warn("Unknown action, ignoring")
		return
	}

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Collect elements based on mode
	elements := a.collectElementsForAction(action)
	if len(elements) == 0 {
		a.logger.Warn("No elements found for action", zap.String("action", actionString))
		return
	}

	// Generate and setup hints
	if err := a.setupHints(elements, action); err != nil {
		a.logger.Error("Failed to setup hints", zap.Error(err), zap.String("action", actionString))
		return
	}

	// Update mode and enable event tap
	a.currentMode = ModeHints
	a.hintsCtx.selectedHint = nil
	a.hintsCtx.canScroll = false
	a.currentAction = action

	// Enable event tap to capture keys (must be last to ensure proper initialization)
	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.logModeActivation(action)
}

// setupHints generates hints and draws them with appropriate styling
func (a *App) setupHints(elements []*accessibility.TreeNode, action Action) error {
	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		return fmt.Errorf("failed to generate hints: %w", err)
	}

	// Set up hints in the hint manager
	hintCollection := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hintCollection)

	// Draw hints with mode-specific styling
	style := hints.BuildStyleForAction(a.config.Hints, getActionString(action))
	if err := a.hintOverlay.DrawHintsWithStyle(hintList, style); err != nil {
		return fmt.Errorf("failed to draw hints: %w", err)
	}

	a.hintOverlay.Show()
	return nil
}

// handleActiveKey dispatches key events by current mode
func (a *App) handleKeyPress(key string) {
	// Only handle keys when in an active mode
	if a.currentMode == ModeIdle {
		return
	}

	// Explicitly dispatch by current mode
	switch a.currentMode {
	case ModeHints:
		// Route hint-specific keys via hints router
		res := a.hintsRouter.RouteKey(key, a.hintsCtx.selectedHint != nil, a.hintsCtx.canScroll, a.currentAction == ActionScroll)
		if res.Exit {
			a.exitMode()
			return
		}
		if res.SwitchToScroll {
			a.logger.Debug("Switching to scroll mode from scroll hints")
			a.hintsCtx.canScroll = true
			a.hintOverlay.Clear()
			a.showScroll()
			return
		}

		// Context menu active takes precedence
		if a.hintsCtx.selectedHint != nil {
			a.handleContextMenuKey(key)
			return
		}
		// Scroll active
		if a.hintsCtx.canScroll {
			a.handleScrollKey(key)
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
			center := image.Point{X: info.Position.X + info.Size.X/2, Y: info.Position.Y + info.Size.Y/2}
			switch a.currentAction {
			case ActionLeftClick:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.LeftClickAtPoint(center, a.config.Hints.LeftClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionRightClick:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.RightClickAtPoint(center, a.config.Hints.RightClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionDoubleClick:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.DoubleClickAtPoint(center, a.config.Hints.DoubleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionTripleClick:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.TripleClickAtPoint(center, a.config.Hints.TripleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionMouseUp:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.LeftMouseUpAtPoint(center); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionMouseDown:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.LeftMouseDownAtPoint(center); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
				// Immediately ask for release
				a.activateMode(ModeHints, ActionMouseUp)
			case ActionMiddleClick:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				if err := accessibility.MiddleClickAtPoint(center, a.config.Hints.MiddleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click element", zap.Error(err))
				}
				a.exitMode()
			case ActionMoveMouse:
				a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
				accessibility.MoveMouseToPoint(center)
				a.exitMode()
			case ActionScroll:
				// Enter scroll mode - scroll to element
				a.hintsCtx.canScroll = true
				a.logger.Info("Hint selected, start scrolling", zap.String("label", a.hintManager.GetInput()))
				accessibility.MoveMouseToPoint(center)
				a.hintOverlay.Clear()
				a.showScroll()
			case ActionContextMenu:
				// Store the hint and wait for action selection
				a.hintsCtx.selectedHint = hint
				a.logger.Info("Hint selected, choose an action", zap.String("label", a.hintManager.GetInput()))
				// Clear all hints and show action menu at the hint location
				a.hintOverlay.Clear()
				a.showContextMenu(hint)
			}
			return
		}
	default:
		// No router for this mode yet
		return
	}
}

// showContextMenu displays the action selection menu at the hint location
func (a *App) showContextMenu(hint *hints.Hint) {

	menuActions := hints.BuildContextMenuLabel()

	actionHints := make([]*hints.Hint, 0, 1)
	baseX := hint.Position.X
	baseY := hint.Position.Y

	actionHint := &hints.Hint{
		Label:    menuActions,
		Element:  hint.Element,
		Position: image.Point{X: baseX, Y: baseY + 5},
	}

	actionHints = append(actionHints, actionHint)

	// Draw target indicator dot first
	dotRadius := 4.0
	dotColor := a.config.Hints.ContextMenuHints.BackgroundColor
	dotBorderColor := a.config.Hints.ContextMenuHints.BorderColor
	dotBorderWidth := float64(a.config.Hints.BorderWidth)

	if err := a.hintOverlay.DrawTargetDot(baseX, baseY, dotRadius, dotColor, dotBorderColor, dotBorderWidth); err != nil {
		a.logger.Error("Failed to draw target dot", zap.Error(err))
	}

	contextMenuStyle := hints.BuildStyleForAction(a.config.Hints, getActionString(ActionContextMenu))

	if err := a.hintOverlay.DrawHintsWithoutArrow(actionHints, contextMenuStyle); err != nil {
		a.logger.Error("Failed to draw action menu", zap.Error(err))
	}
}

// handleContextMenuKey handles action key presses after a hint is selected
func (a *App) handleContextMenuKey(key string) {
	if a.hintsCtx.selectedHint == nil {
		return
	}

	// Handle backspace/delete to go back to hint selection
	if key == "\x7f" || key == "delete" || key == "backspace" {
		a.logger.Debug("Backspace pressed in action mode, returning to hint selection")

		// Clear selected hint and reset hint manager
		a.hintsCtx.selectedHint = nil
		a.hintManager.Reset()
		return
	}

	hint := a.hintsCtx.selectedHint
	info, getErr := hint.Element.Element.GetInfo()
	if getErr != nil {
		a.logger.Error("Failed to get element info", zap.Error(getErr))
		return
	}
	center := image.Point{X: info.Position.X + info.Size.X/2, Y: info.Position.Y + info.Size.Y/2}
	a.logger.Info("Action key pressed", zap.String("key", key))

	var err error
	var successCallback func()
	// Hardcoded action keys (tied to UI labels)
	action := hints.ParseContextMenuKey(key)
	switch action {
	case "left_click": // Left click
		a.logger.Info("Performing left click", zap.String("label", hint.Label))
		err = accessibility.LeftClickAtPoint(center, a.config.Hints.LeftClickHints.RestoreCursor)
	case "right_click": // Right click
		a.logger.Info("Performing right click", zap.String("label", hint.Label))
		err = accessibility.RightClickAtPoint(center, a.config.Hints.RightClickHints.RestoreCursor)
	case "double_click": // Double click
		a.logger.Info("Performing double click", zap.String("label", hint.Label))
		err = accessibility.DoubleClickAtPoint(center, a.config.Hints.DoubleClickHints.RestoreCursor)
	case "triple_click": // Triple click
		a.logger.Info("Performing triple click", zap.String("label", hint.Label))
		err = accessibility.TripleClickAtPoint(center, a.config.Hints.TripleClickHints.RestoreCursor)
	case "mouse_down": // Hold mouse
		a.logger.Info("Performing hold mouse", zap.String("label", hint.Label))
		err = accessibility.LeftMouseDownAtPoint(center)

		// Activate mouse up mode straight away
		successCallback = func() {
			a.activateMode(ModeHints, ActionMouseUp)
		}
	case "mouse_up": // Release mouse
		a.logger.Info("Performing release hold mouse", zap.String("label", hint.Label))
		err = accessibility.LeftMouseUpAtPoint(center)
	case "middle_click": // Middle click
		a.logger.Info("Performing middle click", zap.String("label", hint.Label))
		err = accessibility.MiddleClickAtPoint(center, a.config.Hints.MiddleClickHints.RestoreCursor)
	case "move_mouse": // Move mouse to a position
		a.logger.Info("Performing go to position", zap.String("label", hint.Label))
		accessibility.MoveMouseToPoint(center)
	default:
		a.logger.Debug("Unknown action key, ignoring", zap.String("key", key))
		return
	}

	if err != nil {
		a.logger.Error("Failed to perform action", zap.Error(err))
	}

	a.exitMode()

	if successCallback != nil {
		successCallback()
	}
}

// showScroll displays the action selection menu at the hint location
func (a *App) showScroll() {
	a.drawScrollHighlightBorder()
}

// handleScrollKey handles key presses in scroll mode
func (a *App) handleScrollKey(key string) {
	// Log every byte for debugging
	bytes := []byte(key)
	a.logger.Info("Scroll key pressed",
		zap.String("key", key),
		zap.Int("len", len(key)),
		zap.String("hex", fmt.Sprintf("%#v", key)),
		zap.Any("bytes", bytes))

	var err error

	// Check for control characters
	if len(key) == 1 {
		byteVal := key[0]
		a.logger.Info("Checking control char", zap.Uint8("byte", byteVal))
		// Only handle Ctrl+D / Ctrl+U here; let other single keys fall through
		if byteVal == 4 || byteVal == 21 {
			op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
			if ok {
				a.hintsCtx.lastScrollKey = ""
				switch op {
				case "half_down":
					a.logger.Info("Ctrl+D detected - half page down")
					err = a.scrollController.ScrollDownHalfPage()
					goto done
				case "half_up":
					a.logger.Info("Ctrl+U detected - half page up")
					err = a.scrollController.ScrollUpHalfPage()
					goto done
				}
			}
		}
	}

	// Regular keys
	switch key {
	case "j":
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "down" {
			err = a.scrollController.ScrollDown()
		}
	case "k":
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "up" {
			err = a.scrollController.ScrollUp()
		}
	case "h":
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "left" {
			err = a.scrollController.ScrollLeft()
		}
	case "l":
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "right" {
			err = a.scrollController.ScrollRight()
		}
	case "g": // gg for top (need to press twice)
		op, newLast, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if !ok {
			a.logger.Info("First g pressed, press again for top")
			a.hintsCtx.lastScrollKey = newLast
			return
		}
		if op == "top" {
			a.logger.Info("gg detected - scroll to top")
			err = a.scrollController.ScrollToTop()
			a.hintsCtx.lastScrollKey = ""
			goto done
		}
	case "G": // Shift+G for bottom
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if ok && op == "bottom" {
			a.logger.Info("G key detected - scroll to bottom")
			err = a.scrollController.ScrollToBottom()
			a.hintsCtx.lastScrollKey = ""
		}
	case "\t":
		op, _, ok := scroll.ParseKey(key, a.hintsCtx.lastScrollKey)
		if ok && op == "tab" {
			a.logger.Info("Tab key detected - show hint again")
			a.activateMode(ModeHints, ActionScroll)
		}
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		a.hintsCtx.lastScrollKey = ""
		return
	}

	// Reset last key for most commands
	a.hintsCtx.lastScrollKey = ""

done:
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
}

func (a *App) drawScrollHighlightBorder() {
	a.scrollController.DrawHighlightBorder(a.hintOverlay)
}

// exitMode exits the current mode
func (a *App) exitMode() {
	if a.currentMode == ModeIdle {
		return
	}

	a.logger.Info("Exiting current mode", zap.String("mode", a.getCurrModeString()))

	// Mode-specific cleanup
	switch a.currentMode {
	case ModeHints:
		// If we are in mouse up action, remove the mouse up to prevent further dragging
		if a.currentAction == ActionMouseUp {
			a.logger.Info("Detected MouseUp action, removing mouse up event...")
			err := accessibility.LeftMouseUp()
			if err != nil {
				a.logger.Error("Failed to remove mouse up", zap.Error(err))
			}
		}

		if a.hintManager != nil {
			a.hintManager.Reset()
		}
		a.hintsCtx.selectedHint = nil
		a.hintsCtx.canScroll = false

		// Clear and hide overlay for hints
		a.hintOverlay.Clear()
		a.hintOverlay.Hide()
	default:
		// No domain-specific cleanup for other modes yet
	}

	// Disable event tap when leaving active modes
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	// Update mode after all cleanup is done
	a.currentMode = ModeIdle
	a.currentAction = ActionLeftClick
	a.logger.Debug("Mode transition complete",
		zap.String("to", "idle"))
}

func getModeString(mode Mode) string {
	switch mode {
	case ModeIdle:
		return "idle"
	case ModeHints:
		return "hints"
	default:
		return "unknown"
	}
}

func getActionString(action Action) string {
	switch action {
	case ActionLeftClick:
		return "left_click"
	case ActionRightClick:
		return "right_click"
	case ActionDoubleClick:
		return "double_click"
	case ActionTripleClick:
		return "triple_click"
	case ActionMouseUp:
		return "mouse_up"
	case ActionMouseDown:
		return "mouse_down"
	case ActionMiddleClick:
		return "middle_click"
	case ActionMoveMouse:
		return "move_mouse"
	case ActionScroll:
		return "scroll"
	case ActionContextMenu:
		return "context_menu"
	default:
		return "unknown"
	}
}

// getCurrModeString returns the current mode as a string
func (a *App) getCurrModeString() string {
	return getModeString(a.currentMode)
}

// logModeActivation logs mode-specific activation messages
func (a *App) logModeActivation(action Action) {
	hintCount := len(a.hintManager.GetHints())

	switch action {
	case ActionLeftClick:
		a.logger.Info("Hint mode with left click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionRightClick:
		a.logger.Info("Hint mode with right click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionDoubleClick:
		a.logger.Info("Hint mode with double click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionTripleClick:
		a.logger.Info("Hint mode with triple click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionMouseUp:
		a.logger.Info("Hint mode with mouse up activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionMouseDown:
		a.logger.Info("Hint mode with mouse down activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionMiddleClick:
		a.logger.Info("Hint mode with middle click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionMoveMouse:
		a.logger.Info("Hint mode with move mouse activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ActionScroll:
		a.logger.Info("Scroll mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom")
	case ActionContextMenu:
		a.logger.Info("Context menu mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to choose an action")
	}
}
