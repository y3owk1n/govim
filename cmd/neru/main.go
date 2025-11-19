// Package main provides the entry point for the Neru application.
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
	cli.LaunchFunc = LaunchDaemon
	cli.Execute()
}

// LaunchDaemon is called by the CLI to launch the daemon.
func LaunchDaemon(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	a, err := app.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	a.ConfigPath = configPath
	globalApp = a

	go func() {
		err := a.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		}
	}()

	systray.Run(onReady, onExit)
}
