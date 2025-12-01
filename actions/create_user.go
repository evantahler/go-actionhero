package actions

import (
	"context"

	"github.com/evantahler/go-actionhero/internal/api"
)

// CreateUserInput defines the input parameters for creating a user
type CreateUserInput struct {
	Name     string `json:"name" validate:"required,min=3,max=256"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=256"`
}

// CreateUserOutput defines the output structure when a user is created
type CreateUserOutput struct {
	Created bool   `json:"created"`
	UserID  int    `json:"userId"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

// CreateUserAction creates a user (test action)
type CreateUserAction struct {
	api.BaseAction
}

// NewCreateUserAction creates and configures a new CreateUserAction
func NewCreateUserAction() *CreateUserAction {
	return &CreateUserAction{
		BaseAction: api.BaseAction{
			ActionName:        "user:create",
			ActionDescription: "Creates a new user",
			ActionInputs:      CreateUserInput{},
			ActionWeb: &api.WebConfig{
				Route:  "/users",
				Method: api.HTTPMethodPOST,
			},
		},
	}
}

func init() {
	Register(func() api.Action { return NewCreateUserAction() })
}

// Run executes the action with strong typing
func (a *CreateUserAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	// Marshal params to strongly-typed input
	var input CreateUserInput
	if err := api.MarshalParams(params, &input); err != nil {
		return nil, err
	}

	// TODO: In a real implementation, this would:
	// 1. Validate the input (email format, password strength, etc.)
	// 2. Check if user already exists
	// 3. Hash the password
	// 4. Insert into database
	// 5. Return the created user

	// For now, return mock data with strong typing
	return CreateUserOutput{
		Created: true,
		UserID:  123,
		Name:    input.Name,
		Email:   input.Email,
	}, nil
}
