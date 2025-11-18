package app

import (
	"fmt"
	"image"
	"math"
	"runtime"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/overlay"
	"github.com/y3owk1n/neru/internal/scroll"
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

// activateMode activates a mode with a given action (for hints mode).
func (a *App) activateMode(mode Mode) {
	if mode == ModeIdle {
		// Explicit idle transition
		a.exitMode()
		return
	}
	if mode == ModeHints {
		a.activateHintMode()
		return
	}
	if mode == ModeGrid {
		a.activateGridMode()
		return
	}
	// Unknown or unsupported mode
	a.logger.Warn("Unknown mode", zap.String("mode", getModeString(mode)))
}

// activateHintMode activates hint mode, with an option to preserve action mode state.
func (a *App) activateHintMode() {
	a.activateHintModeInternal(false)
}

// activateHintModeInternal activates hint mode with option to preserve action mode state.
func (a *App) activateHintModeInternal(preserveActionMode bool) {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring hint mode activation")
		return
	}
	// Respect mode enable flag
	if !a.config.Hints.Enabled {
		a.logger.Debug("Hints mode disabled by config, ignoring activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.captureInitialCursorPosition()

	action := ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating hint mode", zap.String("action", actionString))

	// Only exit mode if not preserving action mode state
	if !preserveActionMode {
		a.exitMode()
	}

	if actionString == unknownAction {
		a.logger.Warn("Unknown action string, ignoring")
		return
	}

	// Always resize overlay to the active screen (where mouse is) before collecting elements
	// This ensures proper positioning when switching between multiple displays
	a.overlayManager.ResizeToActiveScreenSync()
	a.hintOverlayNeedsRefresh = false

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Collect elements based on mode
	elements := a.collectElements()
	if len(elements) == 0 {
		a.logger.Warn("No elements found for action", zap.String("action", actionString))
		return
	}

	// Generate and setup hints
	err := a.setupHints(elements)
	if err != nil {
		a.logger.Error("Failed to setup hints", zap.Error(err), zap.String("action", actionString))
		return
	}

	a.hintsCtx.selectedHint = nil
	a.setModeHints()
}

// setupHints generates hints and draws them with appropriate styling.
func (a *App) setupHints(elements []*accessibility.TreeNode) error {
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)

	// Generate and normalize hints
	hintList, err := a.generateAndNormalizeHints(elements)
	if err != nil {
		return err
	}

	// Set up hints in the hint manager
	hintCollection := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hintCollection)

	// Draw hints with mode-specific styling
	drawErr := a.renderer.DrawHints(hintList)
	if drawErr != nil {
		return fmt.Errorf("failed to draw hints: %w", drawErr)
	}
	a.renderer.Show()

	// Log performance metrics
	a.logHintsSetupPerformance(hintList, msBefore)

	return nil
}

// generateAndNormalizeHints generates hints and normalizes their positions.
func (a *App) generateAndNormalizeHints(elements []*accessibility.TreeNode) ([]*hints.Hint, error) {
	// Get active screen bounds to calculate offset for normalization
	screenBounds := bridge.GetActiveScreenBounds()
	screenOffsetX := screenBounds.Min.X
	screenOffsetY := screenBounds.Min.Y

	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hints: %w", err)
	}

	// Normalize hint positions to window-local coordinates
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	for _, hint := range hintList {
		pos := hint.GetPosition()
		hint.Position.X = pos.X - screenOffsetX
		hint.Position.Y = pos.Y - screenOffsetY
	}

	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())
	filtered := make([]*hints.Hint, 0, len(hintList))
	for _, h := range hintList {
		if h.IsVisible(localBounds) {
			filtered = append(filtered, h)
		}
	}

	return filtered, nil
}

// logHintsSetupPerformance logs performance metrics for hint setup.
func (a *App) logHintsSetupPerformance(hintList []*hints.Hint, msBefore runtime.MemStats) {
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	a.logger.Info("Hints setup perf",
		zap.Int("hints", len(hintList)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))
}

