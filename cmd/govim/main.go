package main

import (
	"fmt"
	"image"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/govim/internal/accessibility"
	_ "github.com/y3owk1n/govim/internal/bridge" // Import for CGo compilation
	"github.com/y3owk1n/govim/internal/cli"
	"github.com/y3owk1n/govim/internal/config"
	"github.com/y3owk1n/govim/internal/eventtap"
	"github.com/y3owk1n/govim/internal/hints"
	"github.com/y3owk1n/govim/internal/hotkeys"
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
	config           *config.Config
	logger           *zap.Logger
	hotkeyManager    *hotkeys.Manager
	hintGenerator    *hints.Generator
	hintOverlay      *hints.Overlay
	scrollController *scroll.Controller
	eventTap         *eventtap.EventTap
	currentMode      Mode
	currentHints     *hints.HintCollection
	hintInput        string
	lastScrollKey    string
	selectedHint     *hints.Hint
	enabled          bool
}

func hotkeyHasModifier(key string) bool {
	parts := strings.Split(key, "+")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		switch strings.ToLower(trimmed) {
		case "cmd", "command", "shift", "alt", "option", "ctrl", "control":
			return true
		}
	}
	return false
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

	// Apply additional clickable roles from config
	if len(cfg.Accessibility.AdditionalClickableRoles) > 0 {
		log.Info("Applying additional clickable roles",
			zap.Int("count", len(cfg.Accessibility.AdditionalClickableRoles)),
			zap.Strings("roles", cfg.Accessibility.AdditionalClickableRoles))
	} else {
		log.Debug("No additional clickable roles configured")
	}
	accessibility.SetAdditionalClickableRoles(cfg.Accessibility.AdditionalClickableRoles)

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

	return app, nil
}

// Run runs the application
func (a *App) Run() error {
	a.logger.Info("Starting GoVim")

	// Register hotkeys
	if err := a.registerHotkeys(); err != nil {
		return fmt.Errorf("failed to register hotkeys: %w", err)
	}

	a.logger.Info("GoVim is running. Press hotkeys to activate modes.")
	fmt.Println("✓ GoVim is running")
	fmt.Printf("  Hint mode (direct): %s\n", a.config.Hotkeys.ActivateHintMode)
	fmt.Printf("  Hint mode (with actions): %s\n", a.config.Hotkeys.ActivateHintModeWithActions)
	fmt.Printf("  Scroll mode: %s\n", a.config.Hotkeys.ActivateScrollMode)
	fmt.Printf("  Exit mode: %s\n", a.config.Hotkeys.ExitMode)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	a.logger.Info("Shutting down GoVim")
	return nil
}

// registerHotkeys registers all global hotkeys
func (a *App) registerHotkeys() error {
	// Hint mode hotkey (direct click)
	a.logger.Info("Registering hint mode hotkey", zap.String("key", a.config.Hotkeys.ActivateHintMode))
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ActivateHintMode, func() {
		a.activateHintMode(false)
	}); err != nil {
		return fmt.Errorf("failed to register hint mode hotkey: %w", err)
	}

	// Hint mode with actions hotkey
	a.logger.Info("Registering hint mode with actions hotkey", zap.String("key", a.config.Hotkeys.ActivateHintModeWithActions))
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ActivateHintModeWithActions, func() {
		a.activateHintMode(true)
	}); err != nil {
		return fmt.Errorf("failed to register hint mode with actions hotkey: %w", err)
	}

	// Scroll mode hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ActivateScrollMode, a.activateScrollMode); err != nil {
		return fmt.Errorf("failed to register scroll mode hotkey: %w", err)
	}

	// Exit mode hotkey
	exitHotkey := a.config.Hotkeys.ExitMode
	if hotkeyHasModifier(exitHotkey) {
		if _, err := a.hotkeyManager.Register(exitHotkey, a.exitMode); err != nil {
			return fmt.Errorf("failed to register exit mode hotkey: %w", err)
		}
	} else {
		a.logger.Info("Skipping global exit hotkey registration to avoid intercepting unmodified key",
			zap.String("key", exitHotkey))
	}

	// Reload config hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ReloadConfig, a.reloadConfig); err != nil {
		a.logger.Warn("Failed to register reload config hotkey", zap.Error(err))
	}

	return nil
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

	// Get clickable elements
	defaultRoles, additionalRoles := accessibility.GetClickableRoles()
	a.logger.Debug("Scanning for clickable elements",
		zap.Strings("default_roles", defaultRoles),
		zap.Strings("additional_roles", additionalRoles))

	elements, err := accessibility.GetClickableElements()
	if err != nil {
		a.logger.Error("Failed to get clickable elements", zap.Error(err))
		return
	}

	a.logger.Info("Found clickable elements", zap.Int("count", len(elements)))

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
		a.logger.Info("Type a hint label, then choose action: f=left, d=right, s=double, a=middle")
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
	if a.currentMode == ModeHint || a.currentMode == ModeHintWithActions {
		a.handleHintKey(key)
	} else if a.currentMode == ModeScroll {
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
	cfg := a.config.Hints

	// Create action hints showing available actions
	actionText := fmt.Sprintf("[%s]Left [%s]Right [%s]Double [%s]Middle",
		cfg.ClickActionLeft,
		cfg.ClickActionRight,
		cfg.ClickActionDouble,
		cfg.ClickActionMiddle)

	// Create a single hint at the element's position with the action menu
	actionHint := &hints.Hint{
		Label:    actionText,
		Element:  hint.Element,
		Position: hint.Position,
	}

	// Draw the action menu
	if err := a.hintOverlay.DrawHints([]*hints.Hint{actionHint}); err != nil {
		a.logger.Error("Failed to draw action menu", zap.Error(err))
	}
}

