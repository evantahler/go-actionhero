package config

import (
	"os"
	"testing"
)

const (
	defaultProcessName = "actionhero"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Process.Name != defaultProcessName {
		t.Errorf("Expected process name '%s', got %v", defaultProcessName, cfg.Process.Name)
	}

	if cfg.Logger.Level != "info" {
		t.Errorf("Expected logger level 'info', got %v", cfg.Logger.Level)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected database host 'localhost', got %v", cfg.Database.Host)
	}

	if cfg.Server.Web.Port != 8080 {
		t.Errorf("Expected web server port 8080, got %v", cfg.Server.Web.Port)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Clear environment
	os.Clearenv()

	// Set environment variables
	_ = os.Setenv("ACTIONHERO_PROCESS_NAME", "test-app")
	_ = os.Setenv("ACTIONHERO_LOGGER_LEVEL", "debug")
	_ = os.Setenv("ACTIONHERO_DATABASE_HOST", "db.example.com")
	_ = os.Setenv("ACTIONHERO_SERVER_WEB_PORT", "9000")
	defer func() {
		_ = os.Unsetenv("ACTIONHERO_PROCESS_NAME")
		_ = os.Unsetenv("ACTIONHERO_LOGGER_LEVEL")
		_ = os.Unsetenv("ACTIONHERO_DATABASE_HOST")
		_ = os.Unsetenv("ACTIONHERO_SERVER_WEB_PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cfg.Process.Name != "test-app" {
		t.Errorf("Expected process name 'test-app', got %v", cfg.Process.Name)
	}

	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected logger level 'debug', got %v", cfg.Logger.Level)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected database host 'db.example.com', got %v", cfg.Database.Host)
	}

	if cfg.Server.Web.Port != 9000 {
		t.Errorf("Expected web server port 9000, got %v", cfg.Server.Web.Port)
	}
}

func TestDefaultConfigs(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "ProcessConfig",
			test: func(t *testing.T) {
				cfg := DefaultProcessConfig()
				if cfg.Name != defaultProcessName {
					t.Errorf("Expected name '%s', got %v", defaultProcessName, cfg.Name)
				}
			},
		},
		{
			name: "LoggerConfig",
			test: func(t *testing.T) {
				cfg := DefaultLoggerConfig()
				if cfg.Level != "info" {
					t.Errorf("Expected level 'info', got %v", cfg.Level)
				}
				if !cfg.Colorize {
					t.Error("Expected colorize to be true")
				}
			},
		},
		{
			name: "DatabaseConfig",
			test: func(t *testing.T) {
				cfg := DefaultDatabaseConfig()
				if cfg.Type != "postgres" {
					t.Errorf("Expected type 'postgres', got %v", cfg.Type)
				}
				if cfg.Port != 5432 {
					t.Errorf("Expected port 5432, got %v", cfg.Port)
				}
			},
		},
		{
			name: "RedisConfig",
			test: func(t *testing.T) {
				cfg := DefaultRedisConfig()
				if cfg.Host != "localhost" {
					t.Errorf("Expected host 'localhost', got %v", cfg.Host)
				}
				if cfg.Port != 6379 {
					t.Errorf("Expected port 6379, got %v", cfg.Port)
				}
			},
		},
		{
			name: "SessionConfig",
			test: func(t *testing.T) {
				cfg := DefaultSessionConfig()
				if cfg.CookieName != defaultProcessName {
					t.Errorf("Expected cookie name '%s', got %v", defaultProcessName, cfg.CookieName)
				}
				if cfg.TTL != 86400 {
					t.Errorf("Expected TTL 86400, got %v", cfg.TTL)
				}
			},
		},
		{
			name: "WebServerConfig",
			test: func(t *testing.T) {
				cfg := DefaultWebServerConfig()
				if !cfg.Enabled {
					t.Error("Expected enabled to be true")
				}
				if cfg.Port != 8080 {
					t.Errorf("Expected port 8080, got %v", cfg.Port)
				}
			},
		},
		{
			name: "TasksConfig",
			test: func(t *testing.T) {
				cfg := DefaultTasksConfig()
				if !cfg.Enabled {
					t.Error("Expected enabled to be true")
				}
				if cfg.TaskProcessors != 1 {
					t.Errorf("Expected task processors 1, got %v", cfg.TaskProcessors)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