func (a *App) activateGridMode() {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring grid mode activation")
		return
	}
	// Respect mode enable flag
	if !a.config.Grid.Enabled {
		a.logger.Debug("Grid mode disabled by config, ignoring activation")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.captureInitialCursorPosition()

	action := ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating grid mode", zap.String("action", actionString))

	a.exitMode() // Exit current mode first

	// Always resize overlay to the active screen (where mouse is) before drawing grid
	// This ensures proper positioning when switching between multiple displays
	a.renderer.ResizeActive()
	a.gridOverlayNeedsRefresh = false

	err := a.setupGrid()
	if err != nil {
		a.logger.Error("Failed to setup grid", zap.Error(err), zap.String("action", actionString))
		return
	}

	a.setModeGrid()

	a.logger.Info("Grid mode activated", zap.String("action", actionString))
	a.logger.Info("Type a grid label to select a location")
}

// setupGrid generates grid and draws it.
func (a *App) setupGrid() error {
	// Create grid with active screen bounds (screen containing mouse cursor)
	// This ensures proper multi-monitor support
	gridInstance := a.createGridInstance()

	// Update grid overlay configuration
	a.updateGridOverlayConfig()

	// Ensure the overlay is properly sized for the active screen
	a.overlayManager.ResizeToActiveScreenSync()

	// Reset the grid manager state when setting up the grid
	if a.gridManager != nil {
		a.gridManager.Reset()
	}

	// Initialize manager with the new grid
	a.initializeGridManager(gridInstance)

	a.gridRouter = grid.NewRouter(a.gridManager, a.logger)

	// Draw initial grid
	initErr := a.renderer.DrawGrid(gridInstance, "")
	if initErr != nil {
		return fmt.Errorf("failed to draw grid: %w", initErr)
	}
	a.renderer.Show()

	return nil
}

// createGridInstance creates a new grid instance with proper bounds and characters.
func (a *App) createGridInstance() *grid.Grid {
	screenBounds := bridge.GetActiveScreenBounds()

	// Normalize bounds to window-local coordinates (0,0 origin)
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	bounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	characters := a.config.Grid.Characters
	if strings.TrimSpace(characters) == "" {
		characters = a.config.Hints.HintCharacters
	}
	gridInstance := grid.NewGrid(characters, bounds, a.logger)
	*a.gridCtx.gridInstance = gridInstance

	return gridInstance
}

// updateGridOverlayConfig updates the grid overlay configuration.
func (a *App) updateGridOverlayConfig() {
	// Grid overlay already created in NewApp - update its config and use it
	(*a.gridCtx.gridOverlay).UpdateConfig(a.config.Grid)
}

// initializeGridManager initializes the grid manager with the new grid instance.
func (a *App) initializeGridManager(gridInstance *grid.Grid) {
	// Subgrid configuration and keys (fallback to grid characters): always 3x3
	keys := strings.TrimSpace(a.config.Grid.SublayerKeys)
	if keys == "" {
		keys = a.config.Grid.Characters
	}
	const subRows = 3
	const subCols = 3

	// Initialize manager with the new grid
	a.gridManager = grid.NewManager(gridInstance, subRows, subCols, keys, func(forceRedraw bool) {
		// Update matches only (no full redraw)
		input := a.gridManager.GetInput()

		// special case to handle only when exiting subgrid
		if forceRedraw {
			a.renderer.Clear()
			gridErr := a.renderer.DrawGrid(gridInstance, input)
			if gridErr != nil {
				return
			}
			a.renderer.Show()
		}

		// Set hideUnmatched based on whether we have input and the config setting
		hideUnmatched := a.config.Grid.HideUnmatched && len(input) > 0
		a.renderer.SetHideUnmatched(hideUnmatched)
		a.renderer.UpdateGridMatches(input)
	}, func(cell *grid.Cell) {
		// Draw 3x3 subgrid inside selected cell
		a.renderer.ShowSubgrid(cell)
	}, a.logger)
}

