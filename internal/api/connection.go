package api

import (
	"sync"
)

// SessionData represents session information
type SessionData struct {
	ID         string
	CookieName string
	CreatedAt  int64
	Data       map[string]interface{}
}

// Connection represents a client connection (HTTP, WebSocket, CLI, etc.)
type Connection struct {
	Type          string
	Identifier    string // e.g., IP address
	ID            string // Unique connection ID
	Session       *SessionData
	Subscriptions map[string]bool
	RawConnection interface{} // Underlying connection (e.g., *websocket.Conn)

	mu            sync.RWMutex
	sessionLoaded bool
}

// NewConnection creates a new connection
func NewConnection(connType, identifier, id string, rawConnection interface{}) *Connection {
	return &Connection{
		Type:          connType,
		Identifier:    identifier,
		ID:            id,
		Subscriptions: make(map[string]bool),
		RawConnection: rawConnection,
	}
}

// Subscribe adds a channel subscription
func (c *Connection) Subscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Subscriptions[channel] = true
}

// Unsubscribe removes a channel subscription
func (c *Connection) Unsubscribe(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Subscriptions, channel)
}

// IsSubscribed checks if the connection is subscribed to a channel
func (c *Connection) IsSubscribed(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Subscriptions[channel]
}

// SetSession sets the session data
func (c *Connection) SetSession(session *SessionData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Session = session
	c.sessionLoaded = true
}

// IsSessionLoaded returns whether the session has been loaded
func (c *Connection) IsSessionLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionLoaded
}
