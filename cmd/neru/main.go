package main

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/appwatcher"
	_ "github.com/y3owk1n/neru/internal/bridge" // Import for CGo compilation
	"github.com/y3owk1n/neru/internal/cli"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/electron"
	"github.com/y3owk1n/neru/internal/eventtap"
	"github.com/y3owk1n/neru/internal/hints"
	"github.com/y3owk1n/neru/internal/hotkeys"
	"github.com/y3owk1n/neru/internal/ipc"
	"github.com/y3owk1n/neru/internal/logger"
	"github.com/y3owk1n/neru/internal/scroll"
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
	ConfigPath        string
	logger            *zap.Logger
	hotkeyManager     *hotkeys.Manager
	hintGenerator     *hints.Generator
	hintOverlay       *hints.Overlay
	scrollController  *scroll.Controller
	eventTap          *eventtap.EventTap
	ipcServer         *ipc.Server
	electronManager   *electron.ElectronManager
	appWatcher        *appwatcher.Watcher
	currentMode       Mode
	hintManager       *hints.Manager
	scrollHintManager *hints.Manager
	lastScrollKey     string
	selectedHint      *hints.Hint
	enabled           bool
	// Track whether global hotkeys are currently registered
	hotkeysRegistered bool
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	if err := logger.Init(cfg.Logging.LogLevel, cfg.Logging.LogFile, cfg.Logging.StructuredLogging); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log := logger.Get()

	// Check accessibility permissions
	if cfg.Accessibility.AccessibilityCheckOnStart {
		if !accessibility.CheckAccessibilityPermissions() {
			log.Warn("Accessibility permissions not granted. Please grant permissions in System Settings.")
			fmt.Println("⚠️  Neru requires Accessibility permissions to function.")
			fmt.Println("Please go to: System Settings → Privacy & Security → Accessibility")
			fmt.Println("and enable Neru.")
			return nil, fmt.Errorf("accessibility permissions required")
		}
	}

	// Set global config
	config.SetGlobal(cfg)

	// Apply clickable and scrollable roles from config
	log.Info("Applying clickable roles",
		zap.Int("count", len(cfg.Accessibility.ClickableRoles)),
		zap.Strings("roles", cfg.Accessibility.ClickableRoles))
	accessibility.SetClickableRoles(cfg.Accessibility.ClickableRoles)

	log.Info("Applying scrollable roles",
		zap.Int("count", len(cfg.Accessibility.ScrollableRoles)),
		zap.Strings("roles", cfg.Accessibility.ScrollableRoles))
	accessibility.SetScrollableRoles(cfg.Accessibility.ScrollableRoles)

	// Create app watcher
	appWatcher := appwatcher.NewWatcher()

	// Create hotkey manager
	hotkeyMgr := hotkeys.NewManager(log)
	hotkeys.SetGlobalManager(hotkeyMgr)

	// Create hint generator
	hintGen := hints.NewGenerator(
		cfg.Hints.HintCharacters,
	)

	// Create hint overlay
	hintOverlay, err := hints.NewOverlay(cfg.Hints)
	if err != nil {
		return nil, fmt.Errorf("failed to create overlay: %w", err)
	}

	// Create scroll controller
	scrollCtrl := scroll.NewController(cfg.Scroll, log)

	app := &App{
		config:            cfg,
		logger:            log,
		hotkeyManager:     hotkeyMgr,
		appWatcher:        appWatcher,
		hintGenerator:     hintGen,
		hintOverlay:       hintOverlay,
		scrollController:  scrollCtrl,
		currentMode:       ModeIdle,
		enabled:           true,
		hotkeysRegistered: false,
	}

	// Initialize hint managers
	app.hintManager = hints.NewManager(func(hints []*hints.Hint) {
		if app.currentMode == ModeHintWithActions {
			style := app.config.Hints
			style.BackgroundColor = app.config.Hints.ActionBackgroundColor
			style.TextColor = app.config.Hints.ActionTextColor
			style.MatchedTextColor = app.config.Hints.ActionMatchedTextColor
			style.BorderColor = app.config.Hints.ActionBorderColor
			style.Opacity = app.config.Hints.ActionOpacity
			if err := app.hintOverlay.DrawHintsWithStyle(hints, style); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		} else {
			if err := app.hintOverlay.DrawHints(hints); err != nil {
				app.logger.Error("Failed to redraw hints", zap.Error(err))
			}
		}
	})

	app.scrollHintManager = hints.NewManager(func(hints []*hints.Hint) {
		app.drawFilteredScrollHints(hints)
	})

	// Create electron manager
	if cfg.Accessibility.AdditionalAXSupport.Enable {
		app.electronManager = electron.NewElectronManager(cfg.Accessibility.AdditionalAXSupport.AdditionalElectronBundles, cfg.Accessibility.AdditionalAXSupport.AdditionalChromiumBundles, cfg.Accessibility.AdditionalAXSupport.AdditionalFirefoxBundles)
	}

	// Create event tap for capturing keys in modes
	app.eventTap = eventtap.NewEventTap(app.handleKeyPress, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - key capture won't work")
	} else {

		keys := make([]string, 0, len(cfg.Hotkeys.Bindings))
		for k := range cfg.Hotkeys.Bindings {
			keys = append(keys, k)
		}

		// Configure hotkeys that should pass through to the global hotkey system
		app.eventTap.SetHotkeys(keys)

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
	a.logger.Info("Starting Neru")

	// Start IPC server
	a.ipcServer.Start()
	a.logger.Info("IPC server started")

	// Start app watcher
	a.appWatcher.Start()
	a.logger.Info("App watcher started")

	// Initialize hotkeys based on current focused app and exclusion
	a.refreshHotkeysForAppOrCurrent("")
	a.logger.Info("Hotkeys initialized")

	// setup watcher callbacks
	a.appWatcher.OnActivate(func(appName, bundleID string) {
		a.logger.Debug("App activated", zap.String("bundle_id", bundleID))

		// refresh hotkeys for app
		go a.refreshHotkeysForAppOrCurrent(bundleID)
		a.logger.Debug("Handled hotkey refresh")

		if a.config.Accessibility.AdditionalAXSupport.Enable {
			// handle electrons
			if electron.ShouldEnableElectronSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalElectronBundles) {
				electron.EnsureElectronAccessibility(bundleID)
				a.logger.Debug("Handled electron accessibility")
			}

			// handle chromium
			if electron.ShouldEnableChromiumSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalChromiumBundles) {
				electron.EnsureChromiumAccessibility(bundleID)
				a.logger.Debug("Handled chromium accessibility")
			}

			// handle firefox
			if electron.ShouldEnableFirefoxSupport(bundleID, a.config.Accessibility.AdditionalAXSupport.AdditionalFirefoxBundles) {
				electron.EnsureFirefoxAccessibility(bundleID)
				a.logger.Debug("Handled firefox accessibility")
			}

		}

		a.logger.Debug("Done handling app activation")
	})

	a.logger.Info("Neru is running")
	fmt.Println("✓ Neru is running")

	for k, v := range a.config.Hotkeys.Bindings {
		toShow := v
		if strings.HasPrefix(v, "exec") {
			runes := []rune(v)
			if len(runes) > 30 {
				toShow = string(runes[:30]) + "..."
			}
		}
		fmt.Printf("  %s: %s\n", k, toShow)
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
	// Note: Escape key for exiting modes is hardcoded in handleKeyPress, not registered as global hotkey

	// Register arbitrary bindings from config.Hotkeys.Bindings
	// We intentionally don't fail the entire registration process if one binding fails;
	// instead we log the error and continue so the daemon remains running.
	for k, v := range a.config.Hotkeys.Bindings {
		key := strings.TrimSpace(k)
		action := strings.TrimSpace(v)
		if key == "" || action == "" {
			continue
		}

		a.logger.Info("Registering hotkey binding", zap.String("key", key), zap.String("action", action))

		// Capture values for closure
		bindKey := key
		bindAction := action

		if _, err := a.hotkeyManager.Register(bindKey, func() {
			// Run handler in separate goroutine so the hotkey callback returns quickly.
			go func() {
				defer func() {
					if r := recover(); r != nil {
						a.logger.Error("panic in hotkey handler", zap.Any("recover", r), zap.String("key", bindKey))
					}
				}()

				// Exec mode: run arbitrary bash command
				if strings.HasPrefix(bindAction, "exec ") {
					cmdStr := strings.TrimSpace(strings.TrimPrefix(bindAction, "exec"))
					if cmdStr == "" {
						a.logger.Error("hotkey exec has empty command", zap.String("key", bindKey))
						return
					}

					a.logger.Debug("Executing shell command from hotkey", zap.String("key", bindKey), zap.String("cmd", cmdStr))
					cmd := exec.Command("/bin/bash", "-lc", cmdStr)
					out, err := cmd.CombinedOutput()
					if err != nil {
						a.logger.Error("hotkey exec failed", zap.String("key", bindKey), zap.String("cmd", cmdStr), zap.ByteString("output", out), zap.Error(err))
					} else {
						a.logger.Info("hotkey exec completed", zap.String("key", bindKey), zap.String("cmd", cmdStr), zap.ByteString("output", out))
					}
					return
				}

				// Otherwise treat the action as an internal neru command and dispatch it
				resp := a.handleIPCCommand(ipc.Command{Action: bindAction})
				if !resp.Success {
					a.logger.Error("hotkey action failed", zap.String("key", bindKey), zap.String("action", bindAction), zap.String("message", resp.Message))
				} else {
					a.logger.Info("hotkey action executed", zap.String("key", bindKey), zap.String("action", bindAction))
				}
			}()
		}); err != nil {
			a.logger.Error("Failed to register hotkey binding", zap.String("key", key), zap.String("action", action), zap.Error(err))
			// continue registering other bindings
			continue
		}
	}

	return nil
}

