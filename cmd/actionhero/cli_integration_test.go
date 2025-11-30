package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const binaryName = "actionhero-test"

// TestMain builds the binary before running tests and cleans up after
func TestMain(m *testing.M) {
	// Build the binary for testing
	build := exec.Command("go", "build", "-o", binaryName)
	if err := build.Run(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(binaryName)
	os.Exit(code)
}

// runCLI executes the CLI with the given arguments and returns stdout, stderr, and exit code
func runCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command("./"+binaryName, args...)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v", err)
		}
	}

	return stdout, stderr, exitCode
}

func TestCLI_Help(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Verify help output contains expected content
	expectedStrings := []string{
		"Go ActionHero",
		"status",
		"echo",
		"user:create",
		"Available Commands:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help output to contain %q, but it didn't.\nOutput: %s", expected, stdout)
		}
	}
}

func TestCLI_StatusAction(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "status", "--quiet")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v\nOutput: %s", err, stdout)
	}

	// Verify response structure
	if response["response"] == nil {
		t.Error("Expected 'response' field in output")
	}

	respData, ok := response["response"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'response' to be an object")
	}

	if respData["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", respData["status"])
	}

	if respData["timestamp"] == nil {
		t.Error("Expected 'timestamp' field in response")
	}

	if respData["uptime"] != "running" {
		t.Errorf("Expected uptime 'running', got %v", respData["uptime"])
	}
}

func TestCLI_EchoAction(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "echo", "--message", "Hello from CLI!", "--quiet")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v\nOutput: %s", err, stdout)
	}

	// Verify response structure
	respData, ok := response["response"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'response' to be an object")
	}

	received, ok := respData["received"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'received' to be an object")
	}

	if received["message"] != "Hello from CLI!" {
		t.Errorf("Expected message 'Hello from CLI!', got %v", received["message"])
	}
}

func TestCLI_UserCreateAction(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t,
		"user:create",
		"--name", "Test User",
		"--email", "test@example.com",
		"--password", "testpass123",
		"--quiet",
	)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d\nStderr: %s", exitCode, stderr)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Parse JSON response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v\nOutput: %s", err, stdout)
	}

	// Verify response structure
	respData, ok := response["response"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'response' to be an object")
	}

	if respData["created"] != true {
		t.Errorf("Expected created to be true, got %v", respData["created"])
	}

	if respData["userId"] == nil {
		t.Error("Expected 'userId' field in response")
	}

	if respData["name"] != "Test User" {
		t.Errorf("Expected name 'Test User', got %v", respData["name"])
	}

	if respData["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %v", respData["email"])
	}
}

func TestCLI_ActionHelp(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "user:create", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Verify help output contains expected flags
	expectedStrings := []string{
		"--name string",
		"--email string",
		"--password string",
		"(required)",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help output to contain %q, but it didn't.\nOutput: %s", expected, stdout)
		}
	}
}

func TestCLI_MissingRequiredParameter(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t,
		"user:create",
		"--name", "Test",
		"--quiet",
	)

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for missing required parameters")
	}

	if stdout != "" {
		t.Errorf("Expected empty stdout for error, got: %s", stdout)
	}

	// Verify error message mentions the missing required flags
	if !strings.Contains(stderr, "required flag") {
		t.Errorf("Expected error message about required flags, got: %s", stderr)
	}

	if !strings.Contains(stderr, "email") || !strings.Contains(stderr, "password") {
		t.Errorf("Expected error to mention 'email' and 'password', got: %s", stderr)
	}
}

// Note: Validation tests are skipped because the user:create action is a stub
// that doesn't perform actual validation. In a real implementation, you would
// add tests here for:
// - Invalid email format
// - Password too short
// - Missing required fields
// - Duplicate user detection

func TestCLI_QuietMode(t *testing.T) {
	// Test that quiet mode suppresses logging
	stdout, _, exitCode := runCLI(t, "status", "--quiet")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// In quiet mode, only JSON output should be present
	// No log messages should appear

	// Should be valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Errorf("Expected valid JSON in quiet mode, got error: %v\nOutput: %s", err, stdout)
	}

	// Should not contain log messages
	if strings.Contains(stdout, "INFO") || strings.Contains(stdout, "Initializing") {
		t.Errorf("Expected no log messages in quiet mode, got: %s", stdout)
	}
}

func TestCLI_NoColorFlag(t *testing.T) {
	// Test that --no-color flag works
	stdout, stderr, exitCode := runCLI(t, "status", "--quiet", "--no-color")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}

	// Should be valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v\nOutput: %s", err, stdout)
	}
}
