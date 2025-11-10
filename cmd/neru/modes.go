package main

import (
	"fmt"
	"image"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
)

type HintsContext struct {
	selectedHint  *hints.Hint
	canScroll     bool
	lastScrollKey string
}

type GridContext struct {
	currentAction     Action
	gridInstance      **grid.Grid
	gridOverlay       **grid.GridOverlay
	canScroll         bool
	lastScrollKey     string
	contextMenuActive bool
	selectedPoint     image.Point
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
	if mode == ModeGrid {
		a.activateGridMode(action)
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

func (a *App) activateGridMode(action Action) {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring grid mode activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	actionString := getActionString(action)
	a.logger.Info("Activating grid mode", zap.String("action", actionString))

	a.exitMode() // Exit current mode first

	if actionString == "unknown" {
		a.logger.Warn("Unknown action, ignoring")
		return
	}

	// Generate grid cells
	if err := a.setupGrid(action); err != nil {
		a.logger.Error("Failed to setup grid", zap.Error(err), zap.String("action", actionString))
		return
	}

	// Update mode and enable event tap
	a.currentMode = ModeGrid
	a.gridCtx.canScroll = false
	a.gridCtx.currentAction = action

	// Enable event tap to capture keys
	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.logger.Info("Grid mode activated", zap.String("action", actionString))
	a.logger.Info("Type a grid label to select a location")
}

// setupGrid generates grid and draws it
func (a *App) setupGrid(action Action) error {
	// Create grid with active screen bounds (screen containing mouse cursor)
	// This ensures proper multi-monitor support
	bounds := bridge.GetActiveScreenBounds()
	characters := a.config.Grid.Characters
	if strings.TrimSpace(characters) == "" {
		characters = a.config.Hints.HintCharacters
	}
	minSize := a.config.Grid.MinCellSize
	if minSize <= 0 {
		minSize = 40
	}
	maxSize := a.config.Grid.MaxCellSize
	if maxSize <= 0 {
		maxSize = 200
	}
	gridInstance := grid.NewGrid(characters, minSize, maxSize, bounds)
	*a.gridCtx.gridInstance = gridInstance

	// Grid overlay already created in NewApp - update its config and use it
	(*a.gridCtx.gridOverlay).UpdateConfig(a.config.Grid)

	// Subgrid configuration and keys (fallback to grid characters)
	keys := strings.TrimSpace(a.config.Grid.SublayerKeys)
	if keys == "" {
		keys = a.config.Grid.Characters
	}
	subRows := a.config.Grid.SubgridRows
	subCols := a.config.Grid.SubgridCols
	if subRows < 1 {
		subRows = 3
	}
	if subCols < 1 {
		subCols = 3
	}

	// Initialize manager with the new grid
	a.gridManager = grid.NewManager(gridInstance, subRows, subCols, keys, func() {
		// Update matches only (no full redraw)
		(*a.gridCtx.gridOverlay).UpdateMatches(a.gridManager.GetInput())
	}, func(cell *grid.Cell) {
		// Draw 3x3 subgrid inside selected cell
		(*a.gridCtx.gridOverlay).ShowSubgrid(cell)
	})
	a.gridRouter = grid.NewRouter(a.gridManager)

	// Draw initial grid
	if err := (*a.gridCtx.gridOverlay).Draw(gridInstance, ""); err != nil {
		return fmt.Errorf("failed to draw grid: %w", err)
	}

	(*a.gridCtx.gridOverlay).Show()
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
	case ModeGrid:
		res := a.gridRouter.RouteKey(key, a.gridCtx.canScroll, a.gridCtx.currentAction == ActionScroll)
		if res.Exit {
			a.exitMode()
			return
		}
		if res.SwitchToScroll {
			a.logger.Debug("Switching to scroll mode from scroll grid")
			a.gridCtx.canScroll = true
			a.showGridScroll()
			return
		}

		// If grid context menu is active, route to context menu handler
		if a.gridCtx.contextMenuActive {
			a.handleGridContextMenuKey(key)
			return
		}
		// If grid scroll is active, route to grid scroll handler
		if a.gridCtx.canScroll {
			a.handleGridScrollKey(key)
			return
		}

		// Complete coordinate entered - perform action
		if res.Complete {
			targetPoint := res.TargetPoint
			switch a.gridCtx.currentAction {
			case ActionLeftClick:
				a.logger.Info("Grid left click", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.LeftClickAtPoint(targetPoint, a.config.Hints.LeftClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click", zap.Error(err))
				}
				a.exitMode()
			case ActionRightClick:
				a.logger.Info("Grid right click", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.RightClickAtPoint(targetPoint, a.config.Hints.RightClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click", zap.Error(err))
				}
				a.exitMode()
			case ActionDoubleClick:
				a.logger.Info("Grid double click", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.DoubleClickAtPoint(targetPoint, a.config.Hints.DoubleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click", zap.Error(err))
				}
				a.exitMode()
			case ActionTripleClick:
				a.logger.Info("Grid triple click", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.TripleClickAtPoint(targetPoint, a.config.Hints.TripleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click", zap.Error(err))
				}
				a.exitMode()
			case ActionMouseDown:
				a.logger.Info("Grid mouse down", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.LeftMouseDownAtPoint(targetPoint); err != nil {
					a.logger.Error("Failed to mouse down", zap.Error(err))
				}
				// Immediately switch to mouse up to allow release selection
				a.exitMode()
				a.activateMode(ModeGrid, ActionMouseUp)
			case ActionMouseUp:
				a.logger.Info("Grid mouse up", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.LeftMouseUpAtPoint(targetPoint); err != nil {
					a.logger.Error("Failed to mouse up", zap.Error(err))
				}
				a.exitMode()
			case ActionMiddleClick:
				a.logger.Info("Grid middle click", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				if err := accessibility.MiddleClickAtPoint(targetPoint, a.config.Hints.MiddleClickHints.RestoreCursor); err != nil {
					a.logger.Error("Failed to click", zap.Error(err))
				}
				a.exitMode()
			case ActionMoveMouse:
				a.logger.Info("Grid move mouse", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				accessibility.MoveMouseToPoint(targetPoint)
				a.exitMode()
			case ActionScroll:
				// Grid-specific scroll: move mouse to target and draw scroll border via grid overlay, stay in grid mode
				a.gridCtx.canScroll = true
				a.logger.Info("Grid scroll", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				accessibility.MoveMouseToPoint(targetPoint)
				a.showGridScroll()
			case ActionContextMenu:
				// Show context menu overlay at target point (no hint generation)
				a.logger.Info("Grid context menu", zap.Int("x", targetPoint.X), zap.Int("y", targetPoint.Y))
				// Ensure grid overlay doesn't occlude the menu
				gridOverlay := *a.gridCtx.gridOverlay
				gridOverlay.Clear()
				gridOverlay.Hide()
				// Draw target indicator dot
				dotRadius := 4.0
				dotColor := a.config.Hints.ContextMenuHints.BackgroundColor
				dotBorderColor := a.config.Hints.ContextMenuHints.BorderColor
				dotBorderWidth := float64(a.config.Hints.BorderWidth)
				if err := a.hintOverlay.DrawTargetDot(targetPoint.X, targetPoint.Y, dotRadius, dotColor, dotBorderColor, dotBorderWidth); err != nil {
					a.logger.Error("Failed to draw target dot", zap.Error(err))
				}
				// Build and draw context menu label at the point
				menuActions := hints.BuildContextMenuLabel()
				contextMenuStyle := hints.BuildStyleForAction(a.config.Hints, getActionString(ActionContextMenu))
				actionHints := []*hints.Hint{{
					Label:    menuActions,
					Element:  nil,
					Position: image.Point{X: targetPoint.X, Y: targetPoint.Y + 5},
				}}
				if err := a.hintOverlay.DrawHintsWithoutArrow(actionHints, contextMenuStyle); err != nil {
					a.logger.Error("Failed to draw context menu", zap.Error(err))
				}
				// Ensure hint overlay is visible
				a.hintOverlay.Show()
				// Activate grid context-menu sub-state
				a.gridCtx.contextMenuActive = true
				a.gridCtx.selectedPoint = targetPoint
			}
			return
		}
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

func (a *App) showGridScroll() {
	win := accessibility.GetFrontmostWindow()
	if win == nil {
		return
	}
	defer win.Release()
	bounds := win.GetScrollBounds()
	gridOverlay := *a.gridCtx.gridOverlay

	gridOverlay.Clear()
	gridOverlay.DrawScrollHighlight(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy(), a.config.Scroll.HighlightColor, a.config.Scroll.HighlightWidth)
}

func (a *App) handleGridScrollKey(key string) {
	// Exit grid scroll with Escape
	if key == "\x1b" || key == "escape" {
		a.exitMode()
		return
	}
	var err error
	// Check for control characters
	if len(key) == 1 {
		byteVal := key[0]
		if byteVal == 4 || byteVal == 21 { // Ctrl+D / Ctrl+U
			op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
			if ok {
				a.gridCtx.lastScrollKey = ""
				switch op {
				case "half_down":
					err = a.scrollController.ScrollDownHalfPage()
				case "half_up":
					err = a.scrollController.ScrollUpHalfPage()
				}
			}
		}
	}
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
		return
	}
	// Regular keys
	switch key {
	case "j":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "down" {
			err = a.scrollController.ScrollDown()
		}
	case "k":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "up" {
			err = a.scrollController.ScrollUp()
		}
	case "h":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "left" {
			err = a.scrollController.ScrollLeft()
		}
	case "l":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if !ok {
			return
		}
		if op == "right" {
			err = a.scrollController.ScrollRight()
		}
	case "g":
		op, newLast, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if !ok {
			a.gridCtx.lastScrollKey = newLast
			return
		}
		if op == "top" {
			err = a.scrollController.ScrollToTop()
			a.gridCtx.lastScrollKey = ""
		}
	case "G":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if ok && op == "bottom" {
			err = a.scrollController.ScrollToBottom()
			a.gridCtx.lastScrollKey = ""
		}
	case "\t":
		op, _, ok := scroll.ParseKey(key, a.gridCtx.lastScrollKey)
		if ok && op == "tab" {
			a.logger.Info("Tab key detected - show grid again")
			a.activateMode(ModeGrid, ActionScroll)
		}
	default:
		a.gridCtx.lastScrollKey = ""
		return
	}

	// Reset last key for most commands
	a.gridCtx.lastScrollKey = ""

	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
}

func (a *App) handleGridContextMenuKey(key string) {
	// Exit context menu with Escape
	if key == "\x1b" || key == "escape" {
		// Clear menu overlay
		a.hintOverlay.Clear()
		// Leave grid mode
		a.exitMode()
		return
	}

	center := a.gridCtx.selectedPoint
	var err error
	var successCallback func()

	a.logger.Info("Grid context menu key pressed", zap.String("key", key))

	action := hints.ParseContextMenuKey(key)
	switch action {
	case "left_click":
		err = accessibility.LeftClickAtPoint(center, a.config.Hints.LeftClickHints.RestoreCursor)
	case "right_click":
		err = accessibility.RightClickAtPoint(center, a.config.Hints.RightClickHints.RestoreCursor)
	case "double_click":
		err = accessibility.DoubleClickAtPoint(center, a.config.Hints.DoubleClickHints.RestoreCursor)
	case "triple_click":
		err = accessibility.TripleClickAtPoint(center, a.config.Hints.TripleClickHints.RestoreCursor)
	case "mouse_down":
		err = accessibility.LeftMouseDownAtPoint(center)
		// After mouse down, prepare to allow release selection in grid
		successCallback = func() {
			// Clear menu overlay then switch to grid mouse up mode
			a.hintOverlay.Clear()
			a.activateMode(ModeGrid, ActionMouseUp)
		}
	case "mouse_up":
		err = accessibility.LeftMouseUpAtPoint(center)
	case "middle_click":
		err = accessibility.MiddleClickAtPoint(center, a.config.Hints.MiddleClickHints.RestoreCursor)
	case "move_mouse":
		accessibility.MoveMouseToPoint(center)
		// no error to report
		goto done
	default:
		a.logger.Debug("Unknown grid context action", zap.String("key", key))
		return
	}

	if err != nil {
		a.logger.Error("Failed to perform action", zap.Error(err))
	}

done:
	// Exit grid mode and clear overlay
	a.hintOverlay.Clear()
	a.exitMode()

	if successCallback != nil {
		successCallback()
	}
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
	case ModeGrid:
		// If we are in mouse up action, remove the mouse up to prevent further dragging
		if a.currentAction == ActionMouseUp {
			a.logger.Info("Detected MouseUp action, removing mouse up event...")
			err := accessibility.LeftMouseUp()
			if err != nil {
				a.logger.Error("Failed to remove mouse up", zap.Error(err))
			}
		}
		if a.gridManager != nil {
			a.gridManager.Reset()
		}
		// Reset grid sub-states
		if a.gridCtx != nil {
			a.gridCtx.canScroll = false
			a.gridCtx.lastScrollKey = ""
			a.gridCtx.contextMenuActive = false
		}
		// Hide overlays
		if a.gridCtx != nil && a.gridCtx.gridOverlay != nil && *a.gridCtx.gridOverlay != nil {
			(*a.gridCtx.gridOverlay).Hide()
		}
		// Also clear any context menu drawn on hint overlay
		a.hintOverlay.Clear()
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
	case ModeGrid:
		return "grid"
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
