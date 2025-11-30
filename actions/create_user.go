package actions

import (
	"context"

	"github.com/evantahler/go-actionhero/internal/api"
)

// CreateUserAction creates a user (test action)
type CreateUserAction struct{}

// Name returns the action name
func (a *CreateUserAction) Name() string {
	return "user:create"
}

// Description returns the action description
func (a *CreateUserAction) Description() string {
	return "Creates a new user"
}

// Inputs returns the input schema
func (a *CreateUserAction) Inputs() interface{} {
	return nil
}

// Middleware returns middleware for this action
func (a *CreateUserAction) Middleware() []api.Middleware {
	return nil
}

// Web returns the HTTP configuration
func (a *CreateUserAction) Web() *api.WebConfig {
	return &api.WebConfig{
		Route:  "/users",
		Method: api.HTTPMethodPOST,
	}
}

// Task returns the task configuration
func (a *CreateUserAction) Task() *api.TaskConfig {
	return nil
}

// Run executes the action
func (a *CreateUserAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	return map[string]interface{}{
		"created": true,
		"params":  params,
	}, nil
}
