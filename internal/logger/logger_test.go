package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestLoggerInitAndClose(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	// Init logger to write to our temp file
	if err := Init("debug", logPath, false); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// Use logger functions to exercise Get() and outputs
	Info("test info", zap.String("k", "v"))
	Debug("test debug")

	if Get() == nil {
		t.Fatalf("Get returned nil logger")
	}

	if err := Sync(); err != nil {
		// Some environments return sync errors for stdout/stderr (e.g. bad file descriptor).
		// Close() explicitly ignores Sync errors on stdout/stderr; here we log and continue.
		t.Logf("Sync returned non-fatal error (ignored): %v", err)
	}

	if err := Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// verify log file exists
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected log file at %s: %v", logPath, err)
	}
}