// handleActionKey handles action key presses after a hint is selected
func (a *App) handleActionKey(key string) {
	if a.selectedHint == nil {
		return
	}

	hint := a.selectedHint
	cfg := a.config.Hints

	a.logger.Info("Action key pressed", zap.String("key", key))

	var err error
	switch key {
	case cfg.ClickActionLeft:
		a.logger.Info("Performing left click", zap.String("label", hint.Label))
		err = hint.Element.Element.Click()
	case cfg.ClickActionRight:
		a.logger.Info("Performing right click", zap.String("label", hint.Label))
		err = hint.Element.Element.RightClick()
	case cfg.ClickActionDouble:
		a.logger.Info("Performing double click", zap.String("label", hint.Label))
		err = hint.Element.Element.DoubleClick()
	case cfg.ClickActionMiddle:
		a.logger.Info("Performing middle click", zap.String("label", hint.Label))
		// Middle click not implemented yet, fallback to left click
		a.logger.Warn("Middle click not implemented, using left click")
		err = hint.Element.Element.Click()
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
	number := int(key[0] - '0')
	detector := a.scrollController.GetDetector()
	newArea := detector.SetActiveByNumber(number)

	if newArea != nil {
		a.logger.Info("Switched to scroll area", zap.Int("number", number))
		a.updateScrollHighlight(newArea)
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

	a.logger.Info("Initializing scroll controller")
	// Initialize scroll controller
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
		a.hintOverlay.DrawHints(areaHints)
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
	if a.currentMode == ModeHint || a.currentMode == ModeHintWithActions {
		a.currentHints = nil
		a.hintInput = ""
		a.selectedHint = nil
	} else if a.currentMode == ModeScroll {
		a.scrollController.Cleanup()
	}

	// Disable event tap when exiting any mode
	if a.eventTap != nil {
		a.eventTap.Disable()
	}

	a.currentMode = ModeIdle
}

// reloadConfig reloads the configuration
func (a *App) reloadConfig() {
	a.logger.Info("Reloading configuration")

	newConfig, err := config.Load("")
	if err != nil {
		a.logger.Error("Failed to reload configuration", zap.Error(err))
		return
	}

	a.config = newConfig
	if len(newConfig.Accessibility.AdditionalClickableRoles) > 0 {
		a.logger.Info("Applying additional clickable roles",
			zap.Int("count", len(newConfig.Accessibility.AdditionalClickableRoles)),
			zap.Strings("roles", newConfig.Accessibility.AdditionalClickableRoles))
	} else {
		a.logger.Debug("No additional clickable roles configured")
	}
	accessibility.SetAdditionalClickableRoles(newConfig.Accessibility.AdditionalClickableRoles)
	a.logger.Info("Configuration reloaded successfully")
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

	logger.Sync()
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
	// If CLI arguments provided, run CLI mode
	if len(os.Args) > 1 {
		cli.Run(os.Args[1:])
		return
	}

	// Load configuration
	cfg, err := config.Load("")
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
