package actions

import (
	"context"
	"testing"

	"github.com/evantahler/go-actionhero/internal/api"
	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
)

const stringType = "string"

func TestSwaggerAction_ValidOpenAPIStructure(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{
			Name: "test-server",
		},
		Server: config.ServerConfig{
			Web: config.WebServerConfig{
				Host: "localhost",
				Port: 8080,
			},
		},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	// Register test actions
	if err := apiInstance.RegisterAction(NewStatusAction()); err != nil {
		t.Fatalf("Failed to register status action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewEchoAction()); err != nil {
		t.Fatalf("Failed to register echo action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewCreateUserAction()); err != nil {
		t.Fatalf("Failed to register create user action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewSwaggerAction()); err != nil {
		t.Fatalf("Failed to register swagger action: %v", err)
	}

	// Create context with API and config
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)
	ctx = context.WithValue(ctx, api.ContextKeyConfig, cfg)

	// Create connection
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)

	// Execute swagger action
	action := NewSwaggerAction()
	response, err := action.Run(ctx, nil, conn)
	if err != nil {
		t.Fatalf("Failed to run swagger action: %v", err)
	}

	// Parse response
	doc, ok := response.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response to be a map")
	}

	// Verify OpenAPI version
	if doc["openapi"] != "3.0.0" {
		t.Errorf("Expected openapi version '3.0.0', got '%v'", doc["openapi"])
	}

	// Verify info section
	info, ok := doc["info"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'info' to be present")
	}

	if info["title"] != "test-server" {
		t.Errorf("Expected title 'test-server', got '%v'", info["title"])
	}

	if info["version"] == nil {
		t.Error("Expected 'version' to be present")
	}

	if info["description"] == nil {
		t.Error("Expected 'description' to be present")
	}

	license, ok := info["license"].(map[string]string)
	if !ok || license["name"] != "MIT" {
		t.Error("Expected license with name 'MIT'")
	}

	// Verify servers section
	servers, ok := doc["servers"].([]map[string]string)
	if !ok || len(servers) == 0 {
		t.Fatal("Expected 'servers' to be present and non-empty")
	}

	if servers[0]["url"] != "http://localhost:8080" {
		t.Errorf("Expected server URL 'http://localhost:8080', got '%v'", servers[0]["url"])
	}
}

func TestSwaggerAction_DocumentsAllActions(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{Name: "test-server"},
		Server:  config.ServerConfig{Web: config.WebServerConfig{Host: "localhost", Port: 8080}},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	// Register test actions
	if err := apiInstance.RegisterAction(NewStatusAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewEchoAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewCreateUserAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}
	if err := apiInstance.RegisterAction(NewSwaggerAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)
	ctx = context.WithValue(ctx, api.ContextKeyConfig, cfg)

	// Execute swagger action
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)
	action := NewSwaggerAction()
	response, err := action.Run(ctx, nil, conn)
	if err != nil {
		t.Fatalf("Failed to run swagger action: %v", err)
	}

	doc := response.(map[string]interface{})
	paths, ok := doc["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'paths' to be present")
	}

	// Verify all web actions are documented
	expectedPaths := []string{"/status", "/echo/{message}", "/users", "/swagger"}
	for _, expectedPath := range expectedPaths {
		if paths[expectedPath] == nil {
			t.Errorf("Expected path '%s' to be documented", expectedPath)
		}
	}

	// Verify status endpoint
	statusPath, ok := paths["/status"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected /status to be an object")
	}

	statusGet, ok := statusPath["get"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected /status to have GET method")
	}

	if statusGet["summary"] != "Return the status of the server" {
		t.Errorf("Expected correct summary for status action, got '%v'", statusGet["summary"])
	}

	// Verify tags
	tags, ok := statusGet["tags"].([]string)
	if !ok || len(tags) == 0 {
		t.Error("Expected status action to have tags")
	} else if tags[0] != "status" {
		t.Errorf("Expected tag 'status', got '%v'", tags[0])
	}
}

