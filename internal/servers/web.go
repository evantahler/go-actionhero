package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/evantahler/go-actionhero/internal/api"
	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebServer implements the Server interface for HTTP and WebSocket
type WebServer struct {
	api    *api.API
	config config.WebServerConfig
	logger *util.Logger

	server   *http.Server
	routes   []routeEntry
	upgrader websocket.Upgrader

	// WebSocket connection management
	connections   map[string]*wsConnection
	connectionsMu sync.RWMutex

	// Channels for broadcasting
	broadcast chan broadcastMessage

	// Shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type routeEntry struct {
	pattern    *regexp.Regexp
	paramNames []string
	method     api.HTTPMethod
	action     api.Action
}

type wsConnection struct {
	conn       *websocket.Conn
	connection *api.Connection
	send       chan []byte
}

type broadcastMessage struct {
	channel string
	data    []byte
}

// NewWebServer creates a new web server instance
func NewWebServer(apiInstance *api.API) *WebServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &WebServer{
		api:         apiInstance,
		config:      apiInstance.Config.Server.Web,
		logger:      apiInstance.Logger,
		routes:      make([]routeEntry, 0),
		connections: make(map[string]*wsConnection),
		broadcast:   make(chan broadcastMessage, 256),
		ctx:         ctx,
		cancel:      cancel,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking based on config
				return true
			},
		},
	}
}

// Name returns the server name
func (ws *WebServer) Name() string {
	return "web"
}

// Initialize sets up the web server
func (ws *WebServer) Initialize() error {
	ws.logger.Info("Initializing web server...")

	// Build routes from registered actions
	actions := ws.api.GetActions()
	for _, action := range actions {
		webConfig := action.Web()
		if webConfig == nil {
			continue
		}

		pattern, paramNames, err := compileRoute(webConfig.Route)
		if err != nil {
			return fmt.Errorf("failed to compile route for action %s: %w", action.Name(), err)
		}

		ws.routes = append(ws.routes, routeEntry{
			pattern:    pattern,
			paramNames: paramNames,
			method:     webConfig.Method,
			action:     action,
		})

		ws.logger.Debugf("Registered route: %s %s -> %s", webConfig.Method, webConfig.Route, action.Name())
	}

	// Create HTTP server
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/ws", ws.handleWebSocket)
	mux.HandleFunc("/", ws.handleHTTP)

	// Add static file serving if enabled
	if ws.config.StaticFilesEnabled {
		fs := http.FileServer(http.Dir(ws.config.StaticFilesDirectory))
		mux.Handle(ws.config.StaticFilesRoute+"/", http.StripPrefix(ws.config.StaticFilesRoute, fs))
		ws.logger.Infof("Static files enabled: %s -> %s", ws.config.StaticFilesRoute, ws.config.StaticFilesDirectory)
	}

	// Wrap with CORS middleware
	handler := ws.corsMiddleware(mux)

	ws.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", ws.config.Host, ws.config.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

// Start starts the web server
func (ws *WebServer) Start() error {
	ws.logger.Infof("Starting web server on %s:%d...", ws.config.Host, ws.config.Port)

	// Start broadcast handler
	ws.wg.Add(1)
	go ws.handleBroadcasts()

	// Start HTTP server in goroutine
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ws.logger.Errorf("Web server error: %v", err)
		}
	}()

	ws.logger.Infof("Web server started successfully")
	return nil
}

// Stop stops the web server gracefully
func (ws *WebServer) Stop() error {
	ws.logger.Info("Stopping web server...")

	// Signal shutdown
	ws.cancel()

	// Close all WebSocket connections
	ws.connectionsMu.Lock()
	for _, conn := range ws.connections {
		if err := conn.conn.Close(); err != nil {
			ws.logger.Warnf("Error closing WebSocket connection: %v", err)
		}
	}
	ws.connectionsMu.Unlock()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ws.server.Shutdown(ctx); err != nil {
		ws.logger.Errorf("Error shutting down web server: %v", err)
		return err
	}

	// Wait for goroutines to finish
	ws.wg.Wait()

	ws.logger.Info("Web server stopped successfully")
	return nil
}

