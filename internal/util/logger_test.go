package util

import (
	"bytes"
	"strings"
	"testing"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	cfg := config.DefaultLoggerConfig()
	cfg.Level = "debug"
	logger := NewLogger(cfg)

	if logger == nil {
		t.Fatal("Expected logger to be created")
	}
	if logger.GetLevel() != logrus.DebugLevel {
		t.Errorf("Expected level %v, got %v", logrus.DebugLevel, logger.GetLevel())
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected logrus.Level
	}{
		{"debug", "debug", logrus.DebugLevel},
		{"info", "info", logrus.InfoLevel},
		{"warn", "warn", logrus.WarnLevel},
		{"error", "error", logrus.ErrorLevel},
		{"fatal", "fatal", logrus.FatalLevel},
		{"invalid", "invalid", logrus.InfoLevel}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggerConfig{Level: tt.level}
			logger := NewLogger(cfg)
			if logger.GetLevel() != tt.expected {
				t.Errorf("Expected level %v, got %v", tt.expected, logger.GetLevel())
			}
		})
	}
}

func TestLogger_LogMethods(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.DefaultLoggerConfig()
	cfg.Colorize = false
	cfg.Level = "debug"
	logger := NewLogger(cfg)
	logger.SetOutput(&buf)

	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message not found in output")
	}
	buf.Reset()

	logger.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info message not found in output")
	}
	buf.Reset()

	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn message not found in output")
	}
	buf.Reset()

	logger.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error message not found in output")
	}
	buf.Reset()

	logger.Debugf("formatted %s", "debug")
	if !strings.Contains(buf.String(), "formatted debug") {
		t.Error("Formatted debug message not found in output")
	}
	buf.Reset()

	logger.Infof("formatted %s", "info")
	if !strings.Contains(buf.String(), "formatted info") {
		t.Error("Formatted info message not found in output")
	}
	buf.Reset()

	logger.Warnf("formatted %s", "warn")
	if !strings.Contains(buf.String(), "formatted warn") {
		t.Error("Formatted warn message not found in output")
	}
	buf.Reset()

	logger.Errorf("formatted %s", "error")
	if !strings.Contains(buf.String(), "formatted error") {
		t.Error("Formatted error message not found in output")
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	cfg := config.DefaultLoggerConfig()
	cfg.Colorize = false
	logger := NewLogger(cfg)
	logger.SetOutput(&buf)

	logger.WithField("key", "value").Info("test")
	output := buf.String()
	if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
		t.Error("Expected fields to be included in output")
	}
}

func TestLogger_Colorize(t *testing.T) {
	cfg := config.DefaultLoggerConfig()
	cfg.Colorize = true
	logger := NewLogger(cfg)

	// Check that formatter is set correctly
	_, ok := logger.Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Error("Expected TextFormatter when colorize is true")
	}

	cfg.Colorize = false
	logger2 := NewLogger(cfg)
	_, ok = logger2.Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Error("Expected JSONFormatter when colorize is false")
	}
}

func TestLogger_Timestamp(t *testing.T) {
	cfg := config.DefaultLoggerConfig()
	cfg.Timestamp = true
	cfg.Colorize = false
	logger := NewLogger(cfg)

	formatter, ok := logger.Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Fatal("Expected JSONFormatter")
	}
	if formatter.DisableTimestamp {
		t.Error("Expected timestamps to be enabled")
	}

	cfg.Timestamp = false
	logger2 := NewLogger(cfg)
	formatter2, ok := logger2.Formatter.(*logrus.JSONFormatter)
	if !ok {
		t.Fatal("Expected JSONFormatter")
	}
	if !formatter2.DisableTimestamp {
		t.Error("Expected timestamps to be disabled")
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	cfg := config.DefaultLoggerConfig()
	if cfg.Level != "info" {
		t.Errorf("Expected default level 'info', got %v", cfg.Level)
	}
	if !cfg.Colorize {
		t.Error("Expected default colorize to be true")
	}
	if !cfg.Timestamp {
		t.Error("Expected default timestamp to be true")
	}
}

