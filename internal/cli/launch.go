package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"go.uber.org/zap"
)

// launchCmd represents the command to launch the Neru daemon process.
// This starts the main application in the background, making it ready to receive commands.
var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the neru program",
	Long:  `Launch the neru program. Same as running 'neru' without any subcommand.`,
	Run: func(_ *cobra.Command, _ []string) {
		logger.Debug("Launching program", zap.String("config_path", configPath))
		launchProgram(configPath)
	},
}

func init() {
	launchCmd.Flags().StringVar(&configPath, "config", "", "config file path")

	rootCmd.AddCommand(launchCmd)
}
