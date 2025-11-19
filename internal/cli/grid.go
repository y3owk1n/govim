package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/infra/logger"
)

// gridCmd represents the command to activate grid mode for mouseless navigation.
// Grid mode divides the screen into cells for precise cursor positioning.
var gridCmd = &cobra.Command{
	Use:   "grid",
	Short: "Launch grid mode",
	Long:  `Activate grid mode for mouseless navigation.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Launching grid mode")
		var params []string
		params = append(params, "grid")
		return sendCommand("grid", params)
	},
}

func init() {
	rootCmd.AddCommand(gridCmd)
}
