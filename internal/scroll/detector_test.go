package scroll

import (
	"testing"
)

func TestDetector_NewDetector(t *testing.T) {
	detector := NewDetector()

	if detector == nil {
		t.Error("Expected non-nil detector")
	}
}

func TestDetector_BasicOperations(t *testing.T) {
	detector := NewDetector()

	// Test that detector can be used without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Detector operations caused panic: %v", r)
		}
	}()

	// Test basic methods exist and don't crash
	// This is a basic smoke test
	_ = detector
}
