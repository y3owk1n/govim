package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/logger"
)

// idleCmd represents the command to set the application mode to idle.
// This exits any active mode and returns the application to its default state.
var idleCmd = &cobra.Command{
	Use:   "idle",
	Short: "Set mode to idle",
	Long:  `Exit the current mode and return to idle state.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, args []string) error {
		logger.Debug("Setting mode to idle")
		return sendCommand("idle", args)
	},
}

func init() {
	rootCmd.AddCommand(idleCmd)
}
