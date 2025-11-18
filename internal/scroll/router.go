package scroll

import "go.uber.org/zap"

// ParseKey interprets a key press and previous key state to determine the appropriate scroll operation.
// Returns the operation type, updated last key state, and a flag indicating if the operation should execute.
func ParseKey(key string, lastKey string, logger *zap.Logger) (string, string, bool) {
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
