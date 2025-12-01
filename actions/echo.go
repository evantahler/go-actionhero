package actions

import (
	"context"

	"github.com/evantahler/go-actionhero/internal/api"
)

// EchoInput defines the input for the echo action
type EchoInput struct {
	Message string `json:"message"`
}

// EchoOutput defines the output for the echo action
type EchoOutput struct {
	Received map[string]interface{} `json:"received"`
}

// EchoAction echoes back the parameters
type EchoAction struct {
	api.BaseAction
}

// NewEchoAction creates and configures a new EchoAction
func NewEchoAction() *EchoAction {
	return &EchoAction{
		BaseAction: api.BaseAction{
			ActionName:        "echo",
			ActionDescription: "Echoes back the parameters sent to it",
			ActionInputs:      EchoInput{},
			ActionWeb: &api.WebConfig{
				Route:  "/echo/:message",
				Method: api.HTTPMethodGET,
			},
		},
	}
}

func init() {
	Register(func() api.Action { return NewEchoAction() })
}

// Run executes the action
func (a *EchoAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	// For this action, we just echo back all params as-is
	// (This maintains backward compatibility with existing tests)
	return map[string]interface{}{
		"received": params,
	}, nil
}
