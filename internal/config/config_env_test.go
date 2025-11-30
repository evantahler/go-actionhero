package config

import (
	"os"
	"testing"
)

func TestLoad_EnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `ACTIONHERO_PROCESS_NAME=test-from-env
ACTIONHERO_LOGGER_LEVEL=debug
ACTIONHERO_DATABASE_HOST=db.from.env
ACTIONHERO_SERVER_WEB_PORT=9999
`

	envFile := ".env.test"
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}
	defer func() { _ = os.Remove(envFile) }()

	// Clear environment first
	os.Clearenv()

	// Create .env file
	envFile2 := ".env"
	err = os.WriteFile(envFile2, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(envFile2) }()

	// Clear environment and reload
	os.Clearenv()
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify values from .env are loaded
	if cfg2.Process.Name != "test-from-env" {
		t.Errorf("Expected process name 'test-from-env', got %v", cfg2.Process.Name)
	}
	if cfg2.Logger.Level != "debug" {
		t.Errorf("Expected logger level 'debug', got %v", cfg2.Logger.Level)
	}
	if cfg2.Database.Host != "db.from.env" {
		t.Errorf("Expected database host 'db.from.env', got %v", cfg2.Database.Host)
	}
	if cfg2.Server.Web.Port != 9999 {
		t.Errorf("Expected web server port 9999, got %v", cfg2.Server.Web.Port)
	}
}

func TestLoad_EnvFilePriority(t *testing.T) {
	// Create .env file
	envContent := `ACTIONHERO_PROCESS_NAME=from-env-file
ACTIONHERO_LOGGER_LEVEL=info
`
	envFile := ".env"
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer func() { _ = os.Remove(envFile) }()

	// Set environment variable (should override .env file)
	os.Clearenv()
	_ = os.Setenv("ACTIONHERO_PROCESS_NAME", "from-env-var")
	defer func() { _ = os.Unsetenv("ACTIONHERO_PROCESS_NAME") }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Environment variable should take precedence
	if cfg.Process.Name != "from-env-var" {
		t.Errorf("Expected process name 'from-env-var' (from env var), got %v", cfg.Process.Name)
	}
}

func TestLoad_EnvFileNotFound(t *testing.T) {
	// Clear environment and ensure .env doesn't exist
	os.Clearenv()

	// Remove .env if it exists
	_ = os.Remove(".env")
	_ = os.Remove(".env.local")
	_ = os.Remove(".env.test")
	_ = os.Remove(".env.dev")

	// Should still load successfully with defaults
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error when .env file doesn't exist, got %v", err)
	}

	// Should have default values
	if cfg.Process.Name != "actionhero" {
		t.Errorf("Expected default process name 'actionhero', got %v", cfg.Process.Name)
	}
}
