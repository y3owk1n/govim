package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/y3owk1n/neru/internal/accessibility"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/ipc"
	"go.uber.org/zap"
)

// handleIPCCommand handles IPC commands from the CLI.
func (a *App) handleIPCCommand(cmd ipc.Command) ipc.Response {
	a.logger.Info(
		"Handling IPC command",
		zap.String("action", cmd.Action),
		zap.String("args", strings.Join(cmd.Args, ", ")),
	)

	switch cmd.Action {
	case "ping":
		return a.handlePing(cmd)
	case "start":
		return a.handleStart(cmd)
	case "stop":
		return a.handleStop(cmd)
	case modeHints:
		return a.handleHints(cmd)
	case modeGrid:
		return a.handleGrid(cmd)
	case "action":
		return a.handleAction(cmd)
	case "idle":
		return a.handleIdle(cmd)
	case "status":
		return a.handleStatus(cmd)
	case "config":
		return a.handleConfig(cmd)
	default:
		return ipc.Response{
			Success: false,
			Message: "unknown command: " + cmd.Action,
			Code:    ipc.CodeUnknownCommand,
		}
	}
}

func (a *App) handlePing(_ ipc.Command) ipc.Response {
	return ipc.Response{Success: true, Message: "pong", Code: ipc.CodeOK}
}

func (a *App) handleStart(_ ipc.Command) ipc.Response {
	if a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is already running",
			Code:    ipc.CodeAlreadyRunning,
		}
	}
	a.enabled = true
	return ipc.Response{Success: true, Message: "neru started", Code: ipc.CodeOK}
}

func (a *App) handleStop(_ ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is already stopped",
			Code:    ipc.CodeNotRunning,
		}
	}
	a.enabled = false
	a.exitMode()
	return ipc.Response{Success: true, Message: "neru stopped", Code: ipc.CodeOK}
}

func (a *App) handleHints(_ ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is not running",
			Code:    ipc.CodeNotRunning,
		}
	}
	if !a.config.Hints.Enabled {
		return ipc.Response{
			Success: false,
			Message: "hints mode is disabled by config",
			Code:    ipc.CodeModeDisabled,
		}
	}

	a.activateMode(ModeHints)

	return ipc.Response{Success: true, Message: "hint mode activated", Code: ipc.CodeOK}
}

func (a *App) handleGrid(_ ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is not running",
			Code:    ipc.CodeNotRunning,
		}
	}
	if !a.config.Grid.Enabled {
		return ipc.Response{
			Success: false,
			Message: "grid mode is disabled by config",
			Code:    ipc.CodeModeDisabled,
		}
	}

	a.activateMode(ModeGrid)

	return ipc.Response{Success: true, Message: "grid mode activated", Code: ipc.CodeOK}
}

func (a *App) handleAction(cmd ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is not running",
			Code:    ipc.CodeNotRunning,
		}
	}

	// Parse params
	params := cmd.Args
	if len(params) == 0 {
		return ipc.Response{
			Success: false,
			Message: "no action specified",
			Code:    ipc.CodeInvalidInput,
		}
	}

	// Get the current cursor position
	cursorPos := accessibility.GetCurrentCursorPosition()

	for _, param := range params {
		var err error
		switch param {
		case "scroll":
			a.startInteractiveScroll()
			return ipc.Response{Success: true, Message: "scroll mode activated", Code: ipc.CodeOK}
		default:
			if !isKnownAction(param) {
				return ipc.Response{
					Success: false,
					Message: "unknown action: " + param,
					Code:    ipc.CodeInvalidInput,
				}
			}
			err = performActionAtPoint(param, cursorPos)
		}

		if err != nil {
			return ipc.Response{
				Success: false,
				Message: "action failed: " + err.Error(),
				Code:    ipc.CodeActionFailed,
			}
		}
	}

	return ipc.Response{Success: true, Message: "action performed at cursor", Code: ipc.CodeOK}
}

func (a *App) handleIdle(_ ipc.Command) ipc.Response {
	if !a.enabled {
		return ipc.Response{
			Success: false,
			Message: "neru is not running",
			Code:    ipc.CodeNotRunning,
		}
	}
	a.exitMode()
	return ipc.Response{Success: true, Message: "mode set to idle", Code: ipc.CodeOK}
}

func (a *App) handleStatus(_ ipc.Command) ipc.Response {
	cfgPath := a.resolveConfigPath()
	statusData := ipc.StatusData{
		Enabled: a.enabled,
		Mode:    a.getCurrModeString(),
		Config:  cfgPath,
	}
	return ipc.Response{Success: true, Data: statusData, Code: ipc.CodeOK}
}

func (a *App) handleConfig(_ ipc.Command) ipc.Response {
	if a.config == nil {
		return ipc.Response{Success: false, Message: "config unavailable", Code: ipc.CodeNotRunning}
	}
	return ipc.Response{Success: true, Data: a.config, Code: ipc.CodeOK}
}

// resolveConfigPath resolves the config path for status display.
func (a *App) resolveConfigPath() string {
	cfgPath := a.ConfigPath

	if cfgPath == "" {
		// Fallback to the standard config path if daemon wasn't started
		// with an explicit --config
		cfgPath = config.FindConfigFile()
	}

	// If config file doesn't exist, return default config
	var err error
	_, err = os.Stat(cfgPath)
	if os.IsNotExist(err) {
		return "No config file found, using default config without config file"
	}

	// Expand ~ to home dir and resolve relative paths to absolute
	if strings.HasPrefix(cfgPath, "~") {
		var home string
		var err error
		home, err = os.UserHomeDir()
		if err == nil {
			cfgPath = filepath.Join(home, cfgPath[1:])
		}
	}
	var abs string
	var err2 error
	abs, err2 = filepath.Abs(cfgPath)
	if err2 == nil {
		cfgPath = abs
	}

	return cfgPath
}
