package main

import "go.uber.org/zap"

// logModeActivation logs mode-specific activation messages
func (a *App) logModeActivation(mode Mode) {
	hintCount := len(a.hintManager.GetHints())

	switch mode {
	case ModeHintWithActions:
		a.logger.Info("Hint mode with actions activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label, then choose action: l=left, r=right, d=double, m=middle")
	case ModeHint:
		a.logger.Info("Hint mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Type a hint label to click the element")
	case ModeScroll:
		a.logger.Info("Scroll mode activated", zap.Int("hints", hintCount))
		a.logger.Info("Use j/k to scroll, Ctrl+D/U for half-page, g/G for top/bottom")
	}
}
