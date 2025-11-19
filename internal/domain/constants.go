package domain

// Mode represents the operational modes of the application.
type Mode string

// Mode constants for IPC and hotkey bindings.
// These are used to identify modes in configuration and command handling.
const (
	ModeHints Mode = "hints"
	ModeGrid  Mode = "grid"
	ModeIdle  Mode = "idle"
)

// Command represents IPC command names used for inter-process communication.
// These constants ensure consistency across CLI and daemon communication.
const (
	CommandPing   = "ping"
	CommandStart  = "start"
	CommandStop   = "stop"
	CommandAction = "action"
	CommandStatus = "status"
	CommandConfig = "config"
)

// Special action prefix for shell command execution.
// Hotkeys with actions starting with this prefix will execute shell commands.
const (
	ActionPrefixExec = "exec "
)