func TestSwaggerAction_PathParameters(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{Name: "test-server"},
		Server:  config.ServerConfig{Web: config.WebServerConfig{Host: "localhost", Port: 8080}},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	// Register echo action which has path parameters
	if err := apiInstance.RegisterAction(NewEchoAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)
	ctx = context.WithValue(ctx, api.ContextKeyConfig, cfg)

	// Execute swagger action
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)
	action := NewSwaggerAction()
	response, err := action.Run(ctx, nil, conn)
	if err != nil {
		t.Fatalf("Failed to run swagger action: %v", err)
	}

	doc := response.(map[string]interface{})
	paths := doc["paths"].(map[string]interface{})

	// Verify echo path has path parameters
	echoPath, ok := paths["/echo/{message}"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected /echo/{message} path to be present")
	}

	echoGet := echoPath["get"].(map[string]interface{})
	parameters, ok := echoGet["parameters"].([]map[string]interface{})
	if !ok || len(parameters) == 0 {
		t.Fatal("Expected echo action to have parameters")
	}

	// Verify parameter details
	param := parameters[0]
	if param["name"] != "message" {
		t.Errorf("Expected parameter name 'message', got '%v'", param["name"])
	}
	if param["in"] != "path" {
		t.Errorf("Expected parameter in 'path', got '%v'", param["in"])
	}
	if param["required"] != true {
		t.Error("Expected parameter to be required")
	}

	schema, ok := param["schema"].(map[string]string)
	if !ok || schema["type"] != stringType {
		t.Error("Expected parameter schema type to be 'string'")
	}
}

func TestSwaggerAction_RequestBodySchemas(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{Name: "test-server"},
		Server:  config.ServerConfig{Web: config.WebServerConfig{Host: "localhost", Port: 8080}},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	// Register user:create action which has request body
	if err := apiInstance.RegisterAction(NewCreateUserAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)
	ctx = context.WithValue(ctx, api.ContextKeyConfig, cfg)

	// Execute swagger action
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)
	action := NewSwaggerAction()
	response, err := action.Run(ctx, nil, conn)
	if err != nil {
		t.Fatalf("Failed to run swagger action: %v", err)
	}

	doc := response.(map[string]interface{})
	paths := doc["paths"].(map[string]interface{})
	components := doc["components"].(map[string]interface{})

	// Verify user:create has request body
	usersPath, ok := paths["/users"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected /users path to be present")
	}

	usersPost := usersPath["post"].(map[string]interface{})
	requestBody, ok := usersPost["requestBody"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user:create to have requestBody")
	}

	if requestBody["required"] != true {
		t.Error("Expected requestBody to be required")
	}

	content, ok := requestBody["content"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected requestBody to have content")
	}

	jsonContent := content["application/json"].(map[string]interface{})
	schema := jsonContent["schema"].(map[string]interface{})
	ref, ok := schema["$ref"].(string)
	if !ok {
		t.Fatal("Expected schema to have $ref")
	}

	// Verify schema reference exists in components
	expectedRef := "#/components/schemas/user_create_Request"
	if ref != expectedRef {
		t.Errorf("Expected $ref '%s', got '%s'", expectedRef, ref)
	}

	// Verify the schema is defined in components
	schemas := components["schemas"].(map[string]interface{})
	userSchema, ok := schemas["user_create_Request"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user_create_Request schema to be in components")
	}

	if userSchema["type"] != "object" {
		t.Error("Expected schema type to be 'object'")
	}

	// Verify properties
	properties, ok := userSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema to have properties")
	}

	if properties["name"] == nil {
		t.Error("Expected 'name' property")
	}
	if properties["email"] == nil {
		t.Error("Expected 'email' property")
	}
	if properties["password"] == nil {
		t.Error("Expected 'password' property")
	}

	// Verify email has format
	emailProp := properties["email"].(map[string]interface{})
	if emailProp["format"] != "email" {
		t.Error("Expected email property to have format 'email'")
	}

	// Verify required fields
	required, ok := userSchema["required"].([]string)
	if !ok || len(required) != 3 {
		t.Error("Expected 3 required fields")
	}

	requiredFields := make(map[string]bool)
	for _, field := range required {
		requiredFields[field] = true
	}

	if !requiredFields["name"] || !requiredFields["email"] || !requiredFields["password"] {
		t.Error("Expected name, email, and password to be required")
	}
}

