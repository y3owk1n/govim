package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/electron"
	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

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
	a.setupAppWatcherCallbacks()

	a.logger.Info("Neru is running")
	a.printStartupInfo()

	// Wait for interrupt signal with force-quit support
	return a.waitForShutdown()
}

// setupAppWatcherCallbacks configures the app watcher callbacks
func (a *App) setupAppWatcherCallbacks() {
	a.appWatcher.OnActivate(func(appName, bundleID string) {
		a.handleAppActivation(bundleID)
	})
	// Watch for display parameter changes (monitor unplug/plug, resolution changes)
	a.appWatcher.OnScreenParametersChanged(func() {
		a.handleScreenParametersChange()
	})
}

// handleScreenParametersChange handles display changes and resizes/regenerates overlays
func (a *App) handleScreenParametersChange() {
	if a.screenChangeProcessing {
		return
	}
	a.screenChangeProcessing = true
	defer func() { a.screenChangeProcessing = false }()

	a.logger.Info("Screen parameters changed; adjusting overlays")

	// Handle grid overlay
	if a.config.Grid.Enabled && a.gridCtx != nil && a.gridCtx.gridOverlay != nil {
		// If grid mode is not active, mark for refresh on next activation
		if a.currentMode != ModeGrid {
			a.gridOverlayNeedsRefresh = true
		} else {
			// Grid mode is active - resize the existing overlay window to match new screen bounds
			gridOverlay := *a.gridCtx.gridOverlay

			// Resize overlay window to current active screen (where mouse is)
			gridOverlay.ResizeToActiveScreenSync()

			// Regenerate the grid cells with updated screen bounds
			if err := a.setupGrid(a.gridCtx.currentAction); err != nil {
				a.logger.Error("Failed to refresh grid after screen change", zap.Error(err))
				return
			}

			a.logger.Info("Grid overlay resized and regenerated for new screen bounds")
		}
	}

	// Handle hint overlay
	if a.config.Hints.Enabled && a.hintOverlay != nil {
		// If hints mode is not active, mark for refresh on next activation
		if a.currentMode != ModeHints {
			a.hintOverlayNeedsRefresh = true
		} else {
			// Hints mode is active - resize the overlay and regenerate hints
			a.hintOverlay.ResizeToActiveScreenSync()

			// Regenerate hints for current action
			a.updateRolesForCurrentApp()
			elements := a.collectElementsForAction(a.currentAction)
			if len(elements) > 0 {
				if err := a.setupHints(elements, a.currentAction); err != nil {
					a.logger.Error("Failed to refresh hints after screen change", zap.Error(err))
					return
				}
				a.logger.Info("Hint overlay resized and regenerated for new screen bounds")
			} else {
				a.logger.Warn("No elements found after screen change")
				a.exitMode()
			}
		}
	}
}

// handleAppActivation handles application activation events
func (a *App) handleAppActivation(bundleID string) {
	a.logger.Debug("App activated", zap.String("bundle_id", bundleID))

	// refresh hotkeys for app
	if a.currentMode == ModeIdle {
		go a.refreshHotkeysForAppOrCurrent(bundleID)
		a.logger.Debug("Handled hotkey refresh")
	} else {
		// Defer hotkey refresh to avoid re-entry during active modes
		a.hotkeyRefreshPending = true
		a.logger.Debug("Deferred hotkey refresh due to active mode")
	}

	if a.config.Hints.Enabled {
		if a.config.Accessibility.AdditionalAXSupport.Enable {
			a.handleAdditionalAccessibility(bundleID)
		}
	}

	a.logger.Debug("Done handling app activation")
}

// handleAdditionalAccessibility handles electron/chromium/firefox accessibility
func (a *App) handleAdditionalAccessibility(bundleID string) {
	cfg := a.config.Accessibility.AdditionalAXSupport

	// handle electrons
	if electron.ShouldEnableElectronSupport(bundleID, cfg.AdditionalElectronBundles) {
		electron.EnsureElectronAccessibility(bundleID)
		a.logger.Debug("Handled electron accessibility")
	}

	// handle chromium
	if electron.ShouldEnableChromiumSupport(bundleID, cfg.AdditionalChromiumBundles) {
		electron.EnsureChromiumAccessibility(bundleID)
		a.logger.Debug("Handled chromium accessibility")
	}

	// handle firefox
	if electron.ShouldEnableFirefoxSupport(bundleID, cfg.AdditionalFirefoxBundles) {
		electron.EnsureFirefoxAccessibility(bundleID)
		a.logger.Debug("Handled firefox accessibility")
	}
}

// printStartupInfo prints startup information to console
func (a *App) printStartupInfo() {
	fmt.Println("✓ Neru is running")

	for k, v := range a.config.Hotkeys.Bindings {
		// Skip showing bindings for disabled modes
		mode := v
		if parts := strings.Split(v, " "); len(parts) > 0 {
			mode = parts[0]
		}
		if mode == "hints" && !a.config.Hints.Enabled {
			continue
		}
		if mode == "grid" && !a.config.Grid.Enabled {
			continue
		}

		toShow := v
		if strings.HasPrefix(v, "exec") {
			runes := []rune(v)
			if len(runes) > 30 {
				toShow = string(runes[:30]) + "..."
			}
		}
		fmt.Printf("  %s: %s\n", k, toShow)
	}
}

// waitForShutdown waits for shutdown signal with force-quit support
func (a *App) waitForShutdown() error {
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
