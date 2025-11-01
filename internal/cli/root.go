package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/govim/internal/ipc"
)

var (
	configPath string
	// LaunchFunc is set by main to handle daemon launch
	LaunchFunc func(configPath string)

	// Version information (set via ldflags at build time)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "govim",
	Short: "GoVim - Keyboard-driven navigation for macOS",
	Long: `GoVim is a keyboard-driven navigation tool for macOS that provides
vim-like navigation capabilities across all applications using accessibility APIs.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")

	// Customize version output
	rootCmd.SetVersionTemplate(fmt.Sprintf("GoVim version %s\nGit commit: %s\nBuild date: %s\n", Version, GitCommit, BuildDate))
}

// launchProgram launches the main govim program
func launchProgram(cfgPath string) {
	// Check if already running
	if ipc.IsServerRunning() {
		fmt.Println("GoVim is already running")
		os.Exit(0)
	}

	// Call the launch function set by main
	if LaunchFunc != nil {
		LaunchFunc(cfgPath)
	} else {
		fmt.Fprintln(os.Stderr, "Error: Launch function not initialized")
		os.Exit(1)
	}
}

// sendCommand sends a command to the running govim instance
func sendCommand(action string) error {
	if !ipc.IsServerRunning() {
		return fmt.Errorf("govim is not running. Start it first with 'govim' or 'govim launch'")
	}

	client := ipc.NewClient()
	response, err := client.Send(ipc.Command{Action: action})
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf(response.Message)
	}

	fmt.Println(response.Message)
	return nil
}

// requiresRunningInstance checks if govim is running and exits with error if not
func requiresRunningInstance() error {
	if !ipc.IsServerRunning() {
		fmt.Fprintln(os.Stderr, "Error: govim is not running")
		fmt.Fprintln(os.Stderr, "Start it first with: govim launch")
		os.Exit(1)
	}
	return nil
}
