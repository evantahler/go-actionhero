package config

// WebServerConfig holds web server configuration
type WebServerConfig struct {
	Enabled              bool
	Host                 string
	Port                 int
	APIRoute             string
	AllowedOrigins       string
	AllowedMethods       string
	AllowedHeaders       string
	StaticFilesEnabled   bool
	StaticFilesRoute     string
	StaticFilesDirectory string
}

// DefaultWebServerConfig returns default web server configuration
func DefaultWebServerConfig() WebServerConfig {
	return WebServerConfig{
		Enabled:              true,
		Host:                 "0.0.0.0",
		Port:                 8080,
		APIRoute:             "/api",
		AllowedOrigins:       "*",
		AllowedMethods:       "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowedHeaders:       "Content-Type,Authorization",
		StaticFilesEnabled:   false,
		StaticFilesRoute:     "/public",
		StaticFilesDirectory: "./public",
	}
}
