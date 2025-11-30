package api

import (
	"testing"
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
