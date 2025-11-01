package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/govim/internal/accessibility"
	_ "github.com/y3owk1n/govim/internal/bridge" // Import for CGo compilation
	"github.com/y3owk1n/govim/internal/cli"
	"github.com/y3owk1n/govim/internal/config"
	"github.com/y3owk1n/govim/internal/eventtap"
	"github.com/y3owk1n/govim/internal/hints"
	"github.com/y3owk1n/govim/internal/hotkeys"
	"github.com/y3owk1n/govim/internal/ipc"
	"github.com/y3owk1n/govim/internal/logger"
	"github.com/y3owk1n/govim/internal/scroll"
	"go.uber.org/zap"
)

// Mode represents the current application mode
type Mode int

const (
	ModeIdle Mode = iota
	ModeHint
	ModeHintWithActions
	ModeScroll
)

// App represents the main application
type App struct {
	config *config.Config
	// ConfigPath holds the path the daemon was started with (if any)
	ConfigPath       string
	logger           *zap.Logger
	hotkeyManager    *hotkeys.Manager
	hintGenerator    *hints.Generator
	hintOverlay      *hints.Overlay
	scrollController *scroll.Controller
	eventTap         *eventtap.EventTap
	ipcServer        *ipc.Server
	currentMode      Mode
	currentHints     *hints.HintCollection
	hintInput        string
	lastScrollKey    string
	selectedHint     *hints.Hint
	enabled          bool
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	if err := logger.Init(cfg.Logging.LogLevel, cfg.Logging.LogFile, cfg.Logging.StructuredLogging); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Get()

	// Check accessibility permissions
	if cfg.General.AccessibilityCheckOnStart {
		if !accessibility.CheckAccessibilityPermissions() {
			log.Warn("Accessibility permissions not granted. Please grant permissions in System Settings.")
			fmt.Println("⚠️  GoVim requires Accessibility permissions to function.")
			fmt.Println("Please go to: System Settings → Privacy & Security → Accessibility")
			fmt.Println("and enable GoVim.")
			return nil, fmt.Errorf("accessibility permissions required")
		}
	}

	// Apply clickable and scrollable roles from config
	log.Info("Applying clickable roles",
		zap.Int("count", len(cfg.Accessibility.ClickableRoles)),
		zap.Strings("roles", cfg.Accessibility.ClickableRoles))
	accessibility.SetClickableRoles(cfg.Accessibility.ClickableRoles)

	log.Info("Applying scrollable roles",
		zap.Int("count", len(cfg.Accessibility.ScrollableRoles)),
		zap.Strings("roles", cfg.Accessibility.ScrollableRoles))
	accessibility.SetScrollableRoles(cfg.Accessibility.ScrollableRoles)

	// Create hotkey manager
	hotkeyMgr := hotkeys.NewManager(log)
	hotkeys.SetGlobalManager(hotkeyMgr)

	// Create hint generator
	hintGen := hints.NewGenerator(
		cfg.General.HintCharacters,
		cfg.General.HintStyle,
		cfg.Performance.MaxHintsDisplayed,
	)

	// Create hint overlay
	hintOverlay, err := hints.NewOverlay(cfg.Hints)
	if err != nil {
		return nil, fmt.Errorf("failed to create overlay: %w", err)
	}

	// Create scroll controller
	scrollCtrl := scroll.NewController(cfg.Scroll, log)

	app := &App{
		config:           cfg,
		logger:           log,
		hotkeyManager:    hotkeyMgr,
		hintGenerator:    hintGen,
		hintOverlay:      hintOverlay,
		scrollController: scrollCtrl,
		currentMode:      ModeIdle,
		enabled:          true,
		hintInput:        "",
	}

	// Create event tap for capturing keys in modes
	app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - key capture won't work")
	} else {
		// Ensure event tap is disabled initially (only enable in active modes)
		app.eventTap.Disable()
	}

	// Create IPC server
	ipcServer, err := ipc.NewServer(app.handleIPCCommand, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC server: %w", err)
	}
	app.ipcServer = ipcServer

	return app, nil
}

