package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/fatih/color"
)

func TestDumpConfig(t *testing.T) {
	// Disable colors for testing
	color.NoColor = true

	// Load default config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger with buffer output
	var buf bytes.Buffer
	logger := util.NewLogger(cfg.Logger)
	logger.SetOutput(&buf)

	// Dump config
	dumpConfig(cfg, logger, "list")

	output := buf.String()

	// Check that all sections are present
	sections := []string{
		"PROCESS",
		"LOGGER",
		"DATABASE",
		"REDIS",
		"SESSION",
		"SERVER - WEB",
		"TASKS",
	}

	for _, section := range sections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section '%s' not found in output", section)
		}
	}

	// Check that key values are present
	keys := []string{
		"Name:",
		"Level:",
		"Host:",
		"Port:",
		"Cookie Name:",
		"Enabled:",
	}

	for _, key := range keys {
		if !strings.Contains(output, key) {
			t.Errorf("Expected key '%s' not found in output", key)
		}
	}
}

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{
			name:     "empty password",
			password: "",
			want:     "(empty)",
		},
		{
			name:     "short password",
			password: "abc",
			want:     "***",
		},
		{
			name:     "long password",
			password: "verylongpassword123",
			want:     strings.Repeat("*", len("verylongpassword123")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskPassword(tt.password)
			if got != tt.want {
				t.Errorf("maskPassword(%q) = %q, want %q", tt.password, got, tt.want)
			}
		})
	}
}

func TestDumpConfig_WithEnvOverrides(t *testing.T) {
	// Disable colors for testing
	color.NoColor = true

	// Set environment variables
	os.Setenv("ACTIONHERO_PROCESS_NAME", "test-app")
	os.Setenv("ACTIONHERO_LOGGER_LEVEL", "debug")
	os.Setenv("ACTIONHERO_SERVER_WEB_PORT", "9000")
	defer func() {
		os.Unsetenv("ACTIONHERO_PROCESS_NAME")
		os.Unsetenv("ACTIONHERO_LOGGER_LEVEL")
		os.Unsetenv("ACTIONHERO_SERVER_WEB_PORT")
	}()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create logger with buffer output
	var buf bytes.Buffer
	logger := util.NewLogger(cfg.Logger)
	logger.SetOutput(&buf)

	// Dump config
	dumpConfig(cfg, logger, "list")

	output := buf.String()

	// Check that overridden values appear
	if !strings.Contains(output, "test-app") {
		t.Error("Expected 'test-app' (from env var) not found in output")
	}
	if !strings.Contains(output, "debug") {
		t.Error("Expected 'debug' (from env var) not found in output")
	}
	if !strings.Contains(output, "9000") {
		t.Error("Expected '9000' (from env var) not found in output")
	}
}

