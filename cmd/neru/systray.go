package main

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/getlantern/systray"
	"github.com/y3owk1n/neru/internal/cli"
	"github.com/y3owk1n/neru/internal/logger"
	"go.uber.org/zap"
)

func onReady() {
	systray.SetTitle("⌨️")
	systray.SetTooltip("Neru - Keyboard Navigation")

	// Status submenu for version
	mVersion := systray.AddMenuItem(fmt.Sprintf("Version %s", cli.Version), "Show version")
	mVersion.Disable()
	mVersionCopy := systray.AddMenuItem("Copy version", "Copy version to clipboard")

	// Status toggle
	systray.AddSeparator()
	mStatus := systray.AddMenuItem("Status: Running", "Show current status")
	mStatus.Disable()
	mToggle := systray.AddMenuItem("Disable", "Disable/Enable Neru without quitting")

	// Control actions
	systray.AddSeparator()

	// Hints submenu
	mHints := systray.AddMenuItem("Hints", "Hint mode actions")
	if globalApp != nil && globalApp.config != nil && !globalApp.config.Hints.Enabled {
		mHints.Hide()
	}

	// Grid submenu
	mGrid := systray.AddMenuItem("Grid", "Grid mode actions")
	if globalApp != nil && globalApp.config != nil && !globalApp.config.Grid.Enabled {
		mGrid.Hide()
	}

	// Quit option
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Neru", "Exit the application")

	// Handle clicks in a separate goroutine
	go handleSystrayEvents(
		mVersionCopy, mStatus, mToggle,
		mHints, mGrid,
		mQuit,
	)
}

func handleSystrayEvents(
	mVersionCopy, mStatus, mToggle *systray.MenuItem,
	mHints, mGrid *systray.MenuItem,
	mQuit *systray.MenuItem,
) {
	for {
		select {
		case <-mVersionCopy.ClickedCh:
			handleVersionCopy()
		case <-mToggle.ClickedCh:
			handleToggleEnable(mStatus, mToggle)
		case <-mHints.ClickedCh:
			activateModeFromSystray(ModeHints)
		case <-mGrid.ClickedCh:
			activateModeFromSystray(ModeGrid)
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func handleVersionCopy() {
	err := clipboard.WriteAll(cli.Version)
	if err != nil {
		logger.Error("Error copying version to clipboard", zap.Error(err))
	}
}

func handleToggleEnable(mStatus, mToggle *systray.MenuItem) {
	if globalApp == nil {
		return
	}

	if globalApp.enabled {
		globalApp.enabled = false
		mStatus.SetTitle("Status: Disabled")
		mToggle.SetTitle("Enable")
	} else {
		globalApp.enabled = true
		mStatus.SetTitle("Status: Enabled")
		mToggle.SetTitle("Disable")
	}
}

func activateModeFromSystray(mode Mode) {
	if globalApp != nil {
		globalApp.activateMode(mode)
	}
}

func onExit() {
	if globalApp != nil {
		globalApp.Cleanup()
	}
}
