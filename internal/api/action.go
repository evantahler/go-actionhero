// Package api provides the core API framework types and interfaces for ActionHero
package api

import (
	"context"
)

// HTTPMethod represents HTTP methods
type HTTPMethod string

// HTTP method constants
const (
	HTTPMethodGET     HTTPMethod = "GET"
	HTTPMethodPOST    HTTPMethod = "POST"
	HTTPMethodPUT     HTTPMethod = "PUT"
	HTTPMethodDELETE  HTTPMethod = "DELETE"
	HTTPMethodPATCH   HTTPMethod = "PATCH"
	HTTPMethodOPTIONS HTTPMethod = "OPTIONS"
)

// WebConfig defines HTTP route configuration for an action
type WebConfig struct {
	Route  string     // Route pattern (e.g., "/user/:id")
	Method HTTPMethod // HTTP method
}

// TaskConfig defines background task configuration for an action
type TaskConfig struct {
	Queue     string // Queue name
	Frequency int64  // Frequency in milliseconds (0 = not recurrent)
}

// Action is the interface that all actions must implement
type Action interface {
	// Name returns the unique name of the action (e.g., "user:create")
	Name() string

	// Description returns a human-readable description of the action
	Description() string

	// Inputs returns a struct instance that represents the input schema
	// This struct will be used for validation and type coercion
	Inputs() interface{}

	// Middleware returns a list of middleware to apply to this action
	Middleware() []Middleware

	// Web returns the HTTP route configuration, or nil if not available via HTTP
	Web() *WebConfig

	// Task returns the task configuration, or nil if not available as a task
	Task() *TaskConfig

	// Run executes the action with the given parameters and connection
	// Returns the response data or an error
	Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error)
}
