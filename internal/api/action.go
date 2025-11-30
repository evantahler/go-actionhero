// Package api provides the core API framework types and interfaces for ActionHero
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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

// Action is the interface that all actions must implement.
// Actions should embed BaseAction and implement only the Run method.
//
// For type safety, define input and output structs for your action,
// then implement Run to handle the conversion from interface{} to your types.
type Action interface {
	// Run executes the action with the given parameters and connection.
	// The params will typically be a map[string]interface{} that should be
	// marshaled into your action's input struct.
	// Returns the response data or an error.
	Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error)
}

// BaseAction provides the base structure for all actions.
// This struct should be embedded in all action implementations.
// Similar to the Bun/TypeScript version, actions declare their configuration
// as struct fields rather than methods.
type BaseAction struct {
	// Name is the unique name of the action (e.g., "status", "user:create")
	ActionName string

	// Description is a human-readable description of the action
	ActionDescription string

	// Inputs represents the input schema for validation and type coercion
	ActionInputs interface{}

	// Middleware is a list of middleware to apply to this action
	ActionMiddleware []Middleware

	// Web is the HTTP route configuration, or nil if not available via HTTP
	ActionWeb *WebConfig

	// Task is the task configuration, or nil if not available as a task
	ActionTask *TaskConfig
}

// GetActionName returns the action's name using reflection
func GetActionName(action Action) string {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Try to get ActionName field
	if nameField := val.FieldByName("ActionName"); nameField.IsValid() {
		return nameField.String()
	}

	// Fallback to type name if ActionName is not set
	t := val.Type()
	return t.Name()
}

// GetActionDescription returns the action's description using reflection
func GetActionDescription(action Action) string {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if descField := val.FieldByName("ActionDescription"); descField.IsValid() {
		desc := descField.String()
		if desc != "" {
			return desc
		}
	}

	// Fallback to default description
	return fmt.Sprintf("An Action: %s", GetActionName(action))
}

// GetActionInputs returns the action's input schema using reflection
func GetActionInputs(action Action) interface{} {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if inputsField := val.FieldByName("ActionInputs"); inputsField.IsValid() {
		return inputsField.Interface()
	}

	return nil
}

// GetActionMiddleware returns the action's middleware using reflection
func GetActionMiddleware(action Action) []Middleware {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if mwField := val.FieldByName("ActionMiddleware"); mwField.IsValid() {
		if mw, ok := mwField.Interface().([]Middleware); ok {
			return mw
		}
	}

	return nil
}

// GetActionWeb returns the action's web configuration using reflection
func GetActionWeb(action Action) *WebConfig {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if webField := val.FieldByName("ActionWeb"); webField.IsValid() {
		if web, ok := webField.Interface().(*WebConfig); ok {
			return web
		}
	}

	return nil
}

// GetActionTask returns the action's task configuration using reflection
func GetActionTask(action Action) *TaskConfig {
	val := reflect.ValueOf(action)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if taskField := val.FieldByName("ActionTask"); taskField.IsValid() {
		if task, ok := taskField.Interface().(*TaskConfig); ok {
			return task
		}
	}

	return nil
}

// MarshalParams is a helper function to convert params (interface{}) to a strongly-typed struct.
// Use this at the beginning of your Run method to get type-safe access to parameters.
//
// Example:
//
//	func (a *MyAction) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
//	    var input MyInput
//	    if err := api.MarshalParams(params, &input); err != nil {
//	        return nil, err
//	    }
//	    // Now use input with full type safety
//	    return MyOutput{...}, nil
//	}
func MarshalParams(params interface{}, target interface{}) error {
	if params == nil {
		return nil
	}

	// Use JSON marshaling to convert params to the target struct
	// This handles map[string]interface{} -> struct conversion nicely
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		targetType := reflect.TypeOf(target)
		if targetType.Kind() == reflect.Ptr {
			targetType = targetType.Elem()
		}
		return fmt.Errorf("failed to unmarshal params to %s: %w", targetType.Name(), err)
	}

	return nil
}
