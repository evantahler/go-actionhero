package config

// SessionConfig holds session configuration
type SessionConfig struct {
	CookieName string
	TTL        int // Time to live in seconds
}

// DefaultSessionConfig returns default session configuration
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		CookieName: "actionhero",
		TTL:        86400, // 24 hours
	}
}
