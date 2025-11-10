package scroll

// ParseKey maps a key press and last key state to a scroll operation.
// Returns op (one of: down, up, left, right, half_down, half_up, top, bottom, tab),
// updated lastKey, and ok indicating whether an action should be executed now.
func ParseKey(key string, lastKey string) (op string, newLastKey string, ok bool) {
	// Control characters
	if len(key) == 1 {
		byteVal := key[0]
		switch byteVal {
		case 4: // Ctrl+D
			return "half_down", "", true
		case 21: // Ctrl+U
			return "half_up", "", true
		}
	}

	switch key {
	case "j":
		return "down", "", true
	case "k":
		return "up", "", true
	case "h":
		return "left", "", true
	case "l":
		return "right", "", true
	case "g":
		if lastKey == "g" {
			return "top", "", true
		}
		return "", "g", false
	case "G":
		return "bottom", "", true
	case "\t":
		return "tab", "", true
	default:
		return "", "", false
	}
}
