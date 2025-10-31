package cli

import (
	"github.com/spf13/cobra"
)

var idleCmd = &cobra.Command{
	Use:   "idle",
	Short: "Set mode to idle",
	Long:  `Exit the current mode and return to idle state.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("idle")
	},
}

func init() {
	rootCmd.AddCommand(idleCmd)
}
