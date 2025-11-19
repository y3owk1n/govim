// Package cli provides the command-line interface for the Neru application.
//
// This package implements a comprehensive CLI using the Cobra framework that allows
// users to control the Neru daemon externally. The CLI serves as the primary interface
// for automation, scripting, and manual control of Neru's functionality.
//
// Key Features:
//   - Daemon Control: Start, stop, and manage the Neru daemon lifecycle
//   - Mode Activation: Trigger hints, grid, and scroll modes remotely
//   - Status Reporting: Check daemon status and current mode
//   - Configuration Inspection: View current daemon configuration
//   - Action Execution: Perform mouse actions at the current cursor position
//   - Shell Integration: Generate shell completion scripts for bash, zsh, and fish
//
// The CLI communicates with the Neru daemon through a Unix domain socket IPC system,
// ensuring fast, reliable communication. All CLI commands are designed to be scriptable
// and return structured output for automation purposes.
//
// Command Structure:
//   - neru launch: Start the Neru daemon with optional configuration
//   - neru start: Resume a paused daemon
//   - neru stop: Pause the daemon (disable functionality but keep running)
//   - neru idle: Return to idle state (cancel any active mode)
//   - neru hints: Activate hint mode
//   - neru grid: Activate grid mode
//   - neru action: Perform actions at the current cursor position
//   - neru scroll: Activate scroll mode
//   - neru status: Check daemon status and current mode
//   - neru config: Print the current daemon configuration
//   - neru completion: Generate shell completion scripts
//
// Error Handling:
// All CLI commands return structured error messages with machine-readable error codes
// to facilitate scripting and automation. Common error codes include:
//   - ERR_DAEMON_NOT_RUNNING: Daemon is not running
//   - ERR_MODE_DISABLED: Requested mode is disabled in configuration
//   - ERR_UNKNOWN_COMMAND: Invalid command provided
//
// The CLI is designed to be lightweight and fast, with minimal startup overhead
// to enable responsive automation and scripting workflows.
package cli
