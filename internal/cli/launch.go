package cli

import (
	"github.com/spf13/cobra"
)

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch the govim program",
	Long:  `Launch the govim program. Same as running 'govim' without any subcommand.`,
	Run: func(cmd *cobra.Command, args []string) {
		launchProgram(configPath)
	},
}

func init() {
	rootCmd.AddCommand(launchCmd)
}
