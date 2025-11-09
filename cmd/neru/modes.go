package main

import (
	"fmt"
	"image"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/hints"
	"go.uber.org/zap"
)

// activateHintMode activates hint mode
func (a *App) activateHintMode(mode Mode) {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring hint mode activation")
		return
	}

	if mode == ModeIdle {
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	modeString := getModeString(mode)
	a.logger.Info("Activating hint mode", zap.String("mode", modeString))

	a.exitMode() // Exit current mode first

	if modeString == "unknown" {
		a.logger.Warn("Unknown mode, ignoring")
		return
	}

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Collect elements based on mode
	elements := a.collectElementsForMode(mode)
	if len(elements) == 0 {
		a.logger.Warn("No elements found for mode", zap.String("mode", modeString))
		return
	}

	// Generate and setup hints
	if err := a.setupHints(elements, mode); err != nil {
		a.logger.Error("Failed to setup hints", zap.Error(err), zap.String("mode", modeString))
		return
	}

	// Update mode and enable event tap
	a.currentMode = mode
	a.selectedHint = nil
	a.canScroll = false

	// Enable event tap to capture keys (must be last to ensure proper initialization)
	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.logModeActivation(mode)
}

// setupHints generates hints and draws them with appropriate styling
func (a *App) setupHints(elements []*accessibility.TreeNode, mode Mode) error {
	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		return fmt.Errorf("failed to generate hints: %w", err)
	}

	// Set up hints in the hint manager
	hints := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hints)

	// Draw hints with mode-specific styling
	style := a.getStyleForMode(mode)
	if err := a.hintOverlay.DrawHintsWithStyle(hintList, style); err != nil {
		return fmt.Errorf("failed to draw hints: %w", err)
	}

	a.hintOverlay.Show()
	return nil
}

// getStyleForMode returns the appropriate hint style configuration for the given mode
func (a *App) getStyleForMode(mode Mode) hints.StyleMode {
	style := hints.StyleMode{
		FontSize:     a.config.Hints.FontSize,
		FontFamily:   a.config.Hints.FontFamily,
		BorderRadius: a.config.Hints.BorderRadius,
		Padding:      a.config.Hints.Padding,
		BorderWidth:  a.config.Hints.BorderWidth,
		Opacity:      a.config.Hints.Opacity,
	}

	switch mode {
	case ModeHintLeftClick:
		style.BackgroundColor = a.config.Hints.LeftClickHints.BackgroundColor
		style.TextColor = a.config.Hints.LeftClickHints.TextColor
		style.MatchedTextColor = a.config.Hints.LeftClickHints.MatchedTextColor
		style.BorderColor = a.config.Hints.LeftClickHints.BorderColor
	case ModeHintRightClick:
		style.BackgroundColor = a.config.Hints.RightClickHints.BackgroundColor
		style.TextColor = a.config.Hints.RightClickHints.TextColor
		style.MatchedTextColor = a.config.Hints.RightClickHints.MatchedTextColor
		style.BorderColor = a.config.Hints.RightClickHints.BorderColor
	case ModeHintDoubleClick:
		style.BackgroundColor = a.config.Hints.DoubleClickHints.BackgroundColor
		style.TextColor = a.config.Hints.DoubleClickHints.TextColor
		style.MatchedTextColor = a.config.Hints.DoubleClickHints.MatchedTextColor
		style.BorderColor = a.config.Hints.DoubleClickHints.BorderColor
	case ModeHintTripleClick:
		style.BackgroundColor = a.config.Hints.TripleClickHints.BackgroundColor
		style.TextColor = a.config.Hints.TripleClickHints.TextColor
		style.MatchedTextColor = a.config.Hints.TripleClickHints.MatchedTextColor
		style.BorderColor = a.config.Hints.TripleClickHints.BorderColor
	case ModeHintMouseUp:
		style.BackgroundColor = a.config.Hints.MouseUpHints.BackgroundColor
		style.TextColor = a.config.Hints.MouseUpHints.TextColor
		style.MatchedTextColor = a.config.Hints.MouseUpHints.MatchedTextColor
		style.BorderColor = a.config.Hints.MouseUpHints.BorderColor
	case ModeHintMouseDown:
		style.BackgroundColor = a.config.Hints.MouseDownHints.BackgroundColor
		style.TextColor = a.config.Hints.MouseDownHints.TextColor
		style.MatchedTextColor = a.config.Hints.MouseDownHints.MatchedTextColor
		style.BorderColor = a.config.Hints.MouseDownHints.BorderColor
	case ModeHintMoveMouse:
		style.BackgroundColor = a.config.Hints.MoveMouseHints.BackgroundColor
		style.TextColor = a.config.Hints.MoveMouseHints.TextColor
		style.MatchedTextColor = a.config.Hints.MoveMouseHints.MatchedTextColor
		style.BorderColor = a.config.Hints.MoveMouseHints.BorderColor
	case ModeHintScroll:
		style.BackgroundColor = a.config.Hints.ScrollHints.BackgroundColor
		style.TextColor = a.config.Hints.ScrollHints.TextColor
		style.MatchedTextColor = a.config.Hints.ScrollHints.MatchedTextColor
		style.BorderColor = a.config.Hints.ScrollHints.BorderColor
	case ModeContextMenu:
		style.BackgroundColor = a.config.Hints.ContextMenuHints.BackgroundColor
		style.TextColor = a.config.Hints.ContextMenuHints.TextColor
		style.MatchedTextColor = a.config.Hints.ContextMenuHints.MatchedTextColor
		style.BorderColor = a.config.Hints.ContextMenuHints.BorderColor
	}

	return style
}

