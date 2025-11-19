package app

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/domain"
	"github.com/y3owk1n/neru/internal/infra/electron"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"go.uber.org/zap"
)

// Run starts the main application loop and initializes all subsystems.
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

// setupAppWatcherCallbacks configures callbacks for application watcher events.
func (a *App) setupAppWatcherCallbacks() {
	a.appWatcher.OnActivate(func(_, bundleID string) {
		a.handleAppActivation(bundleID)
	})
	// Watch for display parameter changes (monitor unplug/plug, resolution changes)
	a.appWatcher.OnScreenParametersChanged(func() {
		a.handleScreenParametersChange()
	})
}

// handleScreenParametersChange responds to display configuration changes by updating overlays.
func (a *App) handleScreenParametersChange() {
	if a.state.ScreenChangeProcessing() {
		return
	}
	a.state.SetScreenChangeProcessing(true)
	defer func() { a.state.SetScreenChangeProcessing(false) }()

	a.logger.Info("Screen parameters changed; adjusting overlays")
	if a.overlayManager != nil {
		a.overlayManager.ResizeToActiveScreenSync()
	}

	// Handle grid overlay
	if a.config.Grid.Enabled && a.gridCtx != nil && a.gridCtx.GridOverlay != nil {
		// If grid mode is not active, mark for refresh on next activation
		if a.state.CurrentMode() != ModeGrid {
			a.state.SetGridOverlayNeedsRefresh(true)
		} else {
			// Grid mode is active - resize the existing overlay window to match new screen bounds
			// Resize overlay window to current active screen (where mouse is)
			if a.overlayManager != nil {
				a.overlayManager.ResizeToActiveScreenSync()
			}

			// Regenerate the grid cells with updated screen bounds
			err := a.setupGrid()
			if err != nil {
				a.logger.Error("Failed to refresh grid after screen change", zap.Error(err))
				return
			}

			a.logger.Info("Grid overlay resized and regenerated for new screen bounds")
		}
	}

	// Handle hint overlay
	if a.config.Hints.Enabled && a.hintOverlay != nil {
		// If hints mode is not active, mark for refresh on next activation
		if a.state.CurrentMode() != ModeHints {
			a.state.SetHintOverlayNeedsRefresh(true)
		} else {
			// Hints mode is active - resize the overlay and regenerate hints
			if a.overlayManager != nil {
				a.overlayManager.ResizeToActiveScreenSync()
			}

			// Regenerate hints for current action
			a.updateRolesForCurrentApp()
			elements := a.collectElements()
			if len(elements) > 0 {
				err := a.setupHints(elements)
				if err != nil {
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

	// Resize scroll overlay to current active screen (where mouse is)
	if a.overlayManager != nil {
		a.overlayManager.ResizeToActiveScreenSync()
	}
}

// handleAppActivation responds to application activation events.
func (a *App) handleAppActivation(bundleID string) {
	a.logger.Debug("App activated", zap.String("bundle_id", bundleID))

	// refresh hotkeys for app
	if a.state.CurrentMode() == ModeIdle {
		go a.refreshHotkeysForAppOrCurrent(bundleID)
		a.logger.Debug("Handled hotkey refresh")
	} else {
		// Defer hotkey refresh to avoid re-entry during active modes
		a.state.SetHotkeyRefreshPending(true)
		a.logger.Debug("Deferred hotkey refresh due to active mode")
	}

	if a.config.Hints.Enabled {
		if a.config.Hints.AdditionalAXSupport.Enable {
			a.handleAdditionalAccessibility(bundleID)
		}
	}

	a.logger.Debug("Done handling app activation")
}

// handleAdditionalAccessibility configures accessibility support for Electron/Chromium/Firefox applications.
func (a *App) handleAdditionalAccessibility(bundleID string) {
	cfg := a.config.Hints.AdditionalAXSupport

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

// printStartupInfo displays startup information including registered hotkeys.
func (a *App) printStartupInfo() {
	a.logger.Info("✓ Neru is running")

	for key, value := range a.config.Hotkeys.Bindings {
		// Skip showing bindings for disabled modes
		mode := value
		if parts := strings.Split(value, " "); len(parts) > 0 {
			mode = parts[0]
		}
		if mode == domain.ModeHints && !a.config.Hints.Enabled {
			continue
		}
		if mode == domain.ModeGrid && !a.config.Grid.Enabled {
			continue
		}

		toShow := value
		if strings.HasPrefix(value, "exec") {
			runes := []rune(value)
			if len(runes) > 30 {
				toShow = string(runes[:30]) + "..."
			}
		}
		a.logger.Info(fmt.Sprintf("  %s: %s", key, toShow))
	}
}

// waitForShutdown waits for shutdown signals and handles graceful termination.
func (a *App) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// First signal: graceful shutdown
	<-sigChan
	a.logger.Info("Received shutdown signal, starting graceful shutdown...")
	a.logger.Info("\n⚠️  Shutting down gracefully... (press Ctrl+C again to force quit)")

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
		a.logger.Info("⚠️  Force quitting...")
		os.Exit(1)
	case <-time.After(10 * time.Second):
		// Timeout: force quit
		a.logger.Error("Shutdown timeout exceeded, forcing shutdown")
		a.logger.Info("⚠️  Shutdown timeout, force quitting...")
		os.Exit(1)
	}

	return nil
}

// Cleanup cleans up resources.
func (a *App) Cleanup() {
	a.logger.Info("Cleaning up")

	// Exit any active mode first
	a.exitMode()

	// Stop IPC server first to prevent new requests
	if a.ipcServer != nil {
		err := a.ipcServer.Stop()
		if err != nil {
			a.logger.Error("Failed to stop IPC server", zap.Error(err))
		}
	}

	// Unregister all hotkeys
	if a.hotkeyManager != nil {
		a.hotkeyManager.UnregisterAll()
	}

	// Clean up centralized overlay window
	if a.overlayManager != nil {
		a.overlayManager.Destroy()
	}

	// Cleanup event tap
	if a.eventTap != nil {
		a.eventTap.Destroy()
	}

	// Sync and close logger at the end
	err := logger.Sync()
	if err != nil {
		// Ignore "inappropriate ioctl for device" error which occurs when syncing stdout/stderr
		if !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			a.logger.Error("Failed to sync logger", zap.Error(err))
		}
	}

	// Stop app watcher
	a.appWatcher.Stop()

	// Close logger (syncs and closes log file)
	err2 := logger.Close()
	if err2 != nil {
		// Can't log this since logger is being closed
		fmt.Fprintf(os.Stderr, "Warning: failed to close logger: %v\n", err2)
	}
}
