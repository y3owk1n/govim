package main

import (
	"fmt"
	"image"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/config"
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
func (a *App) getStyleForMode(mode Mode) config.HintsConfig {
	style := a.config.Hints

	switch mode {
	case ModeHintWithActions:
		style.BackgroundColor = a.config.Hints.ActionBackgroundColor
		style.TextColor = a.config.Hints.ActionTextColor
		style.MatchedTextColor = a.config.Hints.ActionMatchedTextColor
		style.BorderColor = a.config.Hints.ActionBorderColor
	case ModeScroll:
		style.BackgroundColor = a.config.Hints.ScrollHintsBackgroundColor
		style.TextColor = a.config.Hints.ScrollHintsTextColor
		style.MatchedTextColor = a.config.Hints.ScrollHintsMatchedTextColor
		style.BorderColor = a.config.Hints.ScrollHintsBorderColor
	case ModeHint:
		style.BackgroundColor = a.config.Hints.BackgroundColor
		style.TextColor = a.config.Hints.TextColor
		style.MatchedTextColor = a.config.Hints.MatchedTextColor
		style.BorderColor = a.config.Hints.BorderColor
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
		if a.currentMode == ModeScroll && !a.canScroll {
			a.logger.Debug("Switching to scroll mode from scroll hints")
			a.canScroll = true
			a.hintOverlay.Clear()
			a.showScroll()
			return
		}

		// If we are in hint mode, we can tab to hint with actions
		if a.currentMode == ModeHint {
			a.logger.Debug("Switching to hint with actions mode from hints")
			a.activateHintMode(ModeHintWithActions)
			return
		}

		// If we are in hint with actions mode, we can tab to hint
		if a.currentMode == ModeHintWithActions && a.selectedHint == nil {
			a.logger.Debug("Switching to hint mode from hint with actions")
			a.activateHintMode(ModeHint)
			return
		}
	}

	a.handleHintKey(key)
}

// handleHintKey handles key presses in hint mode
func (a *App) handleHintKey(key string) {
	// If we have a selected hint (waiting for action), handle action selection
	if a.selectedHint != nil {
		a.handleActionKey(key)
		return
	}

	if a.canScroll {
		a.handleScrollKey(key)
		return
	}

	if hint, ok := a.hintManager.HandleInput(key); ok {
		switch a.currentMode {
		case ModeHintWithActions:
			// Store the hint and wait for action selection
			a.selectedHint = hint
			a.logger.Info("Hint selected, choose action", zap.String("label", a.hintManager.GetInput()))

			// Clear all hints and show action menu at the hint location
			a.hintOverlay.Clear()
			a.showActionMenu(hint)
		case ModeHint:
			// Direct click mode - click immediately
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.LeftClick(a.config.General.RestorePosAfterLeftClick); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
		case ModeScroll:
			a.canScroll = true
			// Enter scroll mode - scroll to element
			a.logger.Info("Hint selected, start scrolling", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.GoToPosition(); err != nil {
				a.logger.Error("Failed to scroll to element", zap.Error(err))
			}

			a.hintOverlay.Clear()
			a.showScroll()
		}
		return
	}

	// Otherwise, key was handled by the hint manager
}

// showActionMenu displays the action selection menu at the hint location
func (a *App) showActionMenu(hint *hints.Hint) {
	// Hardcoded action keys (tied to UI labels)
	actions := []struct {
		key   string
		label string
	}{
		{"l", "eft"},
		{"r", "ight"},
		{"d", "ouble"},
		{"m", "iddle"},
		{"g", "oto pos"},
	}

	// Create hints for each action, positioned horizontally with consistent gaps
	actionHints := make([]*hints.Hint, 0, len(actions))
	baseX := hint.Position.X
	baseY := hint.Position.Y

	// Estimate width per character (approximate for monospace font at size 11)
	charWidth := 7.0
	gapBetweenHints := 8.0 // Consistent gap between hint boxes

	currentX := float64(baseX)

	for _, action := range actions {
		// Format: [key]label (e.g., "[l]eft")
		actionLabel := fmt.Sprintf("[%s]%s", action.key, action.label)

		actionHint := &hints.Hint{
			Label:    actionLabel,
			Element:  hint.Element,
			Position: image.Point{X: int(currentX), Y: baseY},
		}
		actionHints = append(actionHints, actionHint)

		// Calculate width of this hint box (text width + padding * 2)
		textWidth := float64(len(actionLabel)) * charWidth
		padding := 3.0 * 2 // padding on both sides
		hintBoxWidth := textWidth + padding

		// Move to next position (current box width + gap)
		currentX += hintBoxWidth + gapBetweenHints
	}

	actionStyle := a.getStyleForMode(ModeHintWithActions)
	// Use smaller sizing for action hints
	actionStyle.FontSize = 11    // Smaller font
	actionStyle.Padding = 3      // Less padding
	actionStyle.BorderRadius = 3 // Smaller border radius

	// Draw all action hints without arrows and with custom style
	if err := a.hintOverlay.DrawHintsWithoutArrow(actionHints, actionStyle); err != nil {
		a.logger.Error("Failed to draw action menu", zap.Error(err))
	}
}

// handleActionKey handles action key presses after a hint is selected
func (a *App) handleActionKey(key string) {
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
	// Hardcoded action keys (tied to UI labels)
	switch key {
	case "l": // Left click
		a.logger.Info("Performing left click", zap.String("label", hint.Label))
		err = hint.Element.Element.LeftClick(a.config.General.RestorePosAfterLeftClick)
	case "r": // Right click
		a.logger.Info("Performing right click", zap.String("label", hint.Label))
		err = hint.Element.Element.RightClick(a.config.General.RestorePosAfterRightClick)
	case "d": // Double click
		a.logger.Info("Performing double click", zap.String("label", hint.Label))
		err = hint.Element.Element.DoubleClick(a.config.General.RestorePosAfterDoubleClick)
	case "m": // Middle click
		a.logger.Info("Performing middle click", zap.String("label", hint.Label))
		err = hint.Element.Element.MiddleClick(a.config.General.RestorePosAfterMiddleClick)
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
}

// showActionMenu displays the action selection menu at the hint location
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
		a.activateHintMode(ModeScroll)
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
	case ModeHint:
		return "hint"
	case ModeHintWithActions:
		return "hint_with_actions"
	case ModeScroll:
		return "scroll"
	default:
		return "unknown"
	}
}

// getCurrModeString returns the current mode as a string
func (a *App) getCurrModeString() string {
	switch a.currentMode {
	case ModeIdle:
		return "idle"
	case ModeHint:
		return "hint"
	case ModeHintWithActions:
		return "hint_with_actions"
	case ModeScroll:
		return "scroll"
	default:
		return "unknown"
	}
}
