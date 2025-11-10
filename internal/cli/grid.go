package cli

import (
	"github.com/spf13/cobra"
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
		var params []string
		params = append(params, "double_click")
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
		var params []string
		params = append(params, "move_mouse")
		return sendCommand("grid", params)
	},
}

func init() {
	gridCmd.AddCommand(gridLeftClickCmd)
	gridCmd.AddCommand(gridRightClickCmd)
	gridCmd.AddCommand(gridDoubleClickCmd)
	gridCmd.AddCommand(gridMoveMouseCmd)

	// Add grid to root
	rootCmd.AddCommand(gridCmd)
}