// refreshHotkeysForAppOrCurrent registers or unregisters global hotkeys based on
// whether Neru is enabled and whether the currently focused app is excluded.
func (a *App) refreshHotkeysForAppOrCurrent(bundleID string) {
	// If disabled, ensure no hotkeys are registered
	if !a.enabled {
		if a.hotkeysRegistered {
			a.logger.Debug("Neru disabled; unregistering hotkeys")
			a.hotkeyManager.UnregisterAll()
			a.hotkeysRegistered = false
		}
		return
	}

	if bundleID == "" {
		bundleID = a.getFocusedBundleID()
	}

	// If app is excluded, unregister; otherwise ensure registered
	if a.config.IsAppExcluded(bundleID) {
		if a.hotkeysRegistered {
			a.logger.Info("Focused app excluded; unregistering global hotkeys",
				zap.String("bundle_id", bundleID))
			a.hotkeyManager.UnregisterAll()
			a.hotkeysRegistered = false
		}
		return
	}

	if !a.hotkeysRegistered {
		if err := a.registerHotkeys(); err != nil {
			a.logger.Error("Failed to register hotkeys", zap.Error(err))
			return
		}
		a.hotkeysRegistered = true
		a.logger.Debug("Hotkeys registered")
	}
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

// getFocusedBundleID returns the bundle identifier of the currently focused
// application, or an empty string if it cannot be determined.
func (a *App) getFocusedBundleID() string {
	app := accessibility.GetFocusedApplication()
	if app == nil {
		return ""
	}
	defer app.Release()
	return app.GetBundleIdentifier()
}

// isFocusedAppExcluded returns true if the currently focused application's bundle
// ID is in the excluded apps list. Logs context for debugging.
func (a *App) isFocusedAppExcluded() bool {
	bundleID := a.getFocusedBundleID()
	if bundleID != "" && a.config.IsAppExcluded(bundleID) {
		a.logger.Debug("Current app is excluded; ignoring mode activation",
			zap.String("bundle_id", bundleID))
		return true
	}
	return false
}

// activateHintMode activates hint mode
func (a *App) activateHintMode(withActions bool) {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring hint mode activation")
		return
	}
	if a.currentMode == ModeHint || a.currentMode == ModeHintWithActions {
		a.logger.Debug("Hint mode already active")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.logger.Info("Activating hint mode")
	a.exitMode() // Exit current mode first

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

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
	if a.config.General.IncludeMenubarHints {
		if mbElems, merr := accessibility.GetMenuBarClickableElements(); merr == nil {
			elements = append(elements, mbElems...)
			a.logger.Debug("Included menubar elements", zap.Int("count", len(mbElems)))
		} else {
			a.logger.Warn("Failed to get menubar elements", zap.Error(merr))
		}

		// Optionally include additional menubar elements
		for _, bundleID := range a.config.General.AdditionalMenubarHintsTargets {
			if additionalElems, derr := accessibility.GetClickableElementsFromBundleID(bundleID); derr == nil {
				elements = append(elements, additionalElems...)
				a.logger.Debug("Included additional menubar elements", zap.Int("count", len(additionalElems)))
			} else {
				a.logger.Warn("Failed to get additional menubar elements", zap.Error(derr))
			}
		}

	}

	// Optionally include Dock elements
	if a.config.General.IncludeDockHints {
		if dockElems, derr := accessibility.GetClickableElementsFromBundleID("com.apple.dock"); derr == nil {
			elements = append(elements, dockElems...)
			a.logger.Debug("Included dock elements", zap.Int("count", len(dockElems)))
		} else {
			a.logger.Warn("Failed to get dock elements", zap.Error(derr))
		}
	}

	// Optionally include Notification Center elements
	if a.config.General.IncludeNCHints {
		if ncElems, derr := accessibility.GetClickableElementsFromBundleID("com.apple.notificationcenterui"); derr == nil {
			elements = append(elements, ncElems...)
			a.logger.Debug("Included nc elements", zap.Int("count", len(ncElems)))
		} else {
			a.logger.Warn("Failed to get nc elements", zap.Error(derr))
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

	// Set up hints in the hint manager
	hints := hints.NewHintCollection(hintList)
	a.hintManager.SetHints(hints)

	// Draw hints
	// Use distinct colors when entering hint mode with actions
	if withActions {
		style := a.config.Hints
		style.BackgroundColor = a.config.Hints.ActionBackgroundColor
		style.TextColor = a.config.Hints.ActionTextColor
		style.MatchedTextColor = a.config.Hints.ActionMatchedTextColor
		style.BorderColor = a.config.Hints.ActionBorderColor
		style.Opacity = a.config.Hints.ActionOpacity
		if err := a.hintOverlay.DrawHintsWithStyle(hintList, style); err != nil {
			a.logger.Error("Failed to draw hints", zap.Error(err))
			return
		}
	} else {
		if err := a.hintOverlay.DrawHints(hintList); err != nil {
			a.logger.Error("Failed to draw hints", zap.Error(err))
			return
		}
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

	// Reset state
	a.selectedHint = nil

	// Enable event tap to capture keys (must be last to ensure proper initialization)
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

	if hint, ok := a.hintManager.HandleInput(key); ok {
		if a.currentMode == ModeHintWithActions {
			// Store the hint and wait for action selection
			a.selectedHint = hint
			a.logger.Info("Hint selected, choose action", zap.String("label", a.hintManager.GetInput()))

			// Clear all hints and show action menu at the hint location
			a.hintOverlay.Clear()
			a.showActionMenu(hint)
		} else {
			// Direct click mode - click immediately
			a.logger.Info("Clicking element", zap.String("label", a.hintManager.GetInput()))
			if err := hint.Element.Element.LeftClick(a.config.General.RestorePosAfterLeftClick); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
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

	// Create smaller style for action hints
	actionStyle := a.config.Hints
	// Apply distinct colors for action overlay from config
	actionStyle.BackgroundColor = a.config.Hints.ActionBackgroundColor
	actionStyle.TextColor = a.config.Hints.ActionTextColor
	actionStyle.MatchedTextColor = a.config.Hints.ActionMatchedTextColor
	actionStyle.BorderColor = a.config.Hints.ActionBorderColor
	actionStyle.Opacity = a.config.Hints.ActionOpacity
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

	// Handle backspace/delete to remove last character
	if key == "\x7f" || key == "delete" || key == "backspace" {
		if a.scrollHintManager != nil {
			a.resetScrollHints()
		}
		return
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
	default:
		// Handle hint characters for scroll areas
		if len(key) == 1 && isLetter(key[0]) {
			if hint, ok := a.scrollHintManager.HandleInput(key); ok {
				a.logger.Info("Matched scroll area", zap.String("label", hint.Label))
				a.switchScrollAreaByHint(hint)
				// Reset input but keep hints visible
				a.scrollHintManager.Reset()
				return
			}
			return
		}

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

// switchScrollAreaByHint switches to a specific scroll area by hint
func (a *App) switchScrollAreaByHint(hint *hints.Hint) {
	if hint == nil || hint.Element == nil {
		a.logger.Debug("Invalid hint")
		return
	}

	// Find the matching scroll area
	detector := a.scrollController.GetDetector()
	areas := detector.GetAreas()

	for i, area := range areas {
		if area.Element == hint.Element {
			if err := detector.SetActiveArea(i); err != nil {
				a.logger.Error("Failed to set active area", zap.Error(err))
				return
			}
			a.logger.Info("Switched to scroll area", zap.String("label", hint.Label))
			a.updateScrollHighlight(area)
			return
		}
	}

	a.logger.Debug("No matching scroll area found")
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

// resetScrollHints resets all hints to their original state and redraws them
func (a *App) resetScrollHints() {
	if a.scrollHintManager != nil {
		a.scrollHintManager.Reset()
	}
}

// drawFilteredScrollHints redraws the hints with highlighting for matched characters
func (a *App) drawFilteredScrollHints(filtered []*hints.Hint) {
	a.hintOverlay.Clear()

	// Draw the filtered hints
	if err := a.hintOverlay.DrawHints(filtered); err != nil {
		a.logger.Error("Failed to redraw hints", zap.Error(err))
	}

	// Draw highlight on active area
	detector := a.scrollController.GetDetector()
	if detector != nil {
		area := detector.GetActiveArea()
		if area != nil {
			a.updateScrollHighlight(area)
		}
	}

	a.hintOverlay.Show()
}

// activateScrollMode activates scroll mode
func (a *App) activateScrollMode() {
	if !a.enabled {
		a.logger.Debug("Neru is disabled, ignoring scroll mode activation")
		return
	}

	if a.currentMode == ModeScroll {
		a.logger.Debug("Scroll mode already active")
		return
	}

	// Centralized exclusion guard
	if a.isFocusedAppExcluded() {
		return
	}

	a.logger.Info("Activating scroll mode")
	a.exitMode() // Exit current mode first

	// Update roles for the current focused app
	a.updateRolesForCurrentApp()

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

	// Create new scroll hint manager
	a.scrollHintManager = hints.NewManager(func(hints []*hints.Hint) {
		if err := a.hintOverlay.DrawHints(hints); err != nil {
			a.logger.Error("Failed to draw filtered hints", zap.Error(err))
		}
	})

	a.logger.Info("Drawing scroll area labels", zap.Int("count", len(areas)))
	// Draw scroll highlights with numbers for all areas
	a.drawScrollAreaLabels(areas)

	a.logger.Info("Setting mode to scroll")
	a.currentMode = ModeScroll
	a.logger.Info("Scroll mode activated", zap.Int("areas", len(areas)))
	a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom")
	a.logger.Info("Tab to switch areas, or type a hint label to jump to area")

	a.logger.Info("Enabling event tap")
	// Enable event tap to capture scroll keys
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
	// Reset any previous hint state
	a.resetScrollHints()
	a.logger.Info("Scroll mode activation complete")
}

// drawScrollAreaLabels draws hint labels on all scroll areas
func (a *App) drawScrollAreaLabels(areas []*scroll.ScrollArea) {
	// Clear existing overlay
	a.hintOverlay.Clear()

	// Generate simple sequential hints
	labels := a.hintGenerator.GenerateScrollLabels(len(areas))
	if len(labels) == 0 {
		a.logger.Error("Not enough available characters for scroll hints after filtering scroll keys")
		return
	}

	if len(labels) < len(areas) {
		a.logger.Warn("Some scroll areas will not be labeled due to insufficient available characters",
			zap.Int("areas", len(areas)),
			zap.Int("available_labels", len(labels)))
	}

	// Create hints for areas up to the number of available labels
	numHints := len(labels)
	scrollHints := make([]*hints.Hint, numHints)
	for i := 0; i < numHints; i++ {
		area := areas[i]
		scrollHints[i] = &hints.Hint{
			Label:    labels[i],
			Element:  area.Element,
			Position: image.Point{X: area.Bounds.Min.X + 10, Y: area.Bounds.Min.Y + 10},
			Size:     image.Point{X: 30, Y: 30},
		}
	}

	// Set up hints in the scroll hint manager
	hintCollection := hints.NewHintCollection(scrollHints)
	a.scrollHintManager.SetHints(hintCollection)

	// Position hints at top-left of each scroll area
	for i, hint := range scrollHints {
		area := areas[i]
		hint.Position = image.Point{X: area.Bounds.Min.X + 10, Y: area.Bounds.Min.Y + 10}
		hint.Size = image.Point{X: 30, Y: 30}
	}

	// Draw initial hints and highlight
	if err := a.hintOverlay.DrawHints(scrollHints); err != nil {
		a.logger.Error("Failed to draw hints", zap.Error(err))
	}

	detector := a.scrollController.GetDetector()
	if detector != nil {
		if area := detector.GetActiveArea(); area != nil {
			a.updateScrollHighlight(area)
		}
	}

	a.hintOverlay.Show()

	// Set scroll mode and enable key capture
	a.currentMode = ModeScroll
	if a.eventTap != nil {
		a.eventTap.Enable()
	}

	a.logger.Info("Scroll mode activated", zap.Int("areas", len(scrollHints)))
}

// exitMode exits the current mode
func (a *App) exitMode() {
	if a.currentMode == ModeIdle {
		return
	}

	a.logger.Info("Exiting current mode", zap.String("mode", a.getModeString()))

	// First, clean up visual state by resetting hint managers
	switch a.currentMode {
	case ModeHint, ModeHintWithActions:
		if a.hintManager != nil {
			a.hintManager.Reset()
		}
		a.selectedHint = nil
	case ModeScroll:
		// Reset scroll hint manager first
		if a.scrollHintManager != nil {
			a.scrollHintManager.Reset()
			a.scrollHintManager = nil // Ensure it's fully cleared
		}
	}

	// Then clean up controllers that might create visuals
	if a.currentMode == ModeScroll {
		a.scrollController.Cleanup()
	}

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

	// Exit any active mode first
	a.exitMode()

	// Stop IPC server first to prevent new requests
	if a.ipcServer != nil {
		if err := a.ipcServer.Stop(); err != nil {
			a.logger.Error("Failed to stop IPC server", zap.Error(err))
		}
	}

	// Unregister all hotkeys
	if a.hotkeyManager != nil {
		a.hotkeyManager.UnregisterAll()
	}

	// Clean up UI elements
	if a.hintOverlay != nil {
		a.hintOverlay.Destroy()
	}

	// Clean up scroll controller if exists
	if a.scrollController != nil {
		a.scrollController.Cleanup()
	}

	// Cleanup event tap
	if a.eventTap != nil {
		a.eventTap.Destroy()
	}

	// Sync and close logger at the end
	if err := logger.Sync(); err != nil {
		// Ignore "inappropriate ioctl for device" error which occurs when syncing stdout/stderr
		if !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			a.logger.Error("Failed to sync logger", zap.Error(err))
		}
	}

	// Stop app watcher
	a.appWatcher.Stop()

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
			return ipc.Response{Success: false, Message: "neru is already running"}
		}
		a.enabled = true
		return ipc.Response{Success: true, Message: "neru started"}

	case "stop":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is already stopped"}
		}
		a.enabled = false
		a.exitMode()
		return ipc.Response{Success: true, Message: "neru stopped"}

	case "hints":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateHintMode(false)
		return ipc.Response{Success: true, Message: "hint mode activated"}

	case "hints_action":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateHintMode(true)
		return ipc.Response{Success: true, Message: "hint mode with actions activated"}

	case "scroll":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
		}
		a.activateScrollMode()
		return ipc.Response{Success: true, Message: "scroll mode activated"}

	case "idle":
		if !a.enabled {
			return ipc.Response{Success: false, Message: "neru is not running"}
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
	systray.SetTooltip("Neru - Keyboard Navigation")

	// Status submenu for version
	mVersion := systray.AddMenuItem(fmt.Sprintf("Version %s", cli.Version), "Show version")
	mVersion.Disable()

	// Status toggle
	mStatus := systray.AddMenuItem("Status: Running", "Show current status")
	mStatus.Disable()

	// Control actions
	systray.AddSeparator()
	mToggle := systray.AddMenuItem("Disable", "Disable/Enable Neru without quitting")
	mHints := systray.AddMenuItem("Hints Mode", "Show hints")
	mHintsWithActions := systray.AddMenuItem("Hints Mode with Actions", "Show hints with actions")
	mScroll := systray.AddMenuItem("Scroll Mode", "Show scroll hints")

	// Quit option
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Neru", "Exit the application")

	// Handle clicks in a separate goroutine
	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				if globalApp != nil {
					if globalApp.enabled {
						globalApp.enabled = false
						mStatus.SetTitle("Status: Disabled")
						mToggle.SetTitle("Enable")
					} else {
						globalApp.enabled = true
						mStatus.SetTitle("Status: Enabled")
						mToggle.SetTitle("Disable")
					}
				}
			case <-mHints.ClickedCh:
				if globalApp != nil {
					globalApp.activateHintMode(false)
				}
			case <-mHintsWithActions.ClickedCh:
				if globalApp != nil {
					globalApp.activateHintMode(true)
				}
			case <-mScroll.ClickedCh:
				if globalApp != nil {
					globalApp.activateScrollMode()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
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
