package config

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string // postgres, sqlite, etc.
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// DefaultDatabaseConfig returns default database configuration
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Database: "actionhero",
		SSLMode:  "disable",
	}
}

