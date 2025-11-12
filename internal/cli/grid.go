package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/logger"
)

var gridCmd = &cobra.Command{
	Use:   "grid",
	Short: "Launch grid mode",
	Long:  `Activate grid mode for mouseless navigation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var gridLeftClickCmd = &cobra.Command{
	Use:   "left_click",
	Short: "Launch grid mode with left click",
	Long:  `Activate grid mode with left click action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with left click action")
		var params []string
		params = append(params, "left_click")
		return sendCommand("grid", params)
	},
}

var gridRightClickCmd = &cobra.Command{
	Use:   "right_click",
	Short: "Launch grid mode with right click",
	Long:  `Activate grid mode with right click action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with right click action")
		var params []string
		params = append(params, "right_click")
		return sendCommand("grid", params)
	},
}

var gridDoubleClickCmd = &cobra.Command{
	Use:   "double_click",
	Short: "Launch grid mode with double click",
	Long:  `Activate grid mode with double click action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with double click action")
		var params []string
		params = append(params, "double_click")
		return sendCommand("grid", params)
	},
}

var gridTripleClickCmd = &cobra.Command{
	Use:   "triple_click",
	Short: "Launch grid mode with triple click",
	Long:  `Activate grid mode with triple click action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with triple click action")
		var params []string
		params = append(params, "triple_click")
		return sendCommand("grid", params)
	},
}

var gridMouseUpCmd = &cobra.Command{
	Use:   "mouse_up",
	Short: "Launch grid mode with mouse up",
	Long:  `Activate grid mode with mouse up action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with mouse up action")
		var params []string
		params = append(params, "mouse_up")
		return sendCommand("grid", params)
	},
}

var gridMouseDownCmd = &cobra.Command{
	Use:   "mouse_down",
	Short: "Launch grid mode with mouse down",
	Long:  `Activate grid mode with mouse down action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with mouse down action")
		var params []string
		params = append(params, "mouse_down")
		return sendCommand("grid", params)
	},
}

var gridMiddleClickCmd = &cobra.Command{
	Use:   "middle_click",
	Short: "Launch grid mode with middle click",
	Long:  `Activate grid mode with middle click action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with middle click action")
		var params []string
		params = append(params, "middle_click")
		return sendCommand("grid", params)
	},
}

var gridMoveMouseCmd = &cobra.Command{
	Use:   "move_mouse",
	Short: "Launch grid mode with move mouse",
	Long:  `Activate grid mode with move mouse action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with move mouse action")
		var params []string
		params = append(params, "move_mouse")
		return sendCommand("grid", params)
	},
}

var gridScrollCmd = &cobra.Command{
	Use:   "scroll",
	Short: "Launch grid mode with scroll",
	Long:  `Activate grid mode with scroll action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with scroll action")
		var params []string
		params = append(params, "scroll")
		return sendCommand("grid", params)
	},
}

var gridContextMenuCmd = &cobra.Command{
	Use:   "context_menu",
	Short: "Launch grid mode with context menu",
	Long:  `Activate grid mode with context menu action.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Debug("Launching grid mode with context menu action")
		var params []string
		params = append(params, "context_menu")
		return sendCommand("grid", params)
	},
}

func init() {
	gridCmd.AddCommand(gridLeftClickCmd)
	gridCmd.AddCommand(gridRightClickCmd)
	gridCmd.AddCommand(gridDoubleClickCmd)
	gridCmd.AddCommand(gridTripleClickCmd)
	gridCmd.AddCommand(gridMouseUpCmd)
	gridCmd.AddCommand(gridMouseDownCmd)
	gridCmd.AddCommand(gridMiddleClickCmd)
	gridCmd.AddCommand(gridMoveMouseCmd)
	gridCmd.AddCommand(gridScrollCmd)
	gridCmd.AddCommand(gridContextMenuCmd)

	// Add grid to root
	rootCmd.AddCommand(gridCmd)
}