// handleKeyPress routes key presses to the appropriate handler based on mode
func (a *App) handleKeyPress(key string) {
	// Only handle keys when in an active mode
	if a.currentMode == ModeIdle {
		return
	}

	// Handle Escape key to exit any mode
	if key == "\x1b" || key == "escape" {
		a.logger.Debug("Escape pressed in mode", zap.String("mode", a.getCurrModeString()))
		a.exitMode()
		return
	}

	// Handle Tab key
	if key == "\t" {
		a.logger.Debug("Tab pressed in mode", zap.String("mode", a.getCurrModeString()))

		// If we are not in scroll mode yet, we can tab to scroll the current cursor position directly
		if a.currentMode == ModeHintScroll && !a.canScroll {
			a.logger.Debug("Switching to scroll mode from scroll hints")
			a.canScroll = true
			a.hintOverlay.Clear()
			a.showScroll()
			return
		}
	}

	a.handleHintKey(key)
}

// handleHintKey handles key presses in hint mode
func (a *App) handleHintKey(key string) {
	if a.selectedHint != nil {
		a.handleContextMenuKey(key)
		return
	}
	if a.canScroll {
		a.handleScrollKey(key)
		return
	}

	if hint, ok := a.hintManager.HandleInput(key); ok {
		switch a.currentMode {
		case ModeHintLeftClick:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.LeftClick(a.config.Hints.LeftClickHints.RestoreCursor); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintRightClick:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.RightClick(a.config.Hints.RightClickHints.RestoreCursor); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintDoubleClick:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.DoubleClick(a.config.Hints.DoubleClickHints.RestoreCursor); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintTripleClick:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.TripleClick(a.config.Hints.TripleClickHints.RestoreCursor); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintMouseUp:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.LeftMouseUp(false); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintMouseDown:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.LeftMouseDown(); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()

			a.activateHintMode(ModeHintMouseUp)
		case ModeHintMiddleClick:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.MiddleClick(a.config.Hints.MiddleClickHints.RestoreCursor); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintMoveMouse:
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.GoToPosition(); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeHintScroll:
			a.canScroll = true
			// Enter scroll mode - scroll to element
			a.logger.Info("Hint selected, start scrolling", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.GoToPosition(); err != nil {
				a.logger.Error("Failed to scroll to element", zap.Error(err))
			}

			a.hintOverlay.Clear()
			a.showScroll()
		case ModeContextMenu:
			// Store the hint and wait for action selection
			a.selectedHint = hint
			a.logger.Info("Hint selected, choose an action", zap.String("label", a.hintManager.GetInput()))

			// Clear all hints and show action menu at the hint location
			a.hintOverlay.Clear()
			a.showContextMenu(hint)
		}
		return
	}

	// Otherwise, key was handled by the hint manager
}

// showContextMenu displays the action selection menu at the hint location
func (a *App) showContextMenu(hint *hints.Hint) {
	// Hardcoded action keys (tied to UI labels)
	actions := []struct {
		key   string
		label string
	}{
		{"l", "eft"},
		{"r", "ight"},
		{"d", "ouble"},
		{"t", "riple"},
		{"h", "old"},
		{"H", " unhold"},
		{"m", "iddle"},
		{"g", "oto pos"},
	}

	var formattedActions []string
	for _, a := range actions {
		formattedActions = append(formattedActions, fmt.Sprintf("[%s]%s", a.key, a.label))
	}

	menuActions := strings.Join(formattedActions, "\n")

	actionHints := make([]*hints.Hint, 0, len(actions))
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

	contextMenuStyle := a.getStyleForMode(ModeContextMenu)

	if err := a.hintOverlay.DrawHintsWithoutArrow(actionHints, contextMenuStyle); err != nil {
		a.logger.Error("Failed to draw action menu", zap.Error(err))
	}
}

// handleContextMenuKey handles action key presses after a hint is selected
func (a *App) handleContextMenuKey(key string) {
	if a.selectedHint == nil {
		return
	}

	// Handle backspace/delete to go back to hint selection
	if key == "\x7f" || key == "delete" || key == "backspace" {
		a.logger.Debug("Backspace pressed in action mode, returning to hint selection")

		// Clear selected hint and reset hint manager
		a.selectedHint = nil
		a.hintManager.Reset()
		return
	}

	hint := a.selectedHint
	a.logger.Info("Action key pressed", zap.String("key", key))

	var err error
	var successCallback func()
	// Hardcoded action keys (tied to UI labels)
	switch key {
	case "l": // Left click
		a.logger.Info("Performing left click", zap.String("label", hint.Label))
		err = hint.Element.Element.LeftClick(a.config.Hints.LeftClickHints.RestoreCursor)
	case "r": // Right click
		a.logger.Info("Performing right click", zap.String("label", hint.Label))
		err = hint.Element.Element.RightClick(a.config.Hints.RightClickHints.RestoreCursor)
	case "d": // Double click
		a.logger.Info("Performing double click", zap.String("label", hint.Label))
		err = hint.Element.Element.DoubleClick(a.config.Hints.DoubleClickHints.RestoreCursor)
	case "t": // Triple click
		a.logger.Info("Performing triple click", zap.String("label", hint.Label))
		err = hint.Element.Element.TripleClick(a.config.Hints.TripleClickHints.RestoreCursor)
	case "h": // Hold mouse
		a.logger.Info("Performing hold mouse", zap.String("label", hint.Label))
		err = hint.Element.Element.LeftMouseDown()

		// Activate mouse up mode straight away
		successCallback = func() {
			a.activateHintMode(ModeHintMouseUp)
		}
	case "H": // Release mouse
		a.logger.Info("Performing release hold mouse", zap.String("label", hint.Label))
		err = hint.Element.Element.LeftMouseUp(false)
	case "m": // Middle click
		a.logger.Info("Performing middle click", zap.String("label", hint.Label))
		err = hint.Element.Element.MiddleClick(a.config.Hints.MiddleClickHints.RestoreCursor)
	case "g": // Move mouse to a position
		a.logger.Info("Performing go to position", zap.String("label", hint.Label))
		err = hint.Element.Element.GoToPosition()
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
		switch byteVal {
		case 4: // Ctrl+D
			a.logger.Info("Ctrl+D detected - half page down")
			err = a.scrollController.ScrollDownHalfPage()
			goto done
		case 21: // Ctrl+U
			a.logger.Info("Ctrl+U detected - half page up")
			err = a.scrollController.ScrollUpHalfPage()
			goto done
		}
	}

	// Regular keys
	switch key {
	case "j":
		err = a.scrollController.ScrollDown()
	case "k":
		err = a.scrollController.ScrollUp()
	case "h":
		err = a.scrollController.ScrollLeft()
	case "l":
		err = a.scrollController.ScrollRight()
	case "g": // gg for top (need to press twice)
		if a.lastScrollKey == "g" {
			a.logger.Info("gg detected - scroll to top")
			err = a.scrollController.ScrollToTop()
			a.lastScrollKey = ""
			goto done
		} else {
			a.logger.Info("First g pressed, press again for top")
			a.lastScrollKey = "g"
			return
		}
	case "G": // Shift+G for bottom
		a.logger.Info("G key detected - scroll to bottom")
		err = a.scrollController.ScrollToBottom()
		a.lastScrollKey = ""
	case "\t":
		a.logger.Info("? key detected - show hint again")
		a.activateHintMode(ModeHintScroll)
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		a.lastScrollKey = ""
		return
	}

	// Reset last key for most commands
	a.lastScrollKey = ""

done:
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
}

func (a *App) drawScrollHighlightBorder() {
	window := accessibility.GetFrontmostWindow()

	if window == nil {
		a.logger.Debug("No frontmost window")
		return
	}
	defer window.Release()

	bounds := window.GetScrollBounds()

	a.hintOverlay.DrawScrollHighlight(
		bounds.Min.X, bounds.Min.Y,
		bounds.Dx(), bounds.Dy(),
		a.config.Scroll.HighlightColor,
		a.config.Scroll.HighlightWidth,
	)
}

// exitMode exits the current mode
func (a *App) exitMode() {
	if a.currentMode == ModeIdle {
		return
	}

	a.logger.Info("Exiting current mode", zap.String("mode", a.getCurrModeString()))

	// If we are in mouse up mode, remove the mouse up action to prevent further draggings
	if a.currentMode == ModeHintMouseUp {
		a.logger.Info("Detected MouseUp mode, removing mouse up event...")
		err := accessibility.GetFrontmostWindow().LeftMouseUp(true)
		if err != nil {
			a.logger.Error("Failed to remove mouse up", zap.Error(err))
		}
	}

	if a.hintManager != nil {
		a.hintManager.Reset()
	}
	a.selectedHint = nil
	a.canScroll = false

	// Then clear and hide the overlay after all potential visual generators are cleaned
	a.hintOverlay.Clear()
	a.hintOverlay.Hide()

	// Finally, disable event tap and set idle mode
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	// Update mode after all cleanup is done
	a.currentMode = ModeIdle
	a.logger.Debug("Mode transition complete",
		zap.String("to", "idle"))
}

func getModeString(mode Mode) string {
	switch mode {
	case ModeIdle:
		return "idle"
	case ModeHintLeftClick:
		return "left_click"
	case ModeHintRightClick:
		return "right_click"
	case ModeHintDoubleClick:
		return "double_click"
	case ModeHintTripleClick:
		return "triple_click"
	case ModeHintMouseUp:
		return "mouse_up"
	case ModeHintMouseDown:
		return "mouse_down"
	case ModeHintMiddleClick:
		return "middle_click"
	case ModeHintMoveMouse:
		return "move_mouse"
	case ModeHintScroll:
		return "scroll"
	case ModeContextMenu:
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
func (a *App) logModeActivation(mode Mode) {
	hintCount := len(a.hintManager.GetHints())

	switch mode {
	case ModeHintLeftClick:
		a.logger.Info("Hint mode with left click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintRightClick:
		a.logger.Info("Hint mode with right click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintDoubleClick:
		a.logger.Info("Hint mode with double click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintTripleClick:
		a.logger.Info("Hint mode with triple click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintMouseUp:
		a.logger.Info("Hint mode with mouse up activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintMouseDown:
		a.logger.Info("Hint mode with mouse down activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintMiddleClick:
		a.logger.Info("Hint mode with middle click activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintMoveMouse:
		a.logger.Info("Hint mode with move mouse activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeHintScroll:
		a.logger.Info("Scroll mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom")
	case ModeContextMenu:
		a.logger.Info("Context menu mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to choose an action")
	}
}