// Run runs the application
func (a *App) Run() error {
	a.logger.Info("Starting GoVim")

	// Start IPC server
	a.ipcServer.Start()
	a.logger.Info("IPC server started")

	// Register hotkeys
	if err := a.registerHotkeys(); err != nil {
		return fmt.Errorf("failed to register hotkeys: %w", err)
	}

	a.logger.Info("GoVim is running")
	fmt.Println("✓ GoVim is running")

	// Print configured hotkeys
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateHintMode); key != "" {
		fmt.Printf("  Hint mode (direct): %s\n", key)
	}
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateHintModeWithActions); key != "" {
		fmt.Printf("  Hint mode (with actions): %s\n", key)
	}
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateScrollMode); key != "" {
		fmt.Printf("  Scroll mode: %s\n", key)
	}

	// Wait for interrupt signal with force-quit support
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// First signal: graceful shutdown
	<-sigChan
	a.logger.Info("Received shutdown signal, starting graceful shutdown...")
	fmt.Println("\n⚠️  Shutting down gracefully... (press Ctrl+C again to force quit)")

	// Start cleanup in goroutine
	done := make(chan struct{})
	go func() {
		// Quit systray to exit the event loop
		systray.Quit()
		close(done)
	}()

	// Wait for cleanup or second signal
	select {
	case <-done:
		a.logger.Info("Graceful shutdown completed")
		return nil
	case <-sigChan:
		// Second signal: force quit
		a.logger.Warn("Received second signal, forcing shutdown")
		fmt.Println("⚠️  Force quitting...")
		os.Exit(1)
	case <-time.After(10 * time.Second):
		// Timeout: force quit
		a.logger.Error("Shutdown timeout exceeded, forcing shutdown")
		fmt.Println("⚠️  Shutdown timeout, force quitting...")
		os.Exit(1)
	}

	return nil
}

// registerHotkeys registers all global hotkeys
func (a *App) registerHotkeys() error {
	// Hint mode hotkey (direct click)
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateHintMode); key != "" {
		a.logger.Info("Registering hint mode hotkey", zap.String("key", key))
		if _, err := a.hotkeyManager.Register(key, func() {
			a.activateHintMode(false)
		}); err != nil {
			return fmt.Errorf("failed to register hint mode hotkey: %w", err)
		}
	}

	// Hint mode with actions hotkey
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateHintModeWithActions); key != "" {
		a.logger.Info("Registering hint mode with actions hotkey", zap.String("key", key))
		if _, err := a.hotkeyManager.Register(key, func() {
			a.activateHintMode(true)
		}); err != nil {
			return fmt.Errorf("failed to register hint mode with actions hotkey: %w", err)
		}
	}

	// Scroll mode hotkey
	if key := strings.TrimSpace(a.config.Hotkeys.ActivateScrollMode); key != "" {
		a.logger.Info("Registering scroll mode hotkey", zap.String("key", key))
		if _, err := a.hotkeyManager.Register(key, a.activateScrollMode); err != nil {
			return fmt.Errorf("failed to register scroll mode hotkey: %w", err)
		}
	}

	// Note: Escape key for exiting modes is hardcoded in handleKeyPress, not registered as global hotkey
	return nil
}

// updateRolesForCurrentApp updates clickable and scrollable roles based on the current focused app
func (a *App) updateRolesForCurrentApp() {
	// Get the focused application
	focusedApp := accessibility.GetFocusedApplication()
	if focusedApp == nil {
		a.logger.Debug("No focused application, using global roles only")
		// Use global roles
		accessibility.SetClickableRoles(a.config.Accessibility.ClickableRoles)
		accessibility.SetScrollableRoles(a.config.Accessibility.ScrollableRoles)
		return
	}
	defer focusedApp.Release()

	// Get bundle ID
	bundleID := focusedApp.GetBundleIdentifier()
	if bundleID == "" {
		a.logger.Debug("Could not get bundle ID, using global roles only")
		// Use global roles
		accessibility.SetClickableRoles(a.config.Accessibility.ClickableRoles)
		accessibility.SetScrollableRoles(a.config.Accessibility.ScrollableRoles)
		return
	}

	// Get merged roles for this app
	clickableRoles := a.config.GetClickableRolesForApp(bundleID)
	scrollableRoles := a.config.GetScrollableRolesForApp(bundleID)

	a.logger.Debug("Updating roles for current app",
		zap.String("bundle_id", bundleID),
		zap.Int("clickable_count", len(clickableRoles)),
		zap.Int("scrollable_count", len(scrollableRoles)))

	// Apply the merged roles
	accessibility.SetClickableRoles(clickableRoles)
	accessibility.SetScrollableRoles(scrollableRoles)
}

