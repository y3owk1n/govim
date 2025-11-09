package cli

import (
	"github.com/spf13/cobra"
)

var hintsCmd = &cobra.Command{
	Use:   "hints",
	Short: "Launch hints mode in left click mode",
	Long:  `Activate hint mode for direct clicking on UI elements.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var hintsLeftClickCmd = &cobra.Command{
	Use:   "left_click",
	Short: "Launch hints mode with left click",
	Long:  `Activate hint mode with left click (default).`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "left_click")
		return sendCommand("hints", params)
	},
}

var hintsRightClickCmd = &cobra.Command{
	Use:   "right_click",
	Short: "Launch hints mode with right click",
	Long:  `Activate hint mode with right click.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "right_click")
		return sendCommand("hints", params)
	},
}

var hintsDoubleClickCmd = &cobra.Command{
	Use:   "double_click",
	Short: "Launch hints mode with double click",
	Long:  `Activate hint mode with double click.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "double_click")
		return sendCommand("hints", params)
	},
}

var hintsTripleClickCmd = &cobra.Command{
	Use:   "triple_click",
	Short: "Launch hints mode with triple click",
	Long:  `Activate hint mode with triple click.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "triple_click")
		return sendCommand("hints", params)
	},
}

var hintsMouseUpCmd = &cobra.Command{
	Use:   "mouse_up",
	Short: "Launch hints mode with mouse up",
	Long:  `Activate hint mode with mouse up.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "mouse_up")
		return sendCommand("hints", params)
	},
}

var hintsMouseDownCmd = &cobra.Command{
	Use:   "mouse_down",
	Short: "Launch hints mode with mouse down",
	Long:  `Activate hint mode with mouse down.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "mouse_down")
		return sendCommand("hints", params)
	},
}

var hintsMiddleClickCmd = &cobra.Command{
	Use:   "middle_click",
	Short: "Launch hints mode with middle click",
	Long:  `Activate hint mode with middle click.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "middle_click")
		return sendCommand("hints", params)
	},
}

var hintsMoveMouseCmd = &cobra.Command{
	Use:   "move_mouse",
	Short: "Launch hints mode with move mouse",
	Long:  `Activate hint mode with move mouse.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "move_mouse")
		return sendCommand("hints", params)
	},
}

var hintsScrollCmd = &cobra.Command{
	Use:   "scroll",
	Short: "Launch hints mode with scroll",
	Long:  `Activate hint mode with scroll.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "scroll")
		return sendCommand("hints", params)
	},
}

var hintsContextMenuCmd = &cobra.Command{
	Use:   "context_menu",
	Short: "Launch hints mode with context menu",
	Long:  `Activate hint mode with context menu.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var params []string
		params = append(params, "context_menu")
		return sendCommand("hints", params)
	},
}

func init() {
	hintsCmd.AddCommand(hintsLeftClickCmd)
	hintsCmd.AddCommand(hintsRightClickCmd)
	hintsCmd.AddCommand(hintsDoubleClickCmd)
	hintsCmd.AddCommand(hintsTripleClickCmd)
	hintsCmd.AddCommand(hintsMouseUpCmd)
	hintsCmd.AddCommand(hintsMouseDownCmd)
	hintsCmd.AddCommand(hintsMiddleClickCmd)
	hintsCmd.AddCommand(hintsMoveMouseCmd)
	hintsCmd.AddCommand(hintsScrollCmd)
	hintsCmd.AddCommand(hintsContextMenuCmd)

	// Add hints to root
	rootCmd.AddCommand(hintsCmd)
}
