package api

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/sirupsen/logrus"
)

func TestNewConnection(t *testing.T) {
	conn := NewConnection("web", "127.0.0.1", "test-id", nil)

	if conn.Type != "web" {
		t.Errorf("Expected type 'web', got %v", conn.Type)
	}
	if conn.Identifier != "127.0.0.1" {
		t.Errorf("Expected identifier '127.0.0.1', got %v", conn.Identifier)
	}
	if conn.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %v", conn.ID)
	}
	if conn.Subscriptions == nil {
		t.Error("Expected Subscriptions map to be initialized")
	}
}

func TestConnection_Subscribe(t *testing.T) {
	conn := NewConnection("web", "127.0.0.1", "test-id", nil)

	conn.Subscribe("messages")
	if !conn.IsSubscribed("messages") {
		t.Error("Expected connection to be subscribed to 'messages'")
	}
}

func TestConnection_Unsubscribe(t *testing.T) {
	conn := NewConnection("web", "127.0.0.1", "test-id", nil)

	conn.Subscribe("messages")
	conn.Unsubscribe("messages")
	if conn.IsSubscribed("messages") {
		t.Error("Expected connection to not be subscribed to 'messages'")
	}
}

func TestConnection_SetSession(t *testing.T) {
	conn := NewConnection("web", "127.0.0.1", "test-id", nil)
	session := &SessionData{
		ID:         "session-id",
		CookieName: "session",
		CreatedAt:  1234567890,
		Data:       make(map[string]interface{}),
	}

	conn.SetSession(session)

	if conn.Session != session {
		t.Error("Expected session to be set")
	}
	if !conn.IsSessionLoaded() {
		t.Error("Expected session to be marked as loaded")
	}
}

// Mock action for testing logging
type testLogAction struct {
	BaseAction
	shouldError bool
}

func (a *testLogAction) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
	if a.shouldError {
		return nil, util.NewTypedError(util.ErrorTypeConnectionActionRun, "test error")
	}
	return map[string]interface{}{"result": "success"}, nil
}

func TestConnection_Act_LoggingSuccess(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf bytes.Buffer

	// Create logger that writes to buffer
	logger := util.NewLogger(config.LoggerConfig{
		Level:     "info",
		Colorize:  false, // Disable colors for easier testing
		Timestamp: false,
	})
	logger.SetOutput(&logBuf)

	// Use text formatter for easier test assertions
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	// Create API instance
	cfg := &config.Config{}
	apiInstance := New(cfg, logger)

	// Register test action
	action := &testLogAction{
		BaseAction: BaseAction{
			ActionName:        "test:success",
			ActionDescription: "Test action for logging",
		},
		shouldError: false,
	}
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Create connection
	conn := NewConnection("http", "127.0.0.1", "test-conn-id", nil)

	// Execute action
	result := conn.Act(context.Background(), apiInstance, "test:success", map[string]interface{}{"foo": "bar"}, "GET", "http://localhost/test")

	if result.Error != nil {
		t.Fatalf("Expected no error, got %v", result.Error)
	}

	// Get log output
	logOutput := logBuf.String()
	t.Logf("Log output: %s", logOutput)

	// Verify expected strings are in log
	expectedStrings := []string{
		"[ACTION:OK]",
		"test:success",
		"[GET]",
		"127.0.0.1",
		"http://localhost/test",
		`foo`,
		`bar`,
		"ms)",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log to contain %q, but it didn't.\nLog output: %s", expected, logOutput)
		}
	}
}

func TestConnection_Act_LoggingError(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf bytes.Buffer

	// Create logger
	logger := util.NewLogger(config.LoggerConfig{
		Level:     "info",
		Colorize:  false,
		Timestamp: false,
	})
	logger.SetOutput(&logBuf)

	// Use text formatter for easier test assertions
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	// Create API instance
	cfg := &config.Config{}
	apiInstance := New(cfg, logger)

	// Register test action that errors
	action := &testLogAction{
		BaseAction: BaseAction{
			ActionName:        "test:error",
			ActionDescription: "Test error action",
		},
		shouldError: true,
	}
	apiInstance.RegisterAction(action)

	// Create connection
	conn := NewConnection("http", "127.0.0.1", "test-id", nil)

	// Execute action
	result := conn.Act(context.Background(), apiInstance, "test:error", map[string]interface{}{"input": "data"}, "POST", "http://localhost/error")

	if result.Error == nil {
		t.Fatal("Expected error, got nil")
	}

	// Get log output
	logOutput := logBuf.String()
	t.Logf("Log output: %s", logOutput)

	// Verify error logging
	expectedStrings := []string{
		"[ACTION:ERROR]",
		"test:error",
		"[POST]",
		"test error",
		`input`,
		`data`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log to contain %q, but it didn't.\nLog output: %s", expected, logOutput)
		}
	}
}

func TestConnection_Act_LoggingActionNotFound(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf bytes.Buffer

	// Create logger
	logger := util.NewLogger(config.LoggerConfig{
		Level:     "info",
		Colorize:  false,
		Timestamp: false,
	})
	logger.SetOutput(&logBuf)

	// Use text formatter for easier test assertions
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})

	// Create API instance (no actions registered)
	cfg := &config.Config{}
	apiInstance := New(cfg, logger)

	// Create connection
	conn := NewConnection("http", "127.0.0.1", "test-id", nil)

	// Execute nonexistent action
	result := conn.Act(context.Background(), apiInstance, "nonexistent", nil, "GET", "http://localhost/missing")

	if result.Error == nil {
		t.Fatal("Expected error for nonexistent action, got nil")
	}

	// Get log output
	logOutput := logBuf.String()
	t.Logf("Log output: %s", logOutput)

	// Verify 404-style logging
	expectedStrings := []string{
		"[ACTION:ERROR]",
		"nonexistent",
		"action not found",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(logOutput, expected) {
			t.Errorf("Expected log to contain %q, but it didn't.\nLog output: %s", expected, logOutput)
		}
	}
}
