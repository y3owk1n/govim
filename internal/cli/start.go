package cli

import (
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the govim program (resume if paused)",
	Long:  `Start or resume the govim program. This enables govim if it was previously stopped.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("start")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
