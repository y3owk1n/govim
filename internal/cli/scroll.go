package cli

import (
	"github.com/spf13/cobra"
)

var scrollCmd = &cobra.Command{
	Use:   "scroll",
	Short: "Launch scroll mode",
	Long:  `Activate scroll mode for vim-style scrolling.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("scroll")
	},
}

func init() {
	rootCmd.AddCommand(scrollCmd)
}
