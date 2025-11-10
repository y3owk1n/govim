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
	mHintsLeftClick := mHints.AddSubMenuItem("Left Click", "Show left click hints")
	mHintsRightClick := mHints.AddSubMenuItem("Right Click", "Show right click hints")
	mHintsDoubleClick := mHints.AddSubMenuItem("Double Click", "Show double click hints")
	mHintsTripleClick := mHints.AddSubMenuItem("Triple Click", "Show triple click hints")
	mHintsMouseUp := mHints.AddSubMenuItem("Mouse Up", "Show mouse up hints")
	mHintsMouseDown := mHints.AddSubMenuItem("Mouse Down", "Show mouse down hints")
	mHintsMiddleClick := mHints.AddSubMenuItem("Middle Click", "Show middle click hints")
	mHintsMoveMouse := mHints.AddSubMenuItem("Move Mouse", "Show move mouse hints")
	mHintsScroll := mHints.AddSubMenuItem("Scroll", "Show scroll hints")
	mHintsContextMenu := mHints.AddSubMenuItem("Context Menu", "Show context menu hints")
	
	// Grid submenu
	mGrid := systray.AddMenuItem("Grid", "Grid mode actions")
	mGridLeftClick := mGrid.AddSubMenuItem("Left Click", "Grid left click")
	mGridRightClick := mGrid.AddSubMenuItem("Right Click", "Grid right click")
	mGridDoubleClick := mGrid.AddSubMenuItem("Double Click", "Grid double click")
	mGridTripleClick := mGrid.AddSubMenuItem("Triple Click", "Grid triple click")
	mGridMouseUp := mGrid.AddSubMenuItem("Mouse Up", "Grid mouse up")
	mGridMouseDown := mGrid.AddSubMenuItem("Mouse Down", "Grid mouse down")
	mGridMiddleClick := mGrid.AddSubMenuItem("Middle Click", "Grid middle click")
	mGridMoveMouse := mGrid.AddSubMenuItem("Move Mouse", "Grid move mouse")
	mGridScroll := mGrid.AddSubMenuItem("Scroll", "Grid scroll")
	mGridContextMenu := mGrid.AddSubMenuItem("Context Menu", "Grid context menu")

	// Quit option
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Neru", "Exit the application")

	// Handle clicks in a separate goroutine
	go handleSystrayEvents(
		mVersionCopy, mStatus, mToggle,
		mHintsLeftClick, mHintsRightClick, mHintsDoubleClick, mHintsTripleClick,
		mHintsMouseUp, mHintsMouseDown, mHintsMiddleClick, mHintsMoveMouse,
		mHintsScroll, mHintsContextMenu,
		mGridLeftClick, mGridRightClick, mGridDoubleClick, mGridTripleClick,
		mGridMouseUp, mGridMouseDown, mGridMiddleClick, mGridMoveMouse,
		mGridScroll, mGridContextMenu,
		mQuit,
	)
}

func handleSystrayEvents(
	mVersionCopy, mStatus, mToggle *systray.MenuItem,
	mHintsLeftClick, mHintsRightClick, mHintsDoubleClick, mHintsTripleClick *systray.MenuItem,
	mHintsMouseUp, mHintsMouseDown, mHintsMiddleClick, mHintsMoveMouse *systray.MenuItem,
	mHintsScroll, mHintsContextMenu *systray.MenuItem,
	mGridLeftClick, mGridRightClick, mGridDoubleClick, mGridTripleClick *systray.MenuItem,
	mGridMouseUp, mGridMouseDown, mGridMiddleClick, mGridMoveMouse *systray.MenuItem,
	mGridScroll, mGridContextMenu *systray.MenuItem,
	mQuit *systray.MenuItem,
) {
	for {
		select {
		case <-mVersionCopy.ClickedCh:
			handleVersionCopy()
		case <-mToggle.ClickedCh:
			handleToggleEnable(mStatus, mToggle)
		// Hints mode actions
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
		// Grid mode actions
		case <-mGridLeftClick.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionLeftClick)
		case <-mGridRightClick.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionRightClick)
		case <-mGridDoubleClick.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionDoubleClick)
		case <-mGridTripleClick.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionTripleClick)
		case <-mGridMouseUp.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionMouseUp)
		case <-mGridMouseDown.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionMouseDown)
		case <-mGridMiddleClick.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionMiddleClick)
		case <-mGridMoveMouse.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionMoveMouse)
		case <-mGridScroll.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionScroll)
		case <-mGridContextMenu.ClickedCh:
			activateModeFromSystray(ModeGrid, ActionContextMenu)
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
