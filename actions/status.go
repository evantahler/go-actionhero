package actions

import (
	"context"
	"time"

	"github.com/evantahler/go-actionhero/internal/api"
)

// StatusAction returns the server status
type StatusAction struct{}

// Name returns the action name
func (a *StatusAction) Name() string {
	return "status"
}

// Description returns the action description
func (a *StatusAction) Description() string {
	return "Returns server status information"
}

// Inputs returns the input schema
func (a *StatusAction) Inputs() interface{} {
	return nil
}

// Middleware returns middleware for this action
func (a *StatusAction) Middleware() []api.Middleware {
	return nil
}

// Web returns the HTTP configuration
func (a *StatusAction) Web() *api.WebConfig {
	return &api.WebConfig{
		Route:  "/status",
		Method: api.HTTPMethodGET,
	}
}

// Task returns the task configuration
func (a *StatusAction) Task() *api.TaskConfig {
	return nil
}

// Run executes the action
func (a *StatusAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	return map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"uptime":    "running",
	}, nil
}
