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
	mHintsLeftClick := systray.AddMenuItem("Left Click", "Show left click hints")
	mHintsRightClick := systray.AddMenuItem("Right Click", "Show right click hints")
	mHintsDoubleClick := systray.AddMenuItem("Double Click", "Show double click hints")
	mHintsTripleClick := systray.AddMenuItem("Triple Click", "Show triple click hints")
	mHintsMouseUp := systray.AddMenuItem("Mouse Up", "Show mouse up hints")
	mHintsMouseDown := systray.AddMenuItem("Mouse Down", "Show mouse down hints")
	mHintsMiddleClick := systray.AddMenuItem("Middle Click", "Show middle click hints")
	mHintsMoveMouse := systray.AddMenuItem("Move Mouse", "Show move mouse hints")
	mHintsScroll := systray.AddMenuItem("Scroll", "Show scroll hints")
	mHintsContextMenu := systray.AddMenuItem("Context Menu", "Show context menu hints")

	// Quit option
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Neru", "Exit the application")

	// Handle clicks in a separate goroutine
	go handleSystrayEvents(mVersionCopy, mStatus, mToggle, mHintsLeftClick, mHintsRightClick, mHintsDoubleClick, mHintsTripleClick, mHintsMouseUp, mHintsMouseDown, mHintsMiddleClick, mHintsMoveMouse, mHintsScroll, mHintsContextMenu, mQuit)
}

func handleSystrayEvents(
	mVersionCopy, mStatus, mToggle, mHintsLeftClick, mHintsRightClick, mHintsDoubleClick, mHintsTripleClick, mHintsMouseUp, mHintsMouseDown, mHintsMiddleClick, mHintsMoveMouse, mHintsScroll, mHintsContextMenu, mQuit *systray.MenuItem,
) {
	for {
		select {
		case <-mVersionCopy.ClickedCh:
			handleVersionCopy()
		case <-mToggle.ClickedCh:
			handleToggleEnable(mStatus, mToggle)
		case <-mHintsLeftClick.ClickedCh:
			activateModeFromSystray(ModeHints, ActionLeftClick)
		case <-mHintsRightClick.ClickedCh:
			activateModeFromSystray(ModeHints, ActionRightClick)
		case <-mHintsDoubleClick.ClickedCh:
			activateModeFromSystray(ModeHints, ActionDoubleClick)
		case <-mHintsTripleClick.ClickedCh:
			activateModeFromSystray(ModeHints, ActionTripleClick)
		case <-mHintsMouseUp.ClickedCh:
			activateModeFromSystray(ModeHints, ActionMouseUp)
		case <-mHintsMouseDown.ClickedCh:
			activateModeFromSystray(ModeHints, ActionMouseDown)
		case <-mHintsMiddleClick.ClickedCh:
			activateModeFromSystray(ModeHints, ActionMiddleClick)
		case <-mHintsMoveMouse.ClickedCh:
			activateModeFromSystray(ModeHints, ActionMoveMouse)
		case <-mHintsScroll.ClickedCh:
			activateModeFromSystray(ModeHints, ActionScroll)
		case <-mHintsContextMenu.ClickedCh:
			activateModeFromSystray(ModeHints, ActionContextMenu)
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

func activateModeFromSystray(mode Mode, action Action) {
	if globalApp != nil {
		globalApp.activateMode(mode, action)
	}
}

func onExit() {
	if globalApp != nil {
		globalApp.Cleanup()
	}
}
