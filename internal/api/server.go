package api

// Server is the interface that all servers must implement
type Server interface {
	// Name returns the unique name of the server
	Name() string

	// Initialize sets up the server (called before Start)
	Initialize() error

	// Start starts the server
	Start() error

	// Stop stops the server gracefully
	Stop() error
}

