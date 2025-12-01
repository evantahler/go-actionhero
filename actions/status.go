package actions

import (
	"context"
	"time"

	"github.com/evantahler/go-actionhero/internal/api"
)

// StatusInput defines the input for the status action (no inputs required)
type StatusInput struct{}

// StatusOutput defines the output structure for the status action
type StatusOutput struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// StatusAction returns the server status
type StatusAction struct {
	api.BaseAction
}

// NewStatusAction creates and configures a new StatusAction
func NewStatusAction() *StatusAction {
	return &StatusAction{
		BaseAction: api.BaseAction{
			ActionName:        "status",
			ActionDescription: "Return the status of the server",
			ActionInputs:      StatusInput{},
			ActionWeb: &api.WebConfig{
				Route:  "/status",
				Method: api.HTTPMethodGET,
			},
		},
	}
}

func init() {
	Register(func() api.Action { return NewStatusAction() })
}

// Run executes the action with strong typing
func (a *StatusAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	// No need to marshal params since this action takes no input
	var input StatusInput
	if err := api.MarshalParams(params, &input); err != nil {
		return nil, err
	}

	// Return strongly-typed output
	return StatusOutput{
		Status:    "ok",
		Timestamp: time.Now().Unix(),
		Uptime:    "running",
	}, nil
}