func TestSwaggerAction_StandardResponseCodes(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{Name: "test-server"},
		Server:  config.ServerConfig{Web: config.WebServerConfig{Host: "localhost", Port: 8080}},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	if err := apiInstance.RegisterAction(NewStatusAction()); err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	// Create context
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)
	ctx = context.WithValue(ctx, api.ContextKeyConfig, cfg)

	// Execute swagger action
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)
	action := NewSwaggerAction()
	response, err := action.Run(ctx, nil, conn)
	if err != nil {
		t.Fatalf("Failed to run swagger action: %v", err)
	}

	doc := response.(map[string]interface{})
	paths := doc["paths"].(map[string]interface{})
	statusPath := paths["/status"].(map[string]interface{})
	statusGet := statusPath["get"].(map[string]interface{})

	responses, ok := statusGet["responses"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected responses to be present")
	}

	// Verify standard response codes
	expectedCodes := []string{"200", "400", "404", "422", "500"}
	for _, code := range expectedCodes {
		if responses[code] == nil {
			t.Errorf("Expected response code '%s' to be documented", code)
		}

		resp := responses[code].(map[string]interface{})
		if resp["description"] == nil {
			t.Errorf("Expected response code '%s' to have description", code)
		}

		content := resp["content"].(map[string]interface{})
		jsonContent := content["application/json"].(map[string]interface{})
		if jsonContent["schema"] == nil {
			t.Errorf("Expected response code '%s' to have schema", code)
		}
	}

	// Verify 200 response
	resp200 := responses["200"].(map[string]interface{})
	if resp200["description"] != "successful operation" {
		t.Error("Expected 200 response to have correct description")
	}

	// Verify error responses have error schema
	for _, code := range []string{"400", "404", "422", "500"} {
		resp := responses[code].(map[string]interface{})
		content := resp["content"].(map[string]interface{})
		jsonContent := content["application/json"].(map[string]interface{})
		schema := jsonContent["schema"].(map[string]interface{})
		properties := schema["properties"].(map[string]interface{})

		if properties["error"] == nil {
			t.Errorf("Expected error response '%s' to have 'error' property", code)
		}
	}
}

func TestSwaggerAction_MissingAPIInContext(t *testing.T) {
	// Create context without API
	ctx := context.Background()

	// Create connection
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)

	// Execute swagger action
	action := NewSwaggerAction()
	_, err := action.Run(ctx, nil, conn)

	if err == nil {
		t.Error("Expected error when API is missing from context")
	}

	if err.Error() != "API instance not found in context" {
		t.Errorf("Expected specific error message, got '%v'", err)
	}
}

func TestSwaggerAction_MissingConfigInContext(t *testing.T) {
	// Create API instance
	cfg := &config.Config{
		Process: config.ProcessConfig{Name: "test-server"},
		Server:  config.ServerConfig{Web: config.WebServerConfig{Host: "localhost", Port: 8080}},
	}
	logger := util.NewLogger(config.LoggerConfig{Level: "error"})
	apiInstance := api.New(cfg, logger)

	// Create context with API but without config
	ctx := context.Background()
	ctx = context.WithValue(ctx, api.ContextKeyAPI, apiInstance)

	// Create connection
	conn := api.NewConnection("test", "127.0.0.1", "test-id", nil)

	// Execute swagger action
	action := NewSwaggerAction()
	_, err := action.Run(ctx, nil, conn)

	if err == nil {
		t.Error("Expected error when Config is missing from context")
	}

	if err.Error() != "config not found in context" {
		t.Errorf("Expected specific error message, got '%v'", err)
	}
}
