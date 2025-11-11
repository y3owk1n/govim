package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/ipc"
	"go.uber.org/zap"
)

// handleIPCCommand handles IPC commands from the CLI
func (a *App) handleIPCCommand(cmd ipc.Command) ipc.Response {
	a.logger.Info("Handling IPC command", zap.String("action", cmd.Action), zap.String("params", strings.Join(cmd.Args, ", ")))

	switch cmd.Action {
	case "ping":
		return a.handlePing(cmd)
	case "start":
		return a.handleStart(cmd)
	case "stop":
		return a.handleStop(cmd)
	case "hints":
		return a.handleHints(cmd)
	case "grid":
		return a.handleGrid(cmd)
	case "action":
		return a.handleAction(cmd)
	case "idle":
		return a.handleIdle(cmd)
	case "status":
		return a.handleStatus(cmd)
	default:
		return ipc.Response{Success: false, Message: fmt.Sprintf("unknown command: %s", cmd.Action)}
	}
}

func (a *App) handlePing(cmd ipc.Command) ipc.Response {
	return ipc.Response{Success: true, Message: "pong"}
}

func (a *App) handleStart(cmd ipc.Command) ipc.Response {
	if a.enabled {
		return ipc.Response{Success: false, Message: "neru is already running"}
	}
	a.enabled = true
	return ipc.Response{Success: true, Message: "neru started"}
}

func (a *App) handleStop(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{Success: false, Message: "neru is already stopped"}
	}
	a.enabled = false
	a.exitMode()
	return ipc.Response{Success: true, Message: "neru stopped"}
}

func (a *App) handleHints(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{Success: false, Message: "neru is not running"}
	}
	if !a.config.Hints.Enabled {
		return ipc.Response{Success: false, Message: "hints mode is disabled by config"}
	}

	// Parse params
	params := cmd.Args
	for _, param := range params {
		switch param {
		case "left_click":
			a.activateMode(ModeHints, ActionLeftClick)
		case "right_click":
			a.activateMode(ModeHints, ActionRightClick)
		case "double_click":
			a.activateMode(ModeHints, ActionDoubleClick)
		case "triple_click":
			a.activateMode(ModeHints, ActionTripleClick)
		case "mouse_up":
			a.activateMode(ModeHints, ActionMouseUp)
		case "mouse_down":
			a.activateMode(ModeHints, ActionMouseDown)
		case "middle_click":
			a.activateMode(ModeHints, ActionMiddleClick)
		case "move_mouse":
			a.activateMode(ModeHints, ActionMoveMouse)
		case "scroll":
			a.activateMode(ModeHints, ActionScroll)
		case "context_menu":
			a.activateMode(ModeHints, ActionContextMenu)
		default:
			return ipc.Response{Success: false, Message: fmt.Sprintf("unknown hints mode: %s", param)}
		}
	}

	return ipc.Response{Success: true, Message: "hint mode activated"}
}

func (a *App) handleGrid(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{Success: false, Message: "neru is not running"}
	}
	if !a.config.Grid.Enabled {
		return ipc.Response{Success: false, Message: "grid mode is disabled by config"}
	}

	// Parse params
	params := cmd.Args
	for _, param := range params {
		switch param {
		case "left_click":
			a.activateMode(ModeGrid, ActionLeftClick)
		case "right_click":
			a.activateMode(ModeGrid, ActionRightClick)
		case "double_click":
			a.activateMode(ModeGrid, ActionDoubleClick)
		case "triple_click":
			a.activateMode(ModeGrid, ActionTripleClick)
		case "mouse_up":
			a.activateMode(ModeGrid, ActionMouseUp)
		case "mouse_down":
			a.activateMode(ModeGrid, ActionMouseDown)
		case "middle_click":
			a.activateMode(ModeGrid, ActionMiddleClick)
		case "move_mouse":
			a.activateMode(ModeGrid, ActionMoveMouse)
		case "scroll":
			a.activateMode(ModeGrid, ActionScroll)
		case "context_menu":
			a.activateMode(ModeGrid, ActionContextMenu)
		default:
			return ipc.Response{Success: false, Message: fmt.Sprintf("unknown grid action: %s", param)}
		}
	}

	return ipc.Response{Success: true, Message: "grid mode activated"}
}

func (a *App) handleAction(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{Success: false, Message: "neru is not running"}
	}

	// Parse params
	params := cmd.Args
	if len(params) == 0 {
		return ipc.Response{Success: false, Message: "no action specified"}
	}

	// Get the current cursor position
	cursorPos := accessibility.GetCurrentCursorPosition()

	for _, param := range params {
		var err error
		switch param {
		case "left_click":
			err = accessibility.LeftClickAtPoint(cursorPos, false)
		case "right_click":
			err = accessibility.RightClickAtPoint(cursorPos, false)
		case "double_click":
			err = accessibility.DoubleClickAtPoint(cursorPos, false)
		case "triple_click":
			err = accessibility.TripleClickAtPoint(cursorPos, false)
		case "mouse_up":
			err = accessibility.LeftMouseUpAtPoint(cursorPos)
		case "mouse_down":
			err = accessibility.LeftMouseDownAtPoint(cursorPos)
		case "middle_click":
			err = accessibility.MiddleClickAtPoint(cursorPos, false)
		case "scroll":
			// Enable event tap and let user scroll interactively at current position
			// Resize overlay to active screen for multi-monitor support
			a.hintOverlay.ResizeToActiveScreen()

			// Draw highlight border if enabled
			if a.config.Scroll.HighlightScrollArea {
				a.drawScrollHighlightBorder()
				a.hintOverlay.Show()
			}

			// Enable event tap for scroll key handling
			if a.eventTap != nil {
				a.eventTap.Enable()
			}

			a.logger.Info("Interactive scroll activated")
			a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom, Esc to exit")
			return ipc.Response{Success: true, Message: "scroll mode activated"}
		default:
			return ipc.Response{Success: false, Message: fmt.Sprintf("unknown action: %s", param)}
		}

		if err != nil {
			return ipc.Response{Success: false, Message: fmt.Sprintf("action failed: %v", err)}
		}
	}

	return ipc.Response{Success: true, Message: "action performed at cursor"}
}

func (a *App) handleIdle(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{Success: false, Message: "neru is not running"}
	}
	a.exitMode()
	return ipc.Response{Success: true, Message: "mode set to idle"}
}

func (a *App) handleStatus(cmd ipc.Command) ipc.Response {
	cfgPath := a.resolveConfigPath()
	statusData := map[string]any{
		"enabled": a.enabled,
		"mode":    a.getCurrModeString(),
		"config":  cfgPath,
	}
	return ipc.Response{Success: true, Data: statusData}
}

// resolveConfigPath resolves the config path for status display
func (a *App) resolveConfigPath() string {
	cfgPath := a.ConfigPath

	if cfgPath == "" {
		// Fallback to the standard config path if daemon wasn't started
		// with an explicit --config
		cfgPath = config.FindConfigFile()
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return "No config file found, using default config without config file"
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

	return cfgPath
}
