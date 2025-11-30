package actions

import (
	"context"

	"github.com/evantahler/go-actionhero/internal/api"
)

// EchoAction echoes back the parameters
type EchoAction struct{}

// Name returns the action name
func (a *EchoAction) Name() string {
	return "echo"
}

// Description returns the action description
func (a *EchoAction) Description() string {
	return "Echoes back the parameters sent to it"
}

// Inputs returns the input schema
func (a *EchoAction) Inputs() interface{} {
	return nil
}

// Middleware returns middleware for this action
func (a *EchoAction) Middleware() []api.Middleware {
	return nil
}

// Web returns the HTTP configuration
func (a *EchoAction) Web() *api.WebConfig {
	return &api.WebConfig{
		Route:  "/echo/:message",
		Method: api.HTTPMethodGET,
	}
}

// Task returns the task configuration
func (a *EchoAction) Task() *api.TaskConfig {
	return nil
}

// Run executes the action
func (a *EchoAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	return map[string]interface{}{
		"received": params,
	}, nil
}
