package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Process  ProcessConfig
	Logger   LoggerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Session  SessionConfig
	Server   ServerConfig
	Tasks    TasksConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Web WebServerConfig
}

// ProcessConfig holds process configuration
type ProcessConfig struct {
	Name string
}

// DefaultProcessConfig returns default process configuration
func DefaultProcessConfig() ProcessConfig {
	return ProcessConfig{
		Name: "actionhero",
	}
}

// Load loads configuration from files and environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Process:  DefaultProcessConfig(),
		Logger:   DefaultLoggerConfig(),
		Database: DefaultDatabaseConfig(),
		Redis:    DefaultRedisConfig(),
		Session:  DefaultSessionConfig(),
		Server: ServerConfig{
			Web: DefaultWebServerConfig(),
		},
		Tasks: DefaultTasksConfig(),
	}

	// Load .env file (if it exists) - this loads variables into the environment
	// Try multiple locations: .env, .env.local, .env.{NODE_ENV}
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}

	envFiles := []string{".env"}
	if env != "" {
		envFiles = append(envFiles, fmt.Sprintf(".env.%s", env))
	}
	envFiles = append(envFiles, ".env.local")

	// Load .env files (ignore errors if files don't exist)
	for _, envFile := range envFiles {
		_ = godotenv.Load(envFile)
	}

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.actionhero")

	// Environment variables
	viper.SetEnvPrefix("ACTIONHERO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK, we'll use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment-specific config if NODE_ENV is set
	if env != "" {
		viper.SetConfigName(fmt.Sprintf("config.%s", env))
		if err := viper.MergeInConfig(); err != nil {
			// Environment-specific config not found is OK
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading environment config file: %w", err)
			}
		}
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values in viper
func setDefaults() {
	// Process
	viper.SetDefault("process.name", "actionhero")

	// Logger
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.colorize", true)
	viper.SetDefault("logger.timestamp", true)

	// Database
	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "actionhero")
	viper.SetDefault("database.sslmode", "disable")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Session
	viper.SetDefault("session.cookiename", "actionhero")
	viper.SetDefault("session.ttl", 86400)

	// Server
	viper.SetDefault("server.web.enabled", true)
	viper.SetDefault("server.web.host", "0.0.0.0")
	viper.SetDefault("server.web.port", 8080)
	viper.SetDefault("server.web.apiroute", "/api")
	viper.SetDefault("server.web.allowedorigins", "*")
	viper.SetDefault("server.web.allowedmethods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
	viper.SetDefault("server.web.allowedheaders", "Content-Type,Authorization")
	viper.SetDefault("server.web.staticfilesenabled", false)
	viper.SetDefault("server.web.staticfilesroute", "/public")
	viper.SetDefault("server.web.staticfilesdirectory", "./public")

	// Tasks
	viper.SetDefault("tasks.enabled", true)
	viper.SetDefault("tasks.taskprocessors", 1)
	viper.SetDefault("tasks.queues", []string{"default"})
	viper.SetDefault("tasks.timeout", 10000)
	viper.SetDefault("tasks.stuckworkertimeout", 60000)
	viper.SetDefault("tasks.retrystuckjobs", false)
}
