package cli

import (
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Pause the govim program (does not quit)",
	Long:  `Pause the govim program. This disables govim functionality but keeps it running in the background.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("stop")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
