package scroll

import "go.uber.org/zap"

// ParseKey maps a key press and last key state to a scroll operation.
// Returns op (one of: down, up, left, right, half_down, half_up, top, bottom, tab),
// updated lastKey, and ok indicating whether an action should be executed now.
func ParseKey(key string, lastKey string, logger *zap.Logger) (op string, newLastKey string, ok bool) {
	logger.Debug("Scroll router processing key",
		zap.String("key", key),
		zap.String("last_key", lastKey))

	// Control characters
	if len(key) == 1 {
		byteVal := key[0]
		switch byteVal {
		case 4: // Ctrl+D
			logger.Debug("Scroll router: Ctrl+D pressed (half page down)")
			return "half_down", "", true
		case 21: // Ctrl+U
			logger.Debug("Scroll router: Ctrl+U pressed (half page up)")
			return "half_up", "", true
		}
	}

	switch key {
	case "j":
		logger.Debug("Scroll router: j pressed (scroll down)")
		return "down", "", true
	case "k":
		logger.Debug("Scroll router: k pressed (scroll up)")
		return "up", "", true
	case "h":
		logger.Debug("Scroll router: h pressed (scroll left)")
		return "left", "", true
	case "l":
		logger.Debug("Scroll router: l pressed (scroll right)")
		return "right", "", true
	case "g":
		if lastKey == "g" {
			logger.Debug("Scroll router: gg pressed (scroll to top)")
			return "top", "", true
		}
		logger.Debug("Scroll router: g pressed (first g)")
		return "", "g", false
	case "G":
		logger.Debug("Scroll router: G pressed (scroll to bottom)")
		return "bottom", "", true
	case "\t":
		logger.Debug("Scroll router: Tab pressed")
		return "tab", "", true
	default:
		logger.Debug("Scroll router: Unrecognized key")
		return "", "", false
	}
}
