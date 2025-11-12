package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/logger"
)

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "Perform actions at the current cursor position",
	Long:  `Execute mouse actions immediately at the current cursor location without target selection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var actionLeftClickCmd = &cobra.Command{
	Use:   "left_click",
	Short: "Perform left click at current cursor position",
	Long:  `Execute a left click at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing left click action at current cursor position")
		var params []string
		params = append(params, "left_click")
		return sendCommand("action", params)
	},
}

var actionRightClickCmd = &cobra.Command{
	Use:   "right_click",
	Short: "Perform right click at current cursor position",
	Long:  `Execute a right click at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing right click action at current cursor position")
		var params []string
		params = append(params, "right_click")
		return sendCommand("action", params)
	},
}

var actionDoubleClickCmd = &cobra.Command{
	Use:   "double_click",
	Short: "Perform double click at current cursor position",
	Long:  `Execute a double click at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing double click action at current cursor position")
		var params []string
		params = append(params, "double_click")
		return sendCommand("action", params)
	},
}

var actionTripleClickCmd = &cobra.Command{
	Use:   "triple_click",
	Short: "Perform triple click at current cursor position",
	Long:  `Execute a triple click at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing triple click action at current cursor position")
		var params []string
		params = append(params, "triple_click")
		return sendCommand("action", params)
	},
}

var actionMouseUpCmd = &cobra.Command{
	Use:   "mouse_up",
	Short: "Release mouse button at current cursor position",
	Long:  `Release the left mouse button at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing mouse up action at current cursor position")
		var params []string
		params = append(params, "mouse_up")
		return sendCommand("action", params)
	},
}

var actionMouseDownCmd = &cobra.Command{
	Use:   "mouse_down",
	Short: "Press mouse button at current cursor position",
	Long:  `Press and hold the left mouse button at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing mouse down action at current cursor position")
		var params []string
		params = append(params, "mouse_down")
		return sendCommand("action", params)
	},
}

var actionMiddleClickCmd = &cobra.Command{
	Use:   "middle_click",
	Short: "Perform middle click at current cursor position",
	Long:  `Execute a middle click at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Executing middle click action at current cursor position")
		var params []string
		params = append(params, "middle_click")
		return sendCommand("action", params)
	},
}

var actionScrollCmd = &cobra.Command{
	Use:   "scroll",
	Short: "Enter scroll mode at current cursor position",
	Long:  `Activate scroll mode at the current cursor location.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Entering scroll mode at current cursor position")
		var params []string
		params = append(params, "scroll")
		return sendCommand("action", params)
	},
}

func init() {
	actionCmd.AddCommand(actionLeftClickCmd)
	actionCmd.AddCommand(actionRightClickCmd)
	actionCmd.AddCommand(actionDoubleClickCmd)
	actionCmd.AddCommand(actionTripleClickCmd)
	actionCmd.AddCommand(actionMouseUpCmd)
	actionCmd.AddCommand(actionMouseDownCmd)
	actionCmd.AddCommand(actionMiddleClickCmd)
	actionCmd.AddCommand(actionScrollCmd)

	// Add action to root
	rootCmd.AddCommand(actionCmd)
}