// corsMiddleware adds CORS headers to responses
func (ws *WebServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", ws.config.AllowedOrigins)
		w.Header().Set("Access-Control-Allow-Methods", ws.config.AllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", ws.config.AllowedHeaders)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleHTTP handles HTTP requests
func (ws *WebServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Find matching route
	action, params, err := ws.matchRoute(r.Method, r.URL.Path)
	if err != nil {
		ws.sendError(w, http.StatusNotFound, "ROUTE_NOT_FOUND", err.Error())
		return
	}

	// Parse request parameters
	allParams, err := ws.parseRequest(r, params)
	if err != nil {
		ws.sendError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Create connection
	conn := api.NewConnection("http", r.RemoteAddr, uuid.New().String(), nil)

	// Execute action
	response, err := ws.executeAction(r.Context(), action, allParams, conn)
	if err != nil {
		if typedErr, ok := err.(*util.TypedError); ok {
			ws.sendError(w, typedErr.HTTPStatus(), typedErr.Code(), typedErr.Message)
		} else {
			ws.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	// Send response
	ws.sendSuccess(w, response)
}

// matchRoute finds the action that matches the given method and path
func (ws *WebServer) matchRoute(method, path string) (api.Action, map[string]string, error) {
	// Remove API route prefix if present
	if ws.config.APIRoute != "" && strings.HasPrefix(path, ws.config.APIRoute) {
		path = strings.TrimPrefix(path, ws.config.APIRoute)
	}

	for _, route := range ws.routes {
		if string(route.method) != method {
			continue
		}

		matches := route.pattern.FindStringSubmatch(path)
		if matches == nil {
			continue
		}

		// Extract path parameters
		params := make(map[string]string)
		for i, name := range route.paramNames {
			params[name] = matches[i+1]
		}

		return route.action, params, nil
	}

	return nil, nil, fmt.Errorf("no route found for %s %s", method, path)
}

// parseRequest extracts all parameters from the request
func (ws *WebServer) parseRequest(r *http.Request, pathParams map[string]string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Add path parameters
	for k, v := range pathParams {
		params[k] = v
	}

	// Add query parameters
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}

	// Parse body based on content type
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		contentType := r.Header.Get("Content-Type")

		if strings.Contains(contentType, "application/json") {
			// Parse JSON body
			var jsonBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
				return nil, fmt.Errorf("failed to parse JSON body: %w", err)
			}
			// Merge JSON body params
			for k, v := range jsonBody {
				params[k] = v
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				return nil, fmt.Errorf("failed to parse form data: %w", err)
			}
			for k, v := range r.PostForm {
				if len(v) == 1 {
					params[k] = v[0]
				} else {
					params[k] = v
				}
			}
		}
	}

	return params, nil
}

// executeAction executes an action with the given parameters
func (ws *WebServer) executeAction(ctx context.Context, action api.Action, params map[string]interface{}, conn *api.Connection) (interface{}, error) {
	// TODO: Implement input validation
	// TODO: Implement middleware execution

	// Execute action
	response, err := action.Run(ctx, params, conn)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// sendSuccess sends a successful JSON response
func (ws *WebServer) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ws.logger.Errorf("Error encoding response: %v", err)
	}
}

// sendError sends an error JSON response
func (ws *WebServer) sendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ws.logger.Errorf("Error encoding error response: %v", err)
	}
}

