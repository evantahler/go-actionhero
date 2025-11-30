// Package util provides utility functions and types for ActionHero
package util

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents different types of errors in the system
type ErrorType string

// Error type constants for different categories of errors
const (
	// ErrorTypeConnectionActionNotFound occurs when a requested action cannot be found
	ErrorTypeConnectionActionNotFound ErrorType = "CONNECTION_ACTION_NOT_FOUND"
	// ErrorTypeConnectionActionRun occurs when an action fails during execution
	ErrorTypeConnectionActionRun ErrorType = "CONNECTION_ACTION_RUN"
	// ErrorTypeConnectionActionParamRequired occurs when a required parameter is missing
	ErrorTypeConnectionActionParamRequired ErrorType = "CONNECTION_ACTION_PARAM_REQUIRED"
	// ErrorTypeConnectionActionParamValidation occurs when parameter validation fails
	ErrorTypeConnectionActionParamValidation ErrorType = "CONNECTION_ACTION_PARAM_VALIDATION"
	// ErrorTypeConnectionSessionNotFound occurs when a session cannot be found
	ErrorTypeConnectionSessionNotFound ErrorType = "CONNECTION_SESSION_NOT_FOUND"
	// ErrorTypeConnectionNotSubscribed occurs when a connection is not subscribed to a channel
	ErrorTypeConnectionNotSubscribed ErrorType = "CONNECTION_NOT_SUBSCRIBED"
	// ErrorTypeConnectionTypeNotFound occurs when a connection type is not recognized
	ErrorTypeConnectionTypeNotFound ErrorType = "CONNECTION_TYPE_NOT_FOUND"

	// ErrorTypeServerInitialization occurs when server initialization fails
	ErrorTypeServerInitialization ErrorType = "SERVER_INITIALIZATION"
	// ErrorTypeServerStart occurs when server start fails
	ErrorTypeServerStart ErrorType = "SERVER_START"
	// ErrorTypeServerStop occurs when server stop fails
	ErrorTypeServerStop ErrorType = "SERVER_STOP"

	// ErrorTypeActionValidation occurs when action validation fails
	ErrorTypeActionValidation ErrorType = "ACTION_VALIDATION"
)

// TypedError represents an error with a specific type and optional metadata
type TypedError struct {
	Message       string
	Type          ErrorType
	Key           string
	Value         interface{}
	Stack         string
	OriginalError error
}

// Error implements the error interface
func (e *TypedError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("%s: %s (original: %v)", e.Type, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewTypedError creates a new TypedError
func NewTypedError(typ ErrorType, message string, opts ...TypedErrorOption) *TypedError {
	err := &TypedError{
		Type:    typ,
		Message: message,
		Stack:   getStackTrace(),
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

// TypedErrorOption is a function that modifies a TypedError
type TypedErrorOption func(*TypedError)

// WithKey sets the key field
func WithKey(key string) TypedErrorOption {
	return func(e *TypedError) {
		e.Key = key
	}
}

// WithValue sets the value field
func WithValue(value interface{}) TypedErrorOption {
	return func(e *TypedError) {
		e.Value = value
	}
}

// WithOriginalError sets the original error
func WithOriginalError(err error) TypedErrorOption {
	return func(e *TypedError) {
		e.OriginalError = err
	}
}

// getStackTrace returns a formatted stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	lines := strings.Split(string(buf[:n]), "\n")

	// Skip the first line (goroutine info) and return the rest
	if len(lines) > 1 {
		return strings.Join(lines[1:], "\n")
	}
	return ""
}