// handleActiveKey dispatches key events by current mode.
func (a *App) handleKeyPress(key string) {
	// If in idle mode, check if we should handle scroll keys
	if a.currentMode == ModeIdle {
		// Handle escape to exit standalone scroll
		if key == "\x1b" || key == "escape" {
			a.logger.Info("Exiting standalone scroll mode")
			a.overlayManager.Clear()
			a.overlayManager.Hide()
			if a.eventTap != nil {
				a.eventTap.Disable()
			}
			a.isScrollingActive = false
			a.idleScrollLastKey = "" // Reset scroll state
			return
		}
		// Try to handle scroll keys with generic handler using persistent state
		// If it's not a scroll key, it will just be ignored
		a.handleGenericScrollKey(key, &a.idleScrollLastKey)
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
	switch a.currentMode {
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
	switch a.currentMode {
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
	switch a.currentMode {
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

// handleGenericScrollKey handles scroll keys in a generic way.
func (a *App) handleGenericScrollKey(key string, lastScrollKey *string) {
	// Local storage for scroll state if not provided
	var localLastKey string
	if lastScrollKey == nil {
		lastScrollKey = &localLastKey
	}

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
		if a.handleControlScrollKey(key, *lastScrollKey, lastScrollKey) {
			return
		}
	}

	// Regular keys
	a.logger.Debug(
		"Entering switch statement",
		zap.String("key", key),
		zap.String("keyHex", fmt.Sprintf("%#v", key)),
	)
	switch key {
	case "j", "k", "h", "l":
		err = a.handleDirectionalScrollKey(key, *lastScrollKey)
	case "g": // gg for top (need to press twice)
		operation, newLast, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			a.logger.Info("First g pressed, press again for top")
			*lastScrollKey = newLast
			return
		}
		if operation == "top" {
			a.logger.Info("gg detected - scroll to top")
			err = a.scrollController.ScrollToTop()
			*lastScrollKey = ""
			goto done
		}
	case "G": // Shift+G for bottom
		operation, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if ok && operation == "bottom" {
			a.logger.Info("G key detected - scroll to bottom")
			err = a.scrollController.ScrollToBottom()
			*lastScrollKey = ""
		}
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		*lastScrollKey = ""
		return
	}

	// Reset last key for most commands
	*lastScrollKey = ""

done:
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
}

// handleControlScrollKey handles control character scroll keys.
func (a *App) handleControlScrollKey(key string, lastKey string, lastScrollKey *string) bool {
	byteVal := key[0]
	a.logger.Info("Checking control char", zap.Uint8("byte", byteVal))
	// Only handle Ctrl+D / Ctrl+U here; let Tab (9) and other keys fall through to switch
	if byteVal == 4 || byteVal == 21 {
		op, _, ok := scroll.ParseKey(key, lastKey, a.logger)
		if ok {
			*lastScrollKey = ""
			switch op {
			case "half_down":
				a.logger.Info("Ctrl+D detected - half page down")
				return a.scrollController.ScrollDownHalfPage() != nil
			case "half_up":
				a.logger.Info("Ctrl+U detected - half page up")
				return a.scrollController.ScrollUpHalfPage() != nil
			}
		}
	}
	return false
}

// handleDirectionalScrollKey handles directional scroll keys (j, k, h, l).
func (a *App) handleDirectionalScrollKey(key string, lastKey string) error {
	op, _, ok := scroll.ParseKey(key, lastKey, a.logger)
	if !ok {
		return nil
	}

	switch key {
	case "j":
		if op == "down" {
			return a.scrollController.ScrollDown()
		}
	case "k":
		if op == "up" {
			return a.scrollController.ScrollUp()
		}
	case "h":
		if op == "left" {
			return a.scrollController.ScrollLeft()
		}
	case "l":
		if op == "right" {
			return a.scrollController.ScrollRight()
		}
	}
	return nil
}

// drawHintsActionHighlight draws a highlight border around the active screen for hints action mode.
func (a *App) drawHintsActionHighlight() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	// Get active screen bounds
	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	// Draw action highlight using renderer
	a.renderer.DrawActionHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)

	a.logger.Debug("Drawing hints action highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))
}

// drawGridActionHighlight draws a highlight border around the active screen for grid action mode.
func (a *App) drawGridActionHighlight() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	// Get active screen bounds
	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	// Draw action highlight using renderer
	a.renderer.DrawActionHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)

	a.logger.Debug("Drawing grid action highlight",
		zap.Int("x", localBounds.Min.X),
		zap.Int("y", localBounds.Min.Y),
		zap.Int("width", localBounds.Dx()),
		zap.Int("height", localBounds.Dy()))
}

func (a *App) drawScrollHighlightBorder() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.renderer.ResizeActive()

	// Get active screen bounds
	screenBounds := bridge.GetActiveScreenBounds()
	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())

	// Draw scroll highlight using renderer
	a.renderer.DrawScrollHighlight(
		localBounds.Min.X,
		localBounds.Min.Y,
		localBounds.Dx(),
		localBounds.Dy(),
	)
	a.overlayManager.SwitchTo(overlay.ModeScroll)
}

