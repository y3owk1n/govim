package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/infra/ipc"
	"github.com/y3owk1n/neru/internal/infra/logger"
	"go.uber.org/zap"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show neru status",
	Long:  `Display the current status of the neru program.`,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		logger.Debug("Fetching status")
		client := ipc.NewClient()
		response, err := client.Send(ipc.Command{Action: "status"})
		if err != nil {
			return fmt.Errorf("failed to send status command: %w", err)
		}

		if !response.Success {
			if response.Code != "" {
				return fmt.Errorf("%s (code: %s)", response.Message, response.Code)
			}
			return fmt.Errorf("%s", response.Message)
		}

		logger.Info("Neru Status:")
		var sd ipc.StatusData
		raw, err := json.Marshal(response.Data)
		if err == nil {
			err2 := json.Unmarshal(raw, &sd)
			if err2 == nil {
				status := "stopped"
				if sd.Enabled {
					status = "running"
				}
				logger.Info("  Status: " + status)
				logger.Info("  Mode: " + sd.Mode)
				logger.Info("  Config: " + sd.Config)
			} else {
				// Fallback to previous behavior
				if data, ok := response.Data.(map[string]any); ok {
					if enabled, ok := data["enabled"].(bool); ok {
						status := "stopped"
						if enabled {
							status = "running"
						}
						logger.Info("  Status: " + status)
					}
					if mode, ok := data["mode"].(string); ok {
						logger.Info("  Mode: " + mode)
					}
					if configPath, ok := data["config"].(string); ok {
						logger.Info("  Config: " + configPath)
					}
				} else {
					jsonData, err := json.MarshalIndent(response.Data, "  ", "  ")
					if err != nil {
						logger.Error("Failed to marshal status data to JSON", zap.Error(err))
						return fmt.Errorf("failed to marshal status data: %w", err)
					}
					logger.Info(string(jsonData))
				}
			}
		} else {
			jsonData, err := json.MarshalIndent(response.Data, "  ", "  ")
			if err != nil {
				logger.Error("Failed to marshal status data to JSON", zap.Error(err))
				return fmt.Errorf("failed to marshal status data: %w", err)
			}
			logger.Info(string(jsonData))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
