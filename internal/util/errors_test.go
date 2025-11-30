package util

import (
	"errors"
	"strings"
	"testing"
)

func TestTypedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *TypedError
		want     string
		contains []string
	}{
		{
			name:     "simple error",
			err:      NewTypedError(ErrorTypeActionValidation, "test error"),
			contains: []string{"ACTION_VALIDATION", "test error"},
		},
		{
			name: "error with original error",
			err: NewTypedError(
				ErrorTypeConnectionActionRun,
				"failed to run",
				WithOriginalError(errors.New("original error")),
			),
			contains: []string{"CONNECTION_ACTION_RUN", "failed to run", "original error"},
		},
		{
			name: "error with key and value",
			err: NewTypedError(
				ErrorTypeConnectionActionParamRequired,
				"missing param",
				WithKey("email"),
				WithValue("test@example.com"),
			),
			contains: []string{"CONNECTION_ACTION_PARAM_REQUIRED", "missing param"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, want := range tt.contains {
				if !contains(got, want) {
					t.Errorf("TypedError.Error() = %v, should contain %v", got, want)
				}
			}
		})
	}
}

func TestNewTypedError(t *testing.T) {
	err := NewTypedError(ErrorTypeActionValidation, "test message")

	if err.Type != ErrorTypeActionValidation {
		t.Errorf("Expected type %v, got %v", ErrorTypeActionValidation, err.Type)
	}
	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got %v", err.Message)
	}
	if err.Stack == "" {
		t.Error("Expected stack trace to be set")
	}
}

func TestTypedError_WithOptions(t *testing.T) {
	originalErr := errors.New("original")
	err := NewTypedError(
		ErrorTypeConnectionActionRun,
		"test",
		WithKey("testKey"),
		WithValue("testValue"),
		WithOriginalError(originalErr),
	)

	if err.Key != "testKey" {
		t.Errorf("Expected key 'testKey', got %v", err.Key)
	}
	if err.Value != "testValue" {
		t.Errorf("Expected value 'testValue', got %v", err.Value)
	}
	if err.OriginalError != originalErr {
		t.Errorf("Expected original error, got %v", err.OriginalError)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		strings.Contains(s, substr))
}