// exitMode exits the current mode.
func (a *App) exitMode() {
	if a.currentMode == ModeIdle {
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
	switch a.currentMode {
	case ModeHints:
		a.cleanupHintsMode()
	case ModeGrid:
		a.cleanupGridMode()
	default:
		a.cleanupDefaultMode()
	}
}

// cleanupHintsMode handles cleanup for hints mode.
func (a *App) cleanupHintsMode() {
	// Reset action mode state
	a.hintsCtx.inActionMode = false

	if a.hintManager != nil {
		a.hintManager.Reset()
	}
	a.hintsCtx.selectedHint = nil

	// Clear and hide overlay for hints
	a.overlayManager.Clear()
	a.overlayManager.Hide()

	// Also clear and hide action overlay
	if overlay.Get() != nil {
		overlay.Get().Clear()
		overlay.Get().Hide()
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	a.logger.Info("Hints cleanup mem",
		zap.Uint64("alloc_bytes", ms.Alloc),
		zap.Uint64("sys_bytes", ms.Sys))
}

// cleanupGridMode handles cleanup for grid mode.
func (a *App) cleanupGridMode() {
	// Reset action mode state
	a.gridCtx.inActionMode = false

	if a.gridManager != nil {
		a.gridManager.Reset()
	}
	// Hide overlays
	a.logger.Info("Hiding grid overlay")
	a.overlayManager.Hide()

	// Also clear and hide action overlay
	a.overlayManager.Clear()
	a.overlayManager.Hide()
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
	a.currentMode = ModeIdle
	a.logger.Debug("Mode transition complete",
		zap.String("to", "idle"))
	a.overlayManager.SwitchTo(overlay.ModeIdle)

	// If a hotkey refresh was deferred while in an active mode, perform it now
	if a.hotkeyRefreshPending {
		a.hotkeyRefreshPending = false
		go a.refreshHotkeysForAppOrCurrent("")
	}
}

// handleCursorRestoration handles cursor position restoration on exit.
func (a *App) handleCursorRestoration() {
	shouldRestore := a.shouldRestoreCursorOnExit()
	if shouldRestore {
		currentBounds := bridge.GetActiveScreenBounds()
		target := computeRestoredPosition(a.initialCursorPos, a.initialScreenBounds, currentBounds)
		accessibility.MoveMouseToPoint(target)
	}
	a.cursorRestoreCaptured = false
	a.isScrollingActive = false
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
	default:
		return "unknown"
	}
}

// getCurrModeString returns the current mode as a string.
func (a *App) getCurrModeString() string {
	return getModeString(a.currentMode)
}

func (a *App) captureInitialCursorPosition() {
	if a.cursorRestoreCaptured {
		return
	}
	a.initialCursorPos = accessibility.GetCurrentCursorPosition()
	a.initialScreenBounds = bridge.GetActiveScreenBounds()
	a.cursorRestoreCaptured = true
}

func (a *App) shouldRestoreCursorOnExit() bool {
	if a.config == nil {
		return false
	}
	if !a.config.General.RestoreCursorPosition {
		return false
	}
	if !a.cursorRestoreCaptured {
		return false
	}
	if a.isScrollingActive {
		return false
	}
	if a.skipCursorRestoreOnce {
		a.skipCursorRestoreOnce = false
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
		act = string(ActionNameLeftClick)
	case a.config.Action.RightClickKey:
		a.logger.Info(mode + " action: Right click")
		act = string(ActionNameRightClick)
	case a.config.Action.MiddleClickKey:
		a.logger.Info(mode + " action: Middle click")
		act = string(ActionNameMiddleClick)
	case a.config.Action.MouseDownKey:
		a.logger.Info(mode + " action: Mouse down")
		act = string(ActionNameMouseDown)
	case a.config.Action.MouseUpKey:
		a.logger.Info(mode + " action: Mouse up")
		act = string(ActionNameMouseUp)
	default:
		a.logger.Debug("Unknown "+mode+" action key", zap.String("key", key))
		return
	}
	err := performActionAtPoint(act, cursorPos)
	if err != nil {
		a.logger.Error("Failed to perform action", zap.Error(err))
	}
}

// handleHintsActionKey handles action keys when in hints action mode.
func (a *App) handleHintsActionKey(key string) {
	a.handleActionKey(key, "Hints")
}

// handleGridActionKey handles action keys when in grid action mode.
func (a *App) handleGridActionKey(key string) {
	a.handleActionKey(key, "Grid")
}
