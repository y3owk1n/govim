package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/logger"
)

// actionCmd represents the parent command for immediate cursor actions.
// It provides subcommands for performing mouse actions at the current cursor position.
var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "Perform actions at the current cursor position",
	Long:  `Execute mouse actions immediately at the current cursor location without target selection.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

// actionLeftClickCmd handles the immediate left click action command.
// It sends a left click command to the daemon to be executed at the current cursor position.
var actionLeftClickCmd = &cobra.Command{
	Use:   "left_click",
	Short: "Perform left click at current cursor position",
	Long:  `Execute a left click at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Executing left click action at current cursor position")
		var params []string
		params = append(params, "left_click")
		return sendCommand("action", params)
	},
}

// actionRightClickCmd handles the immediate right click action command.
// It sends a right click command to the daemon to be executed at the current cursor position.
var actionRightClickCmd = &cobra.Command{
	Use:   "right_click",
	Short: "Perform right click at current cursor position",
	Long:  `Execute a right click at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Executing right click action at current cursor position")
		var params []string
		params = append(params, "right_click")
		return sendCommand("action", params)
	},
}

// actionMouseUpCmd handles the mouse button release action command.
// It sends a mouse up command to the daemon to release the left mouse button.
var actionMouseUpCmd = &cobra.Command{
	Use:   "mouse_up",
	Short: "Release mouse button at current cursor position",
	Long:  `Release the left mouse button at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Executing mouse up action at current cursor position")
		var params []string
		params = append(params, "mouse_up")
		return sendCommand("action", params)
	},
}

// actionMouseDownCmd handles the mouse button press action command.
// It sends a mouse down command to the daemon to press and hold the left mouse button.
var actionMouseDownCmd = &cobra.Command{
	Use:   "mouse_down",
	Short: "Press mouse button at current cursor position",
	Long:  `Press and hold the left mouse button at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Executing mouse down action at current cursor position")
		var params []string
		params = append(params, "mouse_down")
		return sendCommand("action", params)
	},
}

// actionMiddleClickCmd handles the immediate middle click action command.
// It sends a middle click command to the daemon to be executed at the current cursor position.
var actionMiddleClickCmd = &cobra.Command{
	Use:   "middle_click",
	Short: "Perform middle click at current cursor position",
	Long:  `Execute a middle click at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Executing middle click action at current cursor position")
		var params []string
		params = append(params, "middle_click")
		return sendCommand("action", params)
	},
}

// actionScrollCmd handles the scroll mode activation command.
// It sends a scroll command to the daemon to enter scroll mode at the current cursor position.
var actionScrollCmd = &cobra.Command{
	Use:   "scroll",
	Short: "Enter scroll mode at current cursor position",
	Long:  `Activate scroll mode at the current cursor location.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Entering scroll mode at current cursor position")
		var params []string
		params = append(params, "scroll")
		return sendCommand("action", params)
	},
}

func init() {
	actionCmd.AddCommand(actionLeftClickCmd)
	actionCmd.AddCommand(actionRightClickCmd)
	actionCmd.AddCommand(actionMouseUpCmd)
	actionCmd.AddCommand(actionMouseDownCmd)
	actionCmd.AddCommand(actionMiddleClickCmd)
	actionCmd.AddCommand(actionScrollCmd)

	// Add action to root
	rootCmd.AddCommand(actionCmd)
}
