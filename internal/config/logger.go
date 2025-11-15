package config

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level     string // debug, info, warn, error, fatal
	Colorize  bool   // Enable colored output
	Timestamp bool   // Include timestamps in logs
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:     "info",
		Colorize:  true,
		Timestamp: true,
	}
}
