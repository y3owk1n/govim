package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/ipc"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show neru status",
	Long:  `Display the current status of the neru program.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return requiresRunningInstance()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ipc.NewClient()
		response, err := client.Send(ipc.Command{Action: "status"})
		if err != nil {
			return err
		}

		if !response.Success {
			return fmt.Errorf("%s", response.Message)
		}

		// Pretty print the status data
		fmt.Println("Neru Status:")
		if data, ok := response.Data.(map[string]any); ok {
			if enabled, ok := data["enabled"].(bool); ok {
				status := "stopped"
				if enabled {
					status = "running"
				}
				fmt.Printf("  Status: %s\n", status)
			}
			if mode, ok := data["mode"].(string); ok {
				fmt.Printf("  Mode: %s\n", mode)
			}
			if configPath, ok := data["config"].(string); ok {
				fmt.Printf("  Config: %s\n", configPath)
			}
		} else {
			// Fallback to JSON output if structure is unexpected
			jsonData, _ := json.MarshalIndent(response.Data, "  ", "  ")
			fmt.Println(string(jsonData))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
