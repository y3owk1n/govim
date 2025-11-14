package main

import (
	"fmt"
	"image"
	"runtime"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/bridge"
	"github.com/y3owk1n/neru/internal/grid"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/scroll"
	"go.uber.org/zap"
)

type HintsContext struct {
	selectedHint *hints.Hint
}

type GridContext struct {
	gridInstance **grid.Grid
	gridOverlay  **grid.GridOverlay
}

// activateMode activates a mode with a given action (for hints mode)
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

func (a *App) activateHintMode() {
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

	action := ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating hint mode", zap.String("action", actionString))

	a.exitMode()

	if actionString == "unknown" {
		a.logger.Warn("Unknown action string, ignoring")
		return
	}

	// Always resize overlay to the active screen (where mouse is) before collecting elements
	// This ensures proper positioning when switching between multiple displays
	if a.hintOverlay != nil {
		// Use synchronous resize with timeout to prevent hanging
		a.hintOverlay.ResizeToActiveScreenSync()
		a.hintOverlayNeedsRefresh = false
	}

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Collect elements based on mode
	elements := a.collectElements()
	if len(elements) == 0 {
		a.logger.Warn("No elements found for action", zap.String("action", actionString))
		return
	}

	// Generate and setup hints
	if err := a.setupHints(elements); err != nil {
		a.logger.Error("Failed to setup hints", zap.Error(err), zap.String("action", actionString))
		return
	}

	// Update mode and enable event tap
	a.currentMode = ModeHints
	a.hintsCtx.selectedHint = nil

	// Enable event tap to capture keys (must be last to ensure proper initialization)
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
}

// setupHints generates hints and draws them with appropriate styling
func (a *App) setupHints(elements []*accessibility.TreeNode) error {
	var msBefore runtime.MemStats
	runtime.ReadMemStats(&msBefore)
	// Get active screen bounds to calculate offset for normalization
	screenBounds := bridge.GetActiveScreenBounds()
	screenOffsetX := screenBounds.Min.X
	screenOffsetY := screenBounds.Min.Y

	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		return fmt.Errorf("failed to generate hints: %w", err)
	}

	// Normalize hint positions to window-local coordinates
	// The overlay window is positioned at the screen origin, but the view uses local coordinates
	for _, hint := range hintList {
		hint.Position.X -= screenOffsetX
		hint.Position.Y -= screenOffsetY
	}

	localBounds := image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())
	filtered := make([]*hints.Hint, 0, len(hintList))
	for _, h := range hintList {
		if h.IsVisible(localBounds) {
			filtered = append(filtered, h)
		}
	}
	hintList = filtered

	// Set up hints in the hint manager
	hintCollection := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hintCollection)

	// Draw hints with mode-specific styling
	style := hints.BuildStyle(a.config.Hints)
	if err := a.hintOverlay.DrawHintsWithStyle(hintList, style); err != nil {
		return fmt.Errorf("failed to draw hints: %w", err)
	}

	a.hintOverlay.Show()
	var msAfter runtime.MemStats
	runtime.ReadMemStats(&msAfter)
	a.logger.Info("Hints setup perf",
		zap.Int("hints", len(hintList)),
		zap.Uint64("alloc_bytes_delta", msAfter.Alloc-msBefore.Alloc),
		zap.Uint64("sys_bytes_delta", msAfter.Sys-msBefore.Sys))
	return nil
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

	action := ActionMoveMouse
	actionString := getActionString(action)
	a.logger.Info("Activating grid mode", zap.String("action", actionString))

	a.exitMode() // Exit current mode first

	// Always resize overlay to the active screen (where mouse is) before drawing grid
	// This ensures proper positioning when switching between multiple displays
	if a.gridCtx != nil && a.gridCtx.gridOverlay != nil {
		gridOverlay := *a.gridCtx.gridOverlay

		// Use synchronous resize with timeout to prevent hanging
		gridOverlay.ResizeToActiveScreenSync()
		a.gridOverlayNeedsRefresh = false
	}

	if err := a.setupGrid(); err != nil {
		a.logger.Error("Failed to setup grid", zap.Error(err), zap.String("action", actionString))
		return
	}

	// Update mode and enable event tap
	a.currentMode = ModeGrid

	// Enable event tap to capture keys
	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.logger.Info("Grid mode activated", zap.String("action", actionString))
	a.logger.Info("Type a grid label to select a location")
}

// setupGrid generates grid and draws it
func (a *App) setupGrid() error {
	// Create grid with active screen bounds (screen containing mouse cursor)
	// This ensures proper multi-monitor support
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

	// Grid overlay already created in NewApp - update its config and use it
	(*a.gridCtx.gridOverlay).UpdateConfig(a.config.Grid)

	// Get style for current action
	gridStyle := grid.BuildStyle(a.config.Grid)

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
			(*a.gridCtx.gridOverlay).Clear()

			if err := (*a.gridCtx.gridOverlay).Draw(gridInstance, input, gridStyle); err != nil {
				return
			}

			(*a.gridCtx.gridOverlay).Show()
		}

		// Set hideUnmatched based on whether we have input and the config setting
		hideUnmatched := a.config.Grid.HideUnmatched && len(input) > 0
		(*a.gridCtx.gridOverlay).SetHideUnmatched(hideUnmatched)

		(*a.gridCtx.gridOverlay).UpdateMatches(input)
	}, func(cell *grid.Cell) {
		// Draw 3x3 subgrid inside selected cell
		(*a.gridCtx.gridOverlay).ShowSubgrid(cell, gridStyle)
	}, a.logger)
	a.gridRouter = grid.NewRouter(a.gridManager, a.logger)

	// Draw initial grid
	if err := (*a.gridCtx.gridOverlay).Draw(gridInstance, "", gridStyle); err != nil {
		return fmt.Errorf("failed to draw grid: %w", err)
	}

	(*a.gridCtx.gridOverlay).Show()
	return nil
}

