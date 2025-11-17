package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/ipc"
	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Dump effective config",
	Long:  "Print the currently active Neru configuration as JSON.",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Fetching config")
		client := ipc.NewClient()
		response, err := client.Send(ipc.Command{Action: "config"})
		if err != nil {
			return fmt.Errorf("failed to send config command: %w", err)
		}

		if !response.Success {
			if response.Code != "" {
				return fmt.Errorf("%s (code: %s)", response.Message, response.Code)
			}
			return fmt.Errorf("%s", response.Message)
		}

		// Marshal pretty JSON
		jsonData, err := json.MarshalIndent(response.Data, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal config to JSON", zap.Error(err))
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		logger.Info(string(jsonData))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
