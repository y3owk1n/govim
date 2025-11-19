package coordinates

import (
	"image"
	"math"
)

// ComputeRestoredPosition calculates the restored cursor position when switching screens.
// It maintains the relative position of the cursor within the screen bounds.
func ComputeRestoredPosition(initPos image.Point, from, to image.Rectangle) image.Point {
	// If screens are the same, no adjustment needed
	if from == to {
		return initPos
	}

	// Validate screen bounds
	if from.Dx() == 0 || from.Dy() == 0 || to.Dx() == 0 || to.Dy() == 0 {
		return initPos
	}

	// Calculate relative position (0.0 to 1.0) within the original screen
	rx := float64(initPos.X-from.Min.X) / float64(from.Dx())
	ry := float64(initPos.Y-from.Min.Y) / float64(from.Dy())

	// Clamp relative positions to valid range
	rx = ClampFloat(rx, 0, 1)
	ry = ClampFloat(ry, 0, 1)

	// Calculate new position in target screen
	nx := to.Min.X + int(math.Round(rx*float64(to.Dx())))
	ny := to.Min.Y + int(math.Round(ry*float64(to.Dy())))

	// Clamp to target screen bounds
	nx = ClampInt(nx, to.Min.X, to.Max.X)
	ny = ClampInt(ny, to.Min.Y, to.Max.Y)

	return image.Point{X: nx, Y: ny}
}

// NormalizeToLocalCoordinates converts screen-absolute coordinates to window-local coordinates.
// The overlay window is positioned at the screen origin, but the view uses local coordinates.
func NormalizeToLocalCoordinates(screenBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, screenBounds.Dx(), screenBounds.Dy())
}

// ConvertToAbsoluteCoordinates converts window-local coordinates to screen-absolute coordinates.
func ConvertToAbsoluteCoordinates(
	localPoint image.Point,
	screenBounds image.Rectangle,
) image.Point {
	return image.Point{
		X: localPoint.X + screenBounds.Min.X,
		Y: localPoint.Y + screenBounds.Min.Y,
	}
}

// ClampFloat clamps a float64 value between minVal and maxVal.
func ClampFloat(value, minVal, maxVal float64) float64 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// ClampInt clamps an int value between minVal and maxVal.
func ClampInt(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}
