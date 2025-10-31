package cli

import (
	"github.com/spf13/cobra"
)

var hintsCmd = &cobra.Command{
	Use:   "hints",
	Short: "Launch hints mode",
	Long:  `Activate hint mode for direct clicking on UI elements.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("hints")
	},
}

func init() {
	rootCmd.AddCommand(hintsCmd)
}
