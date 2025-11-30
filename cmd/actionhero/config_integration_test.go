package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigCommand_Integration tests the config command end-to-end
func TestConfigCommand_Integration(t *testing.T) {
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	// Go up to project root (from cmd/actionhero to project root)
	projectRoot := filepath.Join(wd, "..", "..")

	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "actionhero")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/actionhero")
	cmd.Dir = projectRoot
	var buildStderr bytes.Buffer
	cmd.Stderr = &buildStderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v\nStderr: %s", err, buildStderr.String())
	}

	tests := []struct {
		name            string
		args            []string
		env             map[string]string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "config command with defaults",
			args: []string{"config"},
			wantContains: []string{
				"PROCESS",
				"LOGGER",
				"DATABASE",
				"REDIS",
				"SESSION",
				"SERVER - WEB",
				"TASKS",
				"Name: actionhero",
				"Level: info",
				"Host: localhost",
				"Port: 8080",
			},
		},
		{
			name: "config command with env overrides",
			args: []string{"config"},
			env: map[string]string{
				"ACTIONHERO_PROCESS_NAME":      "test-integration",
				"ACTIONHERO_LOGGER_LEVEL":      "debug",
				"ACTIONHERO_SERVER_WEB_PORT":   "9999",
				"ACTIONHERO_DATABASE_PASSWORD": "secret123",
			},
			wantContains: []string{
				"Name: test-integration",
				"Level: debug",
				"Port: 9999",
				"Password: *********", // Should be masked
			},
			wantNotContains: []string{
				"secret123", // Password should not appear in plain text
			},
		},
		{
			name: "config command with colorize disabled",
			args: []string{"config"},
			env: map[string]string{
				"ACTIONHERO_LOGGER_COLORIZE": "false",
			},
			wantContains: []string{
				"PROCESS",
				"Name: actionhero",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			originalEnv := make(map[string]string)
			for key, value := range tt.env {
				originalEnv[key] = os.Getenv(key)
				_ = os.Setenv(key, value)
			}
			defer func() {
				// Restore original environment
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						_ = os.Unsetenv(key)
					} else {
						_ = os.Setenv(key, originalValue)
					}
				}
			}()

			// Run the command
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			}

			output := stdout.String() + stderr.String()

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output should contain %q, but it doesn't.\nOutput:\n%s", want, output)
				}
			}

			// Check for content that should NOT be present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("Output should NOT contain %q, but it does.\nOutput:\n%s", notWant, output)
				}
			}
		})
	}
}

// TestConfigCommand_ErrorHandling tests error cases
func TestConfigCommand_ErrorHandling(t *testing.T) {
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")

	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "actionhero")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/actionhero")
	cmd.Dir = projectRoot
	var buildStderr bytes.Buffer
	cmd.Stderr = &buildStderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v\nStderr: %s", err, buildStderr.String())
	}

	// Test unknown command
	testCmd := exec.Command(binaryPath, "unknown-command")
	var stdout, stderr bytes.Buffer
	testCmd.Stdout = &stdout
	testCmd.Stderr = &stderr
	err = testCmd.Run()
	if err == nil {
		t.Error("Expected error for unknown command, but got none")
	}
	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "unknown command") {
		t.Errorf("Expected 'unknown command' in output, got stdout: %s, stderr: %s", stdout.String(), stderr.String())
	}
}

// TestConfigCommand_Help tests the help command
func TestConfigCommand_Help(t *testing.T) {
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(wd, "..", "..")

	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "actionhero")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/actionhero")
	cmd.Dir = projectRoot
	var buildStderr bytes.Buffer
	cmd.Stderr = &buildStderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v\nStderr: %s", err, buildStderr.String())
	}

	// Test help command
	testCmd := exec.Command(binaryPath, "help")
	var stdout bytes.Buffer
	testCmd.Stdout = &stdout
	err = testCmd.Run()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := stdout.String()
	wantContains := []string{
		"Usage:",
		"start",
		"config",
		"help",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("Help output should contain %q, but it doesn't.\nOutput:\n%s", want, output)
		}
	}
}