// compileRoute converts a route pattern to a regex
func compileRoute(pattern string) (*regexp.Regexp, []string, error) {
	// Extract parameter names
	paramRegex := regexp.MustCompile(`:(\w+)`)
	paramNames := make([]string, 0)

	for _, match := range paramRegex.FindAllStringSubmatch(pattern, -1) {
		paramNames = append(paramNames, match[1])
	}

	// Convert route pattern to regex
	// Replace :param with regex capturing group
	regexPattern := paramRegex.ReplaceAllString(pattern, `([^/]+)`)
	// Escape forward slashes
	regexPattern = strings.ReplaceAll(regexPattern, "/", `\/`)
	// Add anchors
	regexPattern = "^" + regexPattern + "$"

	compiled, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, nil, err
	}

	return compiled, paramNames, nil
}

// handleWebSocket handles WebSocket upgrade and message handling
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.logger.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create connection
	connID := uuid.New().String()
	apiConn := api.NewConnection("websocket", r.RemoteAddr, connID, conn)

	wsConn := &wsConnection{
		conn:       conn,
		connection: apiConn,
		send:       make(chan []byte, 256),
	}

	// Register connection
	ws.connectionsMu.Lock()
	ws.connections[connID] = wsConn
	ws.connectionsMu.Unlock()

	ws.logger.Debugf("WebSocket connection established: %s", connID)

	// Start goroutines for reading and writing
	ws.wg.Add(2)
	go ws.readWebSocket(wsConn)
	go ws.writeWebSocket(wsConn)
}

// readWebSocket reads messages from WebSocket
func (ws *WebServer) readWebSocket(wsConn *wsConnection) {
	defer func() {
		ws.wg.Done()
		_ = ws.removeConnection(wsConn)
	}()

	for {
		var msg map[string]interface{}
		err := wsConn.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				ws.logger.Errorf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle message
		ws.handleWebSocketMessage(wsConn, msg)
	}
}

// writeWebSocket writes messages to WebSocket
func (ws *WebServer) writeWebSocket(wsConn *wsConnection) {
	defer func() {
		ws.wg.Done()
		if err := wsConn.conn.Close(); err != nil {
			ws.logger.Warnf("Error closing WebSocket connection: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-wsConn.send:
			if !ok {
				if err := wsConn.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					ws.logger.Warnf("Error writing close message: %v", err)
				}
				return
			}

			if err := wsConn.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				ws.logger.Errorf("WebSocket write error: %v", err)
				return
			}

		case <-ws.ctx.Done():
			return
		}
	}
}

// handleWebSocketMessage processes incoming WebSocket messages
func (ws *WebServer) handleWebSocketMessage(wsConn *wsConnection, msg map[string]interface{}) {
	messageType, ok := msg["type"].(string)
	if !ok {
		ws.sendWebSocketError(wsConn, "INVALID_MESSAGE", "Message type is required")
		return
	}

	switch messageType {
	case "action":
		ws.handleWebSocketAction(wsConn, msg)
	case "subscribe":
		ws.handleWebSocketSubscribe(wsConn, msg)
	case "unsubscribe":
		ws.handleWebSocketUnsubscribe(wsConn, msg)
	default:
		ws.sendWebSocketError(wsConn, "UNKNOWN_MESSAGE_TYPE", fmt.Sprintf("Unknown message type: %s", messageType))
	}
}