// activateHintMode activates hint mode
func (a *App) activateHintMode(withActions bool) {
	if !a.enabled {
		a.logger.Debug("GoVim is disabled, ignoring hint mode activation")
		return
	}
	if a.currentMode == ModeHint || a.currentMode == ModeHintWithActions {
		a.logger.Debug("Hint mode already active")
		return
	}

	a.logger.Info("Activating hint mode")
	a.exitMode() // Exit current mode first

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	// Enable Electron accessibility if needed before scanning
	if a.config.Accessibility.ElectronSupport.Enable {
		accessibility.EnsureElectronSupport(a.config.Accessibility.ElectronSupport.AdditionalBundles)
	}

	// Get clickable elements
	roles := accessibility.GetClickableRoles()
	a.logger.Debug("Scanning for clickable elements",
		zap.Strings("roles", roles))

	elements, err := accessibility.GetClickableElements()
	if err != nil {
		a.logger.Error("Failed to get clickable elements", zap.Error(err))
		return
	}

	a.logger.Info("Found clickable elements", zap.Int("count", len(elements)))

	// Optionally include menu bar elements
	if a.config.Hints.Menubar {
		if mbElems, merr := accessibility.GetMenuBarClickableElements(); merr == nil {
			elements = append(elements, mbElems...)
			a.logger.Debug("Included menubar elements", zap.Int("count", len(mbElems)))
		} else {
			a.logger.Warn("Failed to get menubar elements", zap.Error(merr))
		}
	}

	// Optionally include Dock elements
	if a.config.Hints.Dock {
		if dockElems, derr := accessibility.GetDockClickableElements(); derr == nil {
			elements = append(elements, dockElems...)
			a.logger.Debug("Included dock elements", zap.Int("count", len(dockElems)))
		} else {
			a.logger.Warn("Failed to get dock elements", zap.Error(derr))
		}
	}

	if len(elements) == 0 {
		a.logger.Warn("No clickable elements found")
		return
	}

	// Generate hints
	hintList, err := a.hintGenerator.Generate(elements)
	if err != nil {
		a.logger.Error("Failed to generate hints", zap.Error(err))
		return
	}

	// Create hint collection
	a.currentHints = hints.NewHintCollection(hintList)

	// Draw hints
	if err := a.hintOverlay.DrawHints(hintList); err != nil {
		a.logger.Error("Failed to draw hints", zap.Error(err))
		return
	}

	a.hintOverlay.Show()

	if withActions {
		a.currentMode = ModeHintWithActions
		a.logger.Info("Hint mode with actions activated", zap.Int("hints", len(hintList)))
		a.logger.Info("Type a hint label, then choose action: l=left, r=right, d=double, m=middle")
	} else {
		a.currentMode = ModeHint
		a.logger.Info("Hint mode activated", zap.Int("hints", len(hintList)))
		a.logger.Info("Type a hint label to click the element")
	}

	// Enable event tap to capture keys
	a.hintInput = ""
	a.selectedHint = nil
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
}

// handleKeyPress routes key presses to the appropriate handler based on mode
func (a *App) handleKeyPress(key string) {
	// Only handle keys when in an active mode
	if a.currentMode == ModeIdle {
		return
	}

	// Handle Escape key to exit any mode
	if key == "\x1b" || key == "escape" {
		a.logger.Debug("Escape pressed in mode", zap.String("mode", a.getModeString()))
		a.exitMode()
		return
	}

	// Route to appropriate handler
	switch a.currentMode {
	case ModeHint, ModeHintWithActions:
		a.handleHintKey(key)
	case ModeScroll:
		a.handleScrollKey(key)
	}
}

