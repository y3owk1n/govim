package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	logFile      *os.File
	logFileMu    sync.Mutex
)

// Init initializes the global logger
func Init(logLevel, logFilePath string, structured bool) error {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	// Close existing log file if any
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("failed to close existing log file: %w", err)
		}
		logFile = nil
	}

	// Determine log level
	level := zapcore.InfoLevel
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// Determine log file path
	if logFilePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		logFilePath = filepath.Join(homeDir, "Library", "Logs", "neru", "app.log")
	}

	// Create log directory
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if structured {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Create file writer
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Store file reference for cleanup
	logFile = file

	// Create console writer
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create core with both console and file output
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), level),
	)

	// Create logger
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Get returns the global logger
func Get() *zap.Logger {
	if globalLogger == nil {
		// Fallback to development logger
		globalLogger, _ = zap.NewDevelopment()
	}
	return globalLogger
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger != nil {
		if err := globalLogger.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the log file and syncs the logger
func Close() error {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	if globalLogger != nil {
		if err := globalLogger.Sync(); err != nil {
			// Ignore common sync errors that occur during shutdown
			if !strings.Contains(err.Error(), "invalid argument") &&
				!strings.Contains(err.Error(), "inappropriate ioctl for device") {
				return fmt.Errorf("failed to sync logger: %w", err)
			}
		}
		globalLogger = nil
	}

	if logFile != nil {
		// Best effort sync, ignore errors on stdout/stderr
		_ = logFile.Sync()
		if err := logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		logFile = nil
	}

	return nil
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// With creates a child logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}
