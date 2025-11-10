package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
