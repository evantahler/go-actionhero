package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evantahler/go-actionhero/internal/api"
	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/gorilla/websocket"
)

// testAction is a simple action for testing
type testAction struct {
	api.BaseAction
	returnData  interface{}
	returnError error
}

func newTestAction(name, route string, method api.HTTPMethod, returnData interface{}, returnError error) *testAction {
	return &testAction{
		BaseAction: api.BaseAction{
			ActionName:        name,
			ActionDescription: "test action",
			ActionWeb: &api.WebConfig{
				Route:  route,
				Method: method,
			},
		},
		returnData:  returnData,
		returnError: returnError,
	}
}

func (a *testAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	if a.returnError != nil {
		return nil, a.returnError
	}
	return map[string]interface{}{
		"data":   a.returnData,
		"params": params,
	}, nil
}

func setupTestServer(t *testing.T) (*WebServer, *api.API) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Web: config.WebServerConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           9999,
				APIRoute:       "/api",
				AllowedOrigins: "*",
				AllowedMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
				AllowedHeaders: "Content-Type,Authorization",
			},
		},
	}

	logger := util.NewLogger(config.LoggerConfig{
		Level:     "error", // Use error level to reduce test output
		Colorize:  false,
		Timestamp: false,
	})

	apiInstance := api.New(cfg, logger)
	webServer := NewWebServer(apiInstance)

	return webServer, apiInstance
}

func TestWebServer_Name(t *testing.T) {
	ws, _ := setupTestServer(t)
	if ws.Name() != "web" {
		t.Errorf("Expected server name 'web', got '%s'", ws.Name())
	}
}

func TestWebServer_Initialize(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register a test action
	action := newTestAction("test:action", "/test", api.HTTPMethodGET, nil, nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Initialize server
	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Check that route was registered
	if len(ws.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(ws.routes))
	}
}

func TestWebServer_CORS(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register a test action
	action := newTestAction("test:cors", "/cors", api.HTTPMethodGET, nil, nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/cors", nil)
	w := httptest.NewRecorder()

	// Handle request
	ws.server.Handler.ServeHTTP(w, req)

	// Check CORS headers
	resp := w.Result()
	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("Expected CORS origin '*', got '%s'", origin)
	}
	if methods := resp.Header.Get("Access-Control-Allow-Methods"); methods != "GET,POST,PUT,DELETE,PATCH,OPTIONS" {
		t.Errorf("Unexpected CORS methods: %s", methods)
	}
	if headers := resp.Header.Get("Access-Control-Allow-Headers"); headers != "Content-Type,Authorization" {
		t.Errorf("Unexpected CORS headers: %s", headers)
	}
}

func TestWebServer_OPTIONS(t *testing.T) {
	ws, _ := setupTestServer(t)
	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Create OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/api/anything", nil)
	w := httptest.NewRecorder()

	// Handle request
	ws.server.Handler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
	}
}

func TestWebServer_RouteMatching(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register test actions
	actions := []api.Action{
		newTestAction("test:get", "/test", api.HTTPMethodGET, "get", nil),
		newTestAction("test:post", "/test", api.HTTPMethodPOST, "post", nil),
		newTestAction("test:param", "/test/:id", api.HTTPMethodGET, "param", nil),
	}

	for _, action := range actions {
		if err := apiInstance.RegisterAction(action); err != nil {
			t.Fatalf("Failed to register action: %v", err)
		}
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		checkData      string
	}{
		{"GET /test", "GET", "/api/test", http.StatusOK, "get"},
		{"POST /test", "POST", "/api/test", http.StatusOK, "post"},
		{"GET with param", "GET", "/api/test/123", http.StatusOK, "param"},
		{"Not found", "GET", "/api/notfound", http.StatusNotFound, ""},
		{"Wrong method", "PUT", "/api/test", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			ws.server.Handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkData != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if !response["success"].(bool) {
					t.Errorf("Expected success=true")
				}

				data := response["data"].(map[string]interface{})
				if data["data"] != tt.checkData {
					t.Errorf("Expected data '%s', got '%v'", tt.checkData, data["data"])
				}
			}
		})
	}
}