// handleHintKey handles key presses in hint mode
func (a *App) handleHintKey(key string) {
	// If we have a selected hint (waiting for action), handle action selection
	if a.selectedHint != nil {
		a.handleActionKey(key)
		return
	}

	// Handle backspace/delete to remove last character
	if key == "\x7f" || key == "delete" || key == "backspace" {
		if len(a.hintInput) > 0 {
			// Remove last character
			a.hintInput = a.hintInput[:len(a.hintInput)-1]
			a.logger.Debug("Backspace pressed, hint input", zap.String("input", a.hintInput))

			// Redraw hints with updated filter
			var filtered []*hints.Hint
			if a.hintInput == "" {
				// Show all hints if input is empty
				filtered = a.currentHints.GetHints()
			} else {
				filtered = a.currentHints.FilterByPrefix(a.hintInput)
			}

			// Update matched prefix for filtered hints and redraw
			for _, hint := range filtered {
				hint.MatchedPrefix = a.hintInput
			}
			a.hintOverlay.Clear()
			if err := a.hintOverlay.DrawHints(filtered); err != nil {
				a.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}
		return
	}

	// Ignore non-letter keys
	if len(key) != 1 || !isLetter(key[0]) {
		return
	}

	// Accumulate input (convert to uppercase to match hints)
	a.hintInput += strings.ToUpper(key)
	a.logger.Debug("Hint input", zap.String("input", a.hintInput))

	// Check if any hints start with this input
	filtered := a.currentHints.FilterByPrefix(a.hintInput)

	if len(filtered) == 0 {
		// No matches - reset
		a.logger.Debug("No matching hints, resetting")
		a.hintInput = ""
		return
	}

	// Update matched prefix for filtered hints and redraw
	for _, hint := range filtered {
		hint.MatchedPrefix = a.hintInput
	}
	a.hintOverlay.Clear()
	if err := a.hintOverlay.DrawHints(filtered); err != nil {
		a.logger.Error("Failed to redraw hints", zap.Error(err))
	}

	// If exactly one match and input matches the full label
	if len(filtered) == 1 && filtered[0].Label == a.hintInput {
		hint := filtered[0]

		if a.currentMode == ModeHintWithActions {
			// Store the hint and wait for action selection
			a.selectedHint = hint
			a.logger.Info("Hint selected, choose action", zap.String("label", a.hintInput))

			// Clear all hints and show action menu at the hint location
			a.hintOverlay.Clear()
			a.showActionMenu(hint)
			return
		} else {
			// Direct click mode - click immediately
			a.logger.Info("Clicking element", zap.String("label", a.hintInput))
			if err := hint.Element.Element.Click(); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
			return
		}
	}

	// Otherwise, keep collecting input
	a.logger.Debug("Partial match", zap.Int("matches", len(filtered)))
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

	// Create smaller style for action hints
	actionStyle := a.config.Hints
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

		// Clear selected hint
		a.selectedHint = nil

		// Remove the last character from hintInput (e.g., "ABD" -> "AB")
		if len(a.hintInput) > 0 {
			a.hintInput = a.hintInput[:len(a.hintInput)-1]
			a.logger.Debug("Removed last character from hint input", zap.String("input", a.hintInput))
		}

		// Redraw hints with updated input filter
		var filtered []*hints.Hint
		if a.hintInput == "" {
			filtered = a.currentHints.GetHints()
		} else {
			filtered = a.currentHints.FilterByPrefix(a.hintInput)
		}

		// Update matched prefix for filtered hints and redraw
		for _, hint := range filtered {
			hint.MatchedPrefix = a.hintInput
		}
		a.hintOverlay.Clear()
		if err := a.hintOverlay.DrawHints(filtered); err != nil {
			a.logger.Error("Failed to redraw hints", zap.Error(err))
		}
		return
	}

	hint := a.selectedHint
	a.logger.Info("Action key pressed", zap.String("key", key))

	var err error
	// Hardcoded action keys (tied to UI labels)
	switch key {
	case "l": // Left click
		a.logger.Info("Performing left click", zap.String("label", hint.Label))
		err = hint.Element.Element.Click()
	case "r": // Right click
		a.logger.Info("Performing right click", zap.String("label", hint.Label))
		err = hint.Element.Element.RightClick()
	case "d": // Double click
		a.logger.Info("Performing double click", zap.String("label", hint.Label))
		err = hint.Element.Element.DoubleClick()
	case "m": // Middle click
		a.logger.Info("Performing middle click", zap.String("label", hint.Label))
		err = hint.Element.Element.MiddleClick()
	default:
		a.logger.Debug("Unknown action key, ignoring", zap.String("key", key))
		return
	}

	if err != nil {
		a.logger.Error("Failed to perform action", zap.Error(err))
	}

	a.exitMode()
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
	// Remove plain d/u - only use Ctrl+D/U
	// case "d": // Removed - use Ctrl+D instead
	// case "u": // Removed - use Ctrl+U instead
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
	case "\t": // Tab (next scroll area)
		a.switchScrollArea(true)
		a.lastScrollKey = ""
		return
	case "1", "2", "3", "4", "5", "6", "7", "8", "9": // Number keys
		a.switchScrollAreaByNumber(key)
		a.lastScrollKey = ""
		return
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

// switchScrollArea switches to next/prev scroll area
func (a *App) switchScrollArea(next bool) {
	detector := a.scrollController.GetDetector()
	var newArea *scroll.ScrollArea

	if next {
		newArea = detector.CycleActiveArea()
	} else {
		newArea = detector.CyclePrevArea()
	}

	if newArea != nil {
		a.logger.Info("Switched scroll area", zap.Int("index", newArea.Index+1))
		a.updateScrollHighlight(newArea)
	}
}

// switchScrollAreaByNumber switches to a specific scroll area by number
func (a *App) switchScrollAreaByNumber(key string) {
	// Validate key is a digit
	if len(key) != 1 || key[0] < '0' || key[0] > '9' {
		a.logger.Debug("Invalid number key", zap.String("key", key))
		return
	}

	number := int(key[0] - '0')
	detector := a.scrollController.GetDetector()
	newArea := detector.SetActiveByNumber(number)

	if newArea != nil {
		a.logger.Info("Switched to scroll area", zap.Int("number", number))
		a.updateScrollHighlight(newArea)
	} else {
		a.logger.Debug("No scroll area at number", zap.Int("number", number))
	}
}

// updateScrollHighlight updates the scroll area highlight
func (a *App) updateScrollHighlight(area *scroll.ScrollArea) {
	bounds := area.Bounds
	a.hintOverlay.DrawScrollHighlight(
		bounds.Min.X, bounds.Min.Y,
		bounds.Dx(), bounds.Dy(),
		a.config.Scroll.HighlightColor,
		a.config.Scroll.HighlightWidth,
	)
	a.hintOverlay.Show()
}

// activateScrollMode activates scroll mode
func (a *App) activateScrollMode() {
	if !a.enabled {
		a.logger.Debug("GoVim is disabled, ignoring scroll mode activation")
		return
	}

	if a.currentMode == ModeScroll {
		a.logger.Debug("Scroll mode already active")
		return
	}

	a.logger.Info("Activating scroll mode")
	a.exitMode() // Exit current mode first

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

	if a.config.Accessibility.ElectronSupport.Enable {
		accessibility.EnsureElectronSupport(a.config.Accessibility.ElectronSupport.AdditionalBundles)
	}

	a.logger.Info("Initializing scroll controller")
	// Activate scroll mode
	if err := a.scrollController.Initialize(); err != nil {
		a.logger.Error("Failed to initialize scroll controller", zap.Error(err))
		return
	}

	a.logger.Info("Getting scroll areas")
	// Get all scroll areas
	areas := a.scrollController.GetAllAreas()
	if len(areas) == 0 {
		a.logger.Warn("No scrollable areas found")
		return
	}

	a.logger.Info("Drawing scroll area labels", zap.Int("count", len(areas)))
	// Draw scroll highlights with numbers for all areas
	a.drawScrollAreaLabels(areas)

	a.logger.Info("Setting mode to scroll")
	a.currentMode = ModeScroll
	a.logger.Info("Scroll mode activated", zap.Int("areas", len(areas)))
	a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom")
	a.logger.Info("Tab to switch areas, or press 1-9 to jump to area")

	a.logger.Info("Enabling event tap")
	// Enable event tap to capture scroll keys
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
	a.logger.Info("Scroll mode activation complete")
}

// drawScrollAreaLabels draws number labels on all scroll areas
func (a *App) drawScrollAreaLabels(areas []*scroll.ScrollArea) {
	// Clear existing overlay
	a.hintOverlay.Clear()

	// Create hints for each scroll area number
	areaHints := make([]*hints.Hint, 0, len(areas))
	for i, area := range areas {
		// Position number at top-left of scroll area
		hint := &hints.Hint{
			Label:    fmt.Sprintf("%d", i+1),
			Position: image.Point{X: area.Bounds.Min.X + 10, Y: area.Bounds.Min.Y + 10},
			Size:     image.Point{X: 30, Y: 30},
		}
		areaHints = append(areaHints, hint)
	}

	// Draw the number hints
	if len(areaHints) > 0 {
		if err := a.hintOverlay.DrawHints(areaHints); err != nil {
			a.logger.Error("Failed to draw hints", zap.Error(err))
		}
	}

	// Draw highlight on active area
	activeArea := a.scrollController.GetDetector().GetActiveArea()
	if activeArea != nil {
		bounds := activeArea.Bounds
		a.hintOverlay.DrawScrollHighlight(
			bounds.Min.X, bounds.Min.Y,
			bounds.Dx(), bounds.Dy(),
			a.config.Scroll.HighlightColor,
			a.config.Scroll.HighlightWidth,
		)
	}

	a.hintOverlay.Show()
}

// exitMode exits the current mode
func (a *App) exitMode() {
	if a.currentMode == ModeIdle {
		return
	}

	a.logger.Info("Exiting current mode", zap.String("mode", a.getModeString()))

	// Clear overlay
	a.hintOverlay.Hide()
	a.hintOverlay.Clear()

	// Clean up mode-specific state
	switch a.currentMode {
	case ModeHint, ModeHintWithActions:
		a.currentHints = nil
		a.hintInput = ""
		a.selectedHint = nil
	case ModeScroll:
		a.scrollController.Cleanup()
	}

	// Disable event tap when exiting any mode
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	// if a.config.Accessibility.ElectronSupport.Enable {
	// 	accessibility.ResetElectronSupport()
	// }

	a.currentMode = ModeIdle
}

// getModeString returns the current mode as a string
func (a *App) getModeString() string {
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

// isLetter checks if a byte is a letter
func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// Cleanup cleans up resources
func (a *App) Cleanup() {
	a.logger.Info("Cleaning up")

	a.exitMode()
	a.hotkeyManager.UnregisterAll()
	a.hintOverlay.Destroy()

	// Stop IPC server
	if a.ipcServer != nil {
		if err := a.ipcServer.Stop(); err != nil {
			a.logger.Error("Failed to stop IPC server", zap.Error(err))
		}
	}

	// Cleanup event tap
	if a.eventTap != nil {
		a.eventTap.Destroy()
	}

	// Close logger (syncs and closes log file)
	if err := logger.Close(); err != nil {
		// Can't log this since logger is being closed
		fmt.Fprintf(os.Stderr, "Warning: failed to close logger: %v\n", err)
	}
}

// handleIPCCommand handles IPC commands from the CLI
func (a *App) handleIPCCommand(cmd ipc.Command) ipc.Response {
	a.logger.Info("Handling IPC command", zap.String("action", cmd.Action))

	switch cmd.Action {
	case "ping":
		return ipc.Response{Success: true, Message: "pong"}

	case "start":
		if a.enabled {
			return ipc.Response{Success: false, Message: "govim is already running"}
		}
		a.enabled = true
		return ipc.Response{Success: true, Message: "govim started"}

	case "stop":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "govim is already stopped"}
		}
		a.enabled = false
		a.exitMode()
		return ipc.Response{Success: true, Message: "govim stopped"}

	case "hints":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "govim is not running"}
		}
		a.activateHintMode(false)
		return ipc.Response{Success: true, Message: "hint mode activated"}

	case "hints_action":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "govim is not running"}
		}
		a.activateHintMode(true)
		return ipc.Response{Success: true, Message: "hint mode with actions activated"}

	case "scroll":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "govim is not running"}
		}
		a.activateScrollMode()
		return ipc.Response{Success: true, Message: "scroll mode activated"}

	case "idle":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "govim is not running"}
		}
		a.exitMode()
		return ipc.Response{Success: true, Message: "mode set to idle"}

	case "status":
		cfgPath := a.ConfigPath
		if cfgPath == "" {
			// Fallback to the standard config path if daemon wasn't started
			// with an explicit --config
			cfgPath = config.GetConfigPath()
		}

		// Expand ~ to home dir and resolve relative paths to absolute
		if strings.HasPrefix(cfgPath, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				cfgPath = filepath.Join(home, cfgPath[1:])
			}
		}
		if abs, err := filepath.Abs(cfgPath); err == nil {
			cfgPath = abs
		}

		statusData := map[string]any{
			"enabled": a.enabled,
			"mode":    a.getModeString(),
			"config":  cfgPath,
		}
		return ipc.Response{Success: true, Data: statusData}

	default:
		return ipc.Response{Success: false, Message: fmt.Sprintf("unknown command: %s", cmd.Action)}
	}
}

var globalApp *App

func onReady() {
	systray.SetTitle("⌨️")
	systray.SetTooltip("GoVim - Keyboard Navigation")

	mQuit := systray.AddMenuItem("Quit GoVim", "Exit the application")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	if globalApp != nil {
		globalApp.Cleanup()
	}
}

func main() {
	// Set the launch function for CLI
	cli.LaunchFunc = LaunchDaemon

	// Let cobra handle all command parsing
	cli.Execute()
}

// LaunchDaemon is called by the CLI to launch the daemon
func LaunchDaemon(configPath string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create app
	app, err := NewApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}
	// Record the config path the daemon was started with so status can report
	// the same path (if provided via --config)
	app.ConfigPath = configPath
	globalApp = app

	// Start app in goroutine
	go func() {
		if err := app.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		}
	}()

	// Run systray (blocks until quit)
	systray.Run(onReady, onExit)
}
