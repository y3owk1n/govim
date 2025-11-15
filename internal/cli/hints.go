package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/logger"
)

var hintsCmd = &cobra.Command{
	Use:   "hints",
	Short: "Launch hints mode in left click mode",
	Long:  `Activate hint mode for direct clicking on UI elements.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Launching hints mode")
		var params []string
		params = append(params, "hints")
		return sendCommand("hints", params)
	},
}

func init() {
	rootCmd.AddCommand(hintsCmd)
}