// handleWebSocketAction executes an action via WebSocket
func (ws *WebServer) handleWebSocketAction(wsConn *wsConnection, msg map[string]interface{}) {
	actionName, ok := msg["action"].(string)
	if !ok {
		ws.sendWebSocketError(wsConn, "INVALID_MESSAGE", "Action name is required")
		return
	}

	action, exists := ws.api.GetAction(actionName)
	if !exists {
		ws.sendWebSocketError(wsConn, "ACTION_NOT_FOUND", fmt.Sprintf("Action not found: %s", actionName))
		return
	}

	params, ok := msg["params"].(map[string]interface{})
	if !ok {
		params = make(map[string]interface{})
	}

	// Execute action
	response, err := ws.executeAction(context.Background(), action, params, wsConn.connection)
	if err != nil {
		if typedErr, ok := err.(*util.TypedError); ok {
			ws.sendWebSocketError(wsConn, typedErr.Code(), typedErr.Message)
		} else {
			ws.sendWebSocketError(wsConn, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	// Send response
	ws.sendWebSocketSuccess(wsConn, response)
}

// handleWebSocketSubscribe handles subscription requests
func (ws *WebServer) handleWebSocketSubscribe(wsConn *wsConnection, msg map[string]interface{}) {
	channel, ok := msg["channel"].(string)
	if !ok {
		ws.sendWebSocketError(wsConn, "INVALID_MESSAGE", "Channel name is required")
		return
	}

	wsConn.connection.Subscribe(channel)
	ws.logger.Debugf("Connection %s subscribed to channel: %s", wsConn.connection.ID, channel)

	// Send confirmation
	response := map[string]interface{}{
		"type":    "subscribed",
		"channel": channel,
	}
	data, _ := json.Marshal(response)
	wsConn.send <- data
}

// handleWebSocketUnsubscribe handles unsubscription requests
func (ws *WebServer) handleWebSocketUnsubscribe(wsConn *wsConnection, msg map[string]interface{}) {
	channel, ok := msg["channel"].(string)
	if !ok {
		ws.sendWebSocketError(wsConn, "INVALID_MESSAGE", "Channel name is required")
		return
	}

	wsConn.connection.Unsubscribe(channel)
	ws.logger.Debugf("Connection %s unsubscribed from channel: %s", wsConn.connection.ID, channel)

	// Send confirmation
	response := map[string]interface{}{
		"type":    "unsubscribed",
		"channel": channel,
	}
	data, _ := json.Marshal(response)
	wsConn.send <- data
}

// sendWebSocketSuccess sends a success message via WebSocket
func (ws *WebServer) sendWebSocketSuccess(wsConn *wsConnection, data interface{}) {
	response := map[string]interface{}{
		"type":    "response",
		"success": true,
		"data":    data,
	}
	responseData, _ := json.Marshal(response)
	wsConn.send <- responseData
}

// sendWebSocketError sends an error message via WebSocket
func (ws *WebServer) sendWebSocketError(wsConn *wsConnection, code, message string) {
	response := map[string]interface{}{
		"type":    "response",
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	responseData, _ := json.Marshal(response)
	wsConn.send <- responseData
}

// removeConnection removes a WebSocket connection
func (ws *WebServer) removeConnection(wsConn *wsConnection) error {
	ws.connectionsMu.Lock()
	delete(ws.connections, wsConn.connection.ID)
	ws.connectionsMu.Unlock()

	close(wsConn.send)
	if err := wsConn.conn.Close(); err != nil {
		ws.logger.Warnf("Error closing WebSocket connection: %v", err)
		return err
	}

	ws.logger.Debugf("WebSocket connection closed: %s", wsConn.connection.ID)
	return nil
}

// handleBroadcasts handles broadcasting messages to subscribed connections
func (ws *WebServer) handleBroadcasts() {
	defer ws.wg.Done()

	for {
		select {
		case msg := <-ws.broadcast:
			ws.connectionsMu.RLock()
			for _, conn := range ws.connections {
				if conn.connection.IsSubscribed(msg.channel) {
					select {
					case conn.send <- msg.data:
					default:
						// Channel full, skip this message
						ws.logger.Warnf("Failed to send broadcast to connection %s (channel full)", conn.connection.ID)
					}
				}
			}
			ws.connectionsMu.RUnlock()

		case <-ws.ctx.Done():
			return
		}
	}
}

// Broadcast sends a message to all connections subscribed to a channel
func (ws *WebServer) Broadcast(channel string, data interface{}) error {
	message := map[string]interface{}{
		"type":    "broadcast",
		"channel": channel,
		"data":    data,
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %w", err)
	}

	select {
	case ws.broadcast <- broadcastMessage{channel: channel, data: messageData}:
		return nil
	case <-ws.ctx.Done():
		return fmt.Errorf("server is shutting down")
	default:
		return fmt.Errorf("broadcast channel is full")
	}
}
