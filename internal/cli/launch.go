package cli

import (
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the neru program",
	Long:  `Launch the neru program. Same as running 'neru' without any subcommand.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchProgram(configPath)
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