// handleActiveKey dispatches key events by current mode
func (a *App) handleKeyPress(key string) {
	// If in idle mode, check if we should handle scroll keys
	if a.currentMode == ModeIdle {
		// Handle escape to exit standalone scroll
		if key == "\x1b" || key == "escape" {
			a.logger.Info("Exiting standalone scroll mode")
			a.scrollOverlay.Clear()
			a.scrollOverlay.Hide()
			if a.eventTap != nil {
				a.eventTap.Disable()
			}
			a.idleScrollLastKey = "" // Reset scroll state
			return
		}
		// Try to handle scroll keys with generic handler using persistent state
		// If it's not a scroll key, it will just be ignored
		a.handleGenericScrollKey(key, &a.idleScrollLastKey)
		return
	}

	// Explicitly dispatch by current mode
	switch a.currentMode {
	case ModeHints:
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
			center := image.Point{X: info.Position.X + info.Size.X/2, Y: info.Position.Y + info.Size.Y/2}

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

			a.logger.Info("Grid move mouse", zap.Int("x", absolutePoint.X), zap.Int("y", absolutePoint.Y))
			accessibility.MoveMouseToPoint(absolutePoint)

			// No need to exit grid mode, just let it going
			// a.exitMode()

			return
		}
	}
}

// handleGenericScrollKey handles scroll keys in a generic way
// Returns true if exit was requested
func (a *App) handleGenericScrollKey(key string, lastScrollKey *string) bool {
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
		byteVal := key[0]
		a.logger.Info("Checking control char", zap.Uint8("byte", byteVal))
		// Only handle Ctrl+D / Ctrl+U here; let Tab (9) and other keys fall through to switch
		if byteVal == 4 || byteVal == 21 {
			op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
			if ok {
				*lastScrollKey = ""
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
	a.logger.Debug("Entering switch statement", zap.String("key", key), zap.String("keyHex", fmt.Sprintf("%#v", key)))
	switch key {
	case "j":
		op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			return false
		}
		if op == "down" {
			err = a.scrollController.ScrollDown()
		}
	case "k":
		op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			return false
		}
		if op == "up" {
			err = a.scrollController.ScrollUp()
		}
	case "h":
		op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			return false
		}
		if op == "left" {
			err = a.scrollController.ScrollLeft()
		}
	case "l":
		op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			return false
		}
		if op == "right" {
			err = a.scrollController.ScrollRight()
		}
	case "g": // gg for top (need to press twice)
		op, newLast, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if !ok {
			a.logger.Info("First g pressed, press again for top")
			*lastScrollKey = newLast
			return false
		}
		if op == "top" {
			a.logger.Info("gg detected - scroll to top")
			err = a.scrollController.ScrollToTop()
			*lastScrollKey = ""
			goto done
		}
	case "G": // Shift+G for bottom
		op, _, ok := scroll.ParseKey(key, *lastScrollKey, a.logger)
		if ok && op == "bottom" {
			a.logger.Info("G key detected - scroll to bottom")
			err = a.scrollController.ScrollToBottom()
			*lastScrollKey = ""
		}
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		*lastScrollKey = ""
		return false
	}

	// Reset last key for most commands
	*lastScrollKey = ""

done:
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
	return false
}

func (a *App) drawScrollHighlightBorder() {
	// Resize overlay to active screen (where mouse cursor is) for multi-monitor support
	a.scrollOverlay.ResizeToActiveScreen()
	a.scrollOverlay.DrawScrollHighlight()
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
		if a.hintManager != nil {
			a.hintManager.Reset()
		}
		a.hintsCtx.selectedHint = nil

		// Clear and hide overlay for hints
		a.hintOverlay.Clear()
		a.hintOverlay.Hide()
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		a.logger.Info("Hints cleanup mem",
			zap.Uint64("alloc_bytes", ms.Alloc),
			zap.Uint64("sys_bytes", ms.Sys))
	case ModeGrid:
		if a.gridManager != nil {
			a.gridManager.Reset()
		}
		// Hide overlays
		if a.gridCtx != nil && a.gridCtx.gridOverlay != nil && *a.gridCtx.gridOverlay != nil {
			a.logger.Info("Hiding grid overlay")
			(*a.gridCtx.gridOverlay).Hide()
		}
	default:
		// No domain-specific cleanup for other modes yet
	}

	// Clear scroll overlay
	a.scrollOverlay.Clear()

	// Disable event tap when leaving active modes
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	// Update mode after all cleanup is done
	a.currentMode = ModeIdle
	a.logger.Debug("Mode transition complete",
		zap.String("to", "idle"))

	// If a hotkey refresh was deferred while in an active mode, perform it now
	if a.hotkeyRefreshPending {
		a.hotkeyRefreshPending = false
		go a.refreshHotkeysForAppOrCurrent("")
	}
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

// getCurrModeString returns the current mode as a string
func (a *App) getCurrModeString() string {
	return getModeString(a.currentMode)
}
