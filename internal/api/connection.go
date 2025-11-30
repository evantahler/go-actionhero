package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/evantahler/go-actionhero/internal/util"
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

// ActResult contains the result of an action execution
type ActResult struct {
	Response interface{}
	Error    error
}

// Act executes an action with the given parameters, handling all middleware,
// logging, and error handling. This is the central method for running actions
// across all connection types (HTTP, WebSocket, CLI, etc.)
func (c *Connection) Act(
	ctx context.Context,
	api *API,
	actionName string,
	params map[string]interface{},
	method string,
	url string,
) ActResult {
	startTime := time.Now()
	loggerStatus := "OK"
	var response interface{}
	var err error

	defer func() {
		// Log the request after execution
		duration := time.Since(startTime).Milliseconds()
		c.logRequest(api.Logger, loggerStatus, actionName, duration, method, url, params, err)
	}()

	// Find the action
	action, exists := api.GetAction(actionName)
	if !exists {
		loggerStatus = "ERROR"
		err = fmt.Errorf("action not found: %s", actionName)
		return ActResult{Response: nil, Error: err}
	}

	// Execute the action
	response, err = action.Run(ctx, params, c)
	if err != nil {
		loggerStatus = "ERROR"
		return ActResult{Response: nil, Error: err}
	}

	return ActResult{Response: response, Error: nil}
}

// logRequest logs the action execution similar to the Bun version
func (c *Connection) logRequest(
	logger *util.Logger,
	status string,
	actionName string,
	duration int64,
	method string,
	url string,
	params map[string]interface{},
	err error,
) {
	// Format status prefix with colors
	var statusPrefix string
	if status == "OK" {
		statusPrefix = logger.ColorizeIf("[ACTION:OK]", util.ColorBlue, true)
	} else {
		statusPrefix = logger.ColorizeIf("[ACTION:ERROR]", util.ColorMagenta, true)
	}

	// Format action name (or "unknown" if not found)
	if actionName == "" {
		actionName = "unknown"
	}

	// Format params as JSON (colorized if enabled)
	paramsJSON := "{}"
	if params != nil {
		// TODO: Sanitize secret params before logging
		if jsonBytes, jsonErr := json.Marshal(params); jsonErr == nil {
			paramsJSON = logger.ColorizeIf(string(jsonBytes), util.ColorGray, false)
		}
	}

	// Format error message
	errorMsg := ""
	if err != nil {
		errorMsg = fmt.Sprintf(" %v", err)
	}

	// Format method
	methodStr := ""
	if method != "" {
		methodStr = fmt.Sprintf(" [%s]", method)
	}

	// Format URL
	urlStr := ""
	if url != "" {
		urlStr = fmt.Sprintf(" (%s)", url)
	}

	// Log the request (matching Bun format)
	logger.Infof("%s %s (%dms)%s %s%s%s %s",
		statusPrefix,
		actionName,
		duration,
		methodStr,
		c.Identifier,
		urlStr,
		errorMsg,
		paramsJSON,
	)
}
