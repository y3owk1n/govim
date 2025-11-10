package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/y3owk1n/neru/internal/ipc"
	"go.uber.org/zap"
)

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
		// Skip registering bindings for disabled modes
		mode := action
		if parts := strings.Split(action, " "); len(parts) > 0 {
			mode = parts[0]
		}
		if mode == "hints" && !a.config.Hints.Enabled {
			continue
		}
		if mode == "grid" && !a.config.Grid.Enabled {
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

				if err := a.executeHotkeyAction(bindKey, bindAction); err != nil {
					a.logger.Error("hotkey action failed",
						zap.String("key", bindKey),
						zap.String("action", bindAction),
						zap.Error(err))
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

// executeHotkeyAction executes a hotkey action (either exec or IPC command)
func (a *App) executeHotkeyAction(key, action string) error {
	// Exec mode: run arbitrary bash command
	if strings.HasPrefix(action, "exec ") {
		return a.executeShellCommand(key, action)
	}

	// Split action into action and params
	actionParts := strings.Split(action, " ")
	action = actionParts[0]
	params := actionParts[1:]

	// Otherwise treat the action as an internal neru command and dispatch it
	resp := a.handleIPCCommand(ipc.Command{Action: action, Args: params})
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}

	a.logger.Info("hotkey action executed", zap.String("key", key), zap.String("action", action))
	return nil
}

// executeShellCommand executes a shell command from a hotkey
func (a *App) executeShellCommand(key, action string) error {
	cmdStr := strings.TrimSpace(strings.TrimPrefix(action, "exec"))
	if cmdStr == "" {
		a.logger.Error("hotkey exec has empty command", zap.String("key", key))
		return fmt.Errorf("empty command")
	}

	a.logger.Debug("Executing shell command from hotkey", zap.String("key", key), zap.String("cmd", cmdStr))
	cmd := exec.Command("/bin/bash", "-lc", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		a.logger.Error("hotkey exec failed", zap.String("key", key), zap.String("cmd", cmdStr), zap.ByteString("output", out), zap.Error(err))
		return err
	}

	a.logger.Info("hotkey exec completed", zap.String("key", key), zap.String("cmd", cmdStr), zap.ByteString("output", out))
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
