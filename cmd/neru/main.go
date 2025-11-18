// Main package for neru.
package main

import (
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/app"
	"github.com/y3owk1n/neru/internal/cli"
	"github.com/y3owk1n/neru/internal/config"
)

var globalApp *app.App

func main() {
	// Set the launch function for CLI
	cli.LaunchFunc = LaunchDaemon

	// Let cobra handle all command parsing
	cli.Execute()
}

// LaunchDaemon is called by the CLI to launch the daemon.
func LaunchDaemon(configPath string) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create app
	a, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}
	// Record the config path the daemon was started with so status can report
	// the same path (if provided via --config)
	a.ConfigPath = configPath
	globalApp = a

	// Start app in goroutine
	go func() {
		err := a.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		}
	}()

	// Run systray (blocks until quit)
	systray.Run(onReady, onExit)
}
