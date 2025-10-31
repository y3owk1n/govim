package cli

import (
	"github.com/spf13/cobra"
)

var hintsActionCmd = &cobra.Command{
	Use:   "hints_action",
	Short: "Launch hints mode with action selection",
	Long:  `Activate hint mode with action selection (choose click type after selecting hint).`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("hints_action")
	},
}

func init() {
	rootCmd.AddCommand(hintsActionCmd)
}
