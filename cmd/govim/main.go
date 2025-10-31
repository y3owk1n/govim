package main

import (
	"fmt"
	"os"
	"os/signal"
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

	// Create event tap for hint mode
	app.eventTap = eventtap.NewEventTap(app.handleHintKey, log)
	if app.eventTap == nil {
		log.Warn("Event tap creation failed - hint mode key capture won't work")
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
	fmt.Printf("  Hint mode: %s\n", a.config.Hotkeys.ActivateHintMode)
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
	// Hint mode hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ActivateHintMode, a.activateHintMode); err != nil {
		return fmt.Errorf("failed to register hint mode hotkey: %w", err)
	}

	// Scroll mode hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ActivateScrollMode, a.activateScrollMode); err != nil {
		return fmt.Errorf("failed to register scroll mode hotkey: %w", err)
	}

	// Exit mode hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ExitMode, a.exitMode); err != nil {
		return fmt.Errorf("failed to register exit mode hotkey: %w", err)
	}

	// Reload config hotkey
	if _, err := a.hotkeyManager.Register(a.config.Hotkeys.ReloadConfig, a.reloadConfig); err != nil {
		a.logger.Warn("Failed to register reload config hotkey", zap.Error(err))
	}

	return nil
}

// activateHintMode activates hint mode
func (a *App) activateHintMode() {
	if !a.enabled {
		a.logger.Debug("GoVim is disabled, ignoring hint mode activation")
		return
	}
	if a.currentMode == ModeHint {
		a.logger.Debug("Hint mode already active")
		return
	}

	a.logger.Info("Activating hint mode")
	a.exitMode() // Exit current mode first

	// Get clickable elements
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
	a.currentMode = ModeHint

	a.logger.Info("Hint mode activated", zap.Int("hints", len(hintList)))
	a.logger.Info("Type a hint label to click the element")
	
	// Enable event tap to capture keys
	a.hintInput = ""
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
}

// handleHintKey handles key presses in hint mode and scroll mode
func (a *App) handleHintKey(key string) {
	// Handle escape in any mode
	if key == "\x1b" || key == "escape" {
		a.logger.Debug("Escape pressed")
		a.exitMode()
		return
	}

	// Handle scroll mode
	if a.currentMode == ModeScroll {
		a.handleScrollKey(key)
		return
	}

	// Handle hint mode
	if a.currentMode == ModeHint && a.currentHints != nil {
		a.logger.Debug("Hint key", zap.String("key", key))
		
		// Only accept lowercase letters
		if len(key) != 1 || key[0] < 'a' || key[0] > 'z' {
			a.logger.Debug("Ignoring non-letter key", zap.String("key", key))
			return
		}
		
		// Add to input buffer
		a.hintInput += key
		a.logger.Debug("Hint input", zap.String("input", a.hintInput))
		
		// Check if any hints start with this input
		filtered := a.currentHints.FilterByPrefix(a.hintInput)
		
		if len(filtered) == 0 {
			// No matches - reset
			a.logger.Debug("No matching hints, resetting")
			a.hintInput = ""
			return
		}
		
		// If exactly one match and input matches the full label, click it
		if len(filtered) == 1 && filtered[0].Label == a.hintInput {
			hint := filtered[0]
			a.logger.Info("Clicking element", zap.String("label", a.hintInput))
			if err := hint.Element.Element.Click(); err != nil {
				a.logger.Error("Failed to click element", zap.Error(err))
			}
			a.exitMode()
			return
		}
		
		// Otherwise, keep collecting input
		a.logger.Debug("Partial match", zap.Int("matches", len(filtered)))
	}
}

// handleScrollKey handles key presses in scroll mode
func (a *App) handleScrollKey(key string) {
	a.logger.Info("Scroll key pressed", zap.String("key", key))
	
	var err error
	switch key {
	case "j":
		a.logger.Info("Scrolling down")
		err = a.scrollController.ScrollDown()
	case "k":
		a.logger.Info("Scrolling up")
		err = a.scrollController.ScrollUp()
	case "h":
		a.logger.Info("Scrolling left")
		err = a.scrollController.ScrollLeft()
	case "l":
		a.logger.Info("Scrolling right")
		err = a.scrollController.ScrollRight()
	default:
		a.logger.Debug("Ignoring non-scroll key", zap.String("key", key))
		return
	}
	
	if err != nil {
		a.logger.Error("Scroll failed", zap.Error(err))
	}
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

	// Initialize scroll controller
	if err := a.scrollController.Initialize(); err != nil {
		a.logger.Error("Failed to initialize scroll controller", zap.Error(err))
		return
	}

	// Get active scroll area
	area := a.scrollController.GetActiveArea()
	if area == nil {
		a.logger.Warn("No scrollable areas found")
		return
	}

	// Draw scroll highlight
	bounds := area.Bounds
	a.hintOverlay.DrawScrollHighlight(
		bounds.Min.X, bounds.Min.Y,
		bounds.Dx(), bounds.Dy(),
		a.config.Scroll.HighlightColor,
		a.config.Scroll.HighlightWidth,
	)
	a.hintOverlay.Show()

	a.currentMode = ModeScroll
	a.logger.Info("Scroll mode activated")
	a.logger.Info("Use J/K to scroll down/up, H/L for left/right, Escape to exit")

	// Enable event tap to capture scroll keys
	if a.eventTap != nil {
		a.eventTap.Enable()
	}
}

// registerScrollKeys registers scroll mode keys using event tap
func (a *App) registerScrollKeys() {
	// Don't use global hotkeys for scroll keys - they conflict with typing
	// Instead, we'll use the event tap to capture them
	a.logger.Debug("Scroll keys will be captured via event tap")
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
	if a.currentMode == ModeHint {
		a.currentHints = nil
		a.hintInput = ""
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
	a.logger.Info("Configuration reloaded successfully")
}

// getModeString returns the current mode as a string
func (a *App) getModeString() string {
	switch a.currentMode {
	case ModeIdle:
		return "idle"
	case ModeHint:
		return "hint"
	case ModeScroll:
		return "scroll"
	default:
		return "unknown"
	}
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
