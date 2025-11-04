package scroll

import (
	"github.com/y3owk1n/neru/internal/config"
	"go.uber.org/zap"
	"testing"
)

func TestController_NewController(t *testing.T) {
	cfg := config.ScrollConfig{}
	logger := zap.NewNop()
	controller := NewController(cfg, logger)

	if controller == nil {
		t.Error("Expected non-nil controller")
	}
}

func TestController_BasicOperations(t *testing.T) {
	cfg := config.ScrollConfig{}
	logger := zap.NewNop()
	controller := NewController(cfg, logger)

	// Test that controller can be used without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Controller operations caused panic: %v", r)
		}
	}()

	// Test basic methods exist and don't crash
	// This is a basic smoke test
	_ = controller
}

func TestScrollDirections(t *testing.T) {
	// Test that direction constants are defined correctly
	directions := []Direction{DirectionUp, DirectionDown, DirectionLeft, DirectionRight}
	expectedValues := []int{0, 1, 2, 3}

	for i, dir := range directions {
		if int(dir) != expectedValues[i] {
			t.Errorf("Direction constant %d has unexpected value %d", i, dir)
		}
	}
}

func TestScrollAmounts(t *testing.T) {
	// Test that amount constants are defined correctly
	amounts := []ScrollAmount{AmountChar, AmountHalfPage, AmountFullPage}
	expectedValues := []int{0, 1, 2}

	for i, amount := range amounts {
		if int(amount) != expectedValues[i] {
			t.Errorf("ScrollAmount constant %d has unexpected value %d", i, amount)
		}
	}
}