func TestWebServer_PathParameters(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register action with path parameters
	action := newTestAction("test:params", "/users/:userId/posts/:postId", api.HTTPMethodGET, nil, nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Make request
	req := httptest.NewRequest("GET", "/api/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	ws.server.Handler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	params := data["params"].(map[string]interface{})

	if params["userId"] != "123" {
		t.Errorf("Expected userId '123', got '%v'", params["userId"])
	}
	if params["postId"] != "456" {
		t.Errorf("Expected postId '456', got '%v'", params["postId"])
	}
}

func TestWebServer_QueryParameters(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	action := newTestAction("test:query", "/query", api.HTTPMethodGET, nil, nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Make request with query parameters
	req := httptest.NewRequest("GET", "/api/query?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()

	ws.server.Handler.ServeHTTP(w, req)

	// Check response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	params := data["params"].(map[string]interface{})

	if params["foo"] != "bar" {
		t.Errorf("Expected foo='bar', got '%v'", params["foo"])
	}
	if params["baz"] != "qux" {
		t.Errorf("Expected baz='qux', got '%v'", params["baz"])
	}
}

func TestWebServer_JSONBody(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	action := newTestAction("test:json", "/json", api.HTTPMethodPOST, nil, nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Create JSON body
	body := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}
	bodyBytes, _ := json.Marshal(body)

	// Make request
	req := httptest.NewRequest("POST", "/api/json", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ws.server.Handler.ServeHTTP(w, req)

	// Check response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	params := data["params"].(map[string]interface{})

	if params["name"] != "John" {
		t.Errorf("Expected name='John', got '%v'", params["name"])
	}
	if params["age"].(float64) != 30 {
		t.Errorf("Expected age=30, got '%v'", params["age"])
	}
}

func TestWebServer_ErrorHandling(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register action that returns an error
	action := newTestAction("test:error", "/error", api.HTTPMethodGET, nil,
		util.NewTypedError(util.ErrorTypeConnectionActionRun, "Something went wrong"))
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Make request
	req := httptest.NewRequest("GET", "/api/error", nil)
	w := httptest.NewRecorder()

	ws.server.Handler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"].(bool) {
		t.Errorf("Expected success=false")
	}

	errorData := response["error"].(map[string]interface{})
	if errorData["code"] != "CONNECTION_ACTION_RUN" {
		t.Errorf("Expected error code 'CONNECTION_ACTION_RUN', got '%v'", errorData["code"])
	}
}

func TestWebServer_CompileRoute(t *testing.T) {
	tests := []struct {
		pattern     string
		path        string
		shouldMatch bool
		params      map[string]string
	}{
		{"/users", "/users", true, map[string]string{}},
		{"/users/:id", "/users/123", true, map[string]string{"id": "123"}},
		{"/users/:userId/posts/:postId", "/users/123/posts/456", true, map[string]string{"userId": "123", "postId": "456"}},
		{"/users/:id", "/users", false, nil},
		{"/users", "/posts", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" -> "+tt.path, func(t *testing.T) {
			regex, paramNames, err := compileRoute(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile route: %v", err)
			}

			matches := regex.FindStringSubmatch(tt.path)
			didMatch := matches != nil

			if didMatch != tt.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.shouldMatch, didMatch)
			}

			if didMatch && tt.params != nil {
				extractedParams := make(map[string]string)
				for i, name := range paramNames {
					extractedParams[name] = matches[i+1]
				}

				for k, v := range tt.params {
					if extractedParams[k] != v {
						t.Errorf("Expected param %s='%s', got '%s'", k, v, extractedParams[k])
					}
				}
			}
		})
	}
}

func TestWebServer_WebSocket(t *testing.T) {
	ws, apiInstance := setupTestServer(t)

	// Register test action
	action := newTestAction("test:ws", "/test", api.HTTPMethodGET, "websocket response", nil)
	if err := apiInstance.RegisterAction(action); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Start server
	if err := ws.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() { _ = ws.Stop() }()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:9999/ws", nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Send action request
	request := map[string]interface{}{
		"type":   "action",
		"action": "test:ws",
		"params": map[string]interface{}{
			"foo": "bar",
		},
	}
	if err := conn.WriteJSON(request); err != nil {
		t.Fatalf("Failed to send WebSocket message: %v", err)
	}

	// Read response
	var response map[string]interface{}
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read WebSocket response: %v", err)
	}

	// Check response
	if response["type"] != "response" {
		t.Errorf("Expected type='response', got '%v'", response["type"])
	}
	if !response["success"].(bool) {
		t.Errorf("Expected success=true")
	}
}

func TestWebServer_WebSocketSubscription(t *testing.T) {
	ws, _ := setupTestServer(t)

	if err := ws.Initialize(); err != nil {
		t.Fatalf("Failed to initialize server: %v", err)
	}

	// Start server
	if err := ws.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() { _ = ws.Stop() }()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Create WebSocket connection
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:9999/ws", nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() { _ = conn.Close() }()

	// Subscribe to channel
	request := map[string]interface{}{
		"type":    "subscribe",
		"channel": "test-channel",
	}
	if err := conn.WriteJSON(request); err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	// Read subscription confirmation
	var response map[string]interface{}
	if err := conn.ReadJSON(&response); err != nil {
		t.Fatalf("Failed to read subscription response: %v", err)
	}

	if response["type"] != "subscribed" {
		t.Errorf("Expected type='subscribed', got '%v'", response["type"])
	}
	if response["channel"] != "test-channel" {
		t.Errorf("Expected channel='test-channel', got '%v'", response["channel"])
	}

	// Send broadcast
	broadcastData := map[string]interface{}{
		"message": "Hello from broadcast",
	}
	if err := ws.Broadcast("test-channel", broadcastData); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Read broadcast message
	var broadcast map[string]interface{}
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}
	if err := conn.ReadJSON(&broadcast); err != nil {
		t.Fatalf("Failed to read broadcast: %v", err)
	}

	if broadcast["type"] != "broadcast" {
		t.Errorf("Expected type='broadcast', got '%v'", broadcast["type"])
	}
	if broadcast["channel"] != "test-channel" {
		t.Errorf("Expected channel='test-channel', got '%v'", broadcast["channel"])
	}

	// Unsubscribe
	unsubRequest := map[string]interface{}{
		"type":    "unsubscribe",
		"channel": "test-channel",
	}
	if err := conn.WriteJSON(unsubRequest); err != nil {
		t.Fatalf("Failed to send unsubscribe message: %v", err)
	}

	// Read unsubscription confirmation
	var unsubResponse map[string]interface{}
	if err := conn.ReadJSON(&unsubResponse); err != nil {
		t.Fatalf("Failed to read unsubscription response: %v", err)
	}

	if unsubResponse["type"] != "unsubscribed" {
		t.Errorf("Expected type='unsubscribed', got '%v'", unsubResponse["type"])
	}
}
