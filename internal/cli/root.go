package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/ipc"
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
	Use:   "neru",
	Short: "Neru - Keyboard-driven navigation for macOS",
	Long: `Neru is a keyboard-driven navigation tool for macOS that provides
vim-like navigation capabilities across all applications using accessibility APIs.`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Auto-launch when running from app bundle without arguments
		if isRunningFromAppBundle() && len(args) == 0 {
			fmt.Println("Launching Neru from app bundle...")
			launchProgram(configPath)
			return nil
		}
		return cmd.Help()
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
	rootCmd.SetVersionTemplate(fmt.Sprintf("Neru version %s\nGit commit: %s\nBuild date: %s\n", Version, GitCommit, BuildDate))
}

// isRunningFromAppBundle checks if the binary is running from a macOS app bundle
func isRunningFromAppBundle() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Check if running from .app/Contents/MacOS/
	return strings.Contains(realPath, ".app/Contents/MacOS")
}

// launchProgram launches the main neru program
func launchProgram(cfgPath string) {
	// Check if already running
	if ipc.IsServerRunning() {
		fmt.Println("Neru is already running")
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

// sendCommand sends a command to the running neru instance
func sendCommand(action string) error {
	if !ipc.IsServerRunning() {
		return fmt.Errorf("neru is not running. Start it first with 'neru' or 'neru launch'")
	}

	client := ipc.NewClient()
	response, err := client.Send(ipc.Command{Action: action})
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("%s", response.Message)
	}

	fmt.Println(response.Message)
	return nil
}

// requiresRunningInstance checks if neru is running and exits with error if not
func requiresRunningInstance() error {
	if !ipc.IsServerRunning() {
		fmt.Fprintln(os.Stderr, "Error: neru is not running")
		fmt.Fprintln(os.Stderr, "Start it first with: neru launch")
		os.Exit(1)
	}
	return nil
}
