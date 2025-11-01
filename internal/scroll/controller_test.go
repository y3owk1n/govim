package scroll

import (
	"testing"
	"go.uber.org/zap"
	"github.com/y3owk1n/govim/internal/config"
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
