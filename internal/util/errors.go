package util

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents different types of errors in the system
type ErrorType string

const (
	// Connection errors
	ErrorTypeConnectionActionNotFound      ErrorType = "CONNECTION_ACTION_NOT_FOUND"
	ErrorTypeConnectionActionRun           ErrorType = "CONNECTION_ACTION_RUN"
	ErrorTypeConnectionActionParamRequired ErrorType = "CONNECTION_ACTION_PARAM_REQUIRED"
	ErrorTypeConnectionActionParamValidation ErrorType = "CONNECTION_ACTION_PARAM_VALIDATION"
	ErrorTypeConnectionSessionNotFound     ErrorType = "CONNECTION_SESSION_NOT_FOUND"
	ErrorTypeConnectionNotSubscribed        ErrorType = "CONNECTION_NOT_SUBSCRIBED"
	ErrorTypeConnectionTypeNotFound         ErrorType = "CONNECTION_TYPE_NOT_FOUND"

	// Server errors
	ErrorTypeServerInitialization ErrorType = "SERVER_INITIALIZATION"
	ErrorTypeServerStart          ErrorType = "SERVER_START"
	ErrorTypeServerStop           ErrorType = "SERVER_STOP"

	// Action errors
	ErrorTypeActionValidation ErrorType = "ACTION_VALIDATION"
)

// TypedError represents an error with a specific type and optional metadata
type TypedError struct {
	Message      string
	Type         ErrorType
	Key          string
	Value        interface{}
	Stack        string
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

