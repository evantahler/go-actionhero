package actions

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/evantahler/go-actionhero/internal/api"
)

const swaggerVersion = "3.0.0"

// SwaggerAction returns API documentation in OpenAPI format
type SwaggerAction struct {
	api.BaseAction
}

// NewSwaggerAction creates and configures a new SwaggerAction
func NewSwaggerAction() *SwaggerAction {
	return &SwaggerAction{
		BaseAction: api.BaseAction{
			ActionName:        "swagger",
			ActionDescription: "Return API documentation in the OpenAPI specification",
			ActionWeb: &api.WebConfig{
				Route:  "/swagger",
				Method: api.HTTPMethodGET,
			},
		},
	}
}

func init() {
	Register(func() api.Action { return NewSwaggerAction() })
}

// Run executes the swagger action
func (a *SwaggerAction) Run(ctx context.Context, params interface{}, conn *api.Connection) (interface{}, error) {
	apiInstance := api.APIFromContext(ctx)
	if apiInstance == nil {
		return nil, fmt.Errorf("API instance not found in context")
	}

	cfg := api.ConfigFromContext(ctx)
	if cfg == nil {
		return nil, fmt.Errorf("config not found in context")
	}

	paths := make(map[string]interface{})
	components := map[string]interface{}{
		"schemas": make(map[string]interface{}),
	}

	actions := apiInstance.GetActions()
	for _, action := range actions {
		webConfig := api.GetActionWeb(action)
		if webConfig == nil || webConfig.Route == "" {
			continue
		}

		// Convert :param format to OpenAPI {param} format
		path := convertRouteToSwagger(webConfig.Route)
		method := strings.ToLower(string(webConfig.Method))
		actionName := api.GetActionName(action)
		tag := strings.Split(actionName, ":")[0]
		summary := api.GetActionDescription(action)
		if summary == "" {
			summary = actionName
		}

		// Extract path parameters
		pathParams := extractPathParameters(webConfig.Route)

		// Build request body for non-GET/HEAD methods with inputs
		var requestBody interface{}
		inputs := api.GetActionInputs(action)
		if inputs != nil && method != "get" && method != "head" {
			schemaName := strings.ReplaceAll(actionName, ":", "_") + "_Request"
			schema := buildSchemaFromStruct(inputs)
			components["schemas"].(map[string]interface{})[schemaName] = schema

			requestBody = map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/" + schemaName,
						},
					},
				},
			}
		}

		// Build path/method entry
		if paths[path] == nil {
			paths[path] = make(map[string]interface{})
		}

		operation := map[string]interface{}{
			"summary":   summary,
			"tags":      []string{tag},
			"responses": buildSwaggerResponses(),
		}

		if len(pathParams) > 0 {
			operation["parameters"] = pathParams
		}

		if requestBody != nil {
			operation["requestBody"] = requestBody
		}

		paths[path].(map[string]interface{})[method] = operation
	}

	document := map[string]interface{}{
		"openapi": swaggerVersion,
		"info": map[string]interface{}{
			"version":     "1.0.0",
			"title":       cfg.Process.Name,
			"description": "Go ActionHero API Server",
			"license": map[string]string{
				"name": "MIT",
			},
		},
		"servers": []map[string]string{
			{
				"url":         fmt.Sprintf("http://%s:%d", cfg.Server.Web.Host, cfg.Server.Web.Port),
				"description": "API Server",
			},
		},
		"paths":      paths,
		"components": components,
	}

	return document, nil
}

// convertRouteToSwagger converts :param format to {param} format
func convertRouteToSwagger(route string) string {
	re := regexp.MustCompile(`:(\w+)`)
	return re.ReplaceAllString(route, "{$1}")
}

// extractPathParameters extracts path parameters from a route
func extractPathParameters(route string) []map[string]interface{} {
	re := regexp.MustCompile(`:(\w+)`)
	matches := re.FindAllStringSubmatch(route, -1)

	if len(matches) == 0 {
		return nil
	}

	params := make([]map[string]interface{}, 0, len(matches))
	for _, match := range matches {
		paramName := match[1]
		params = append(params, map[string]interface{}{
			"name":        paramName,
			"in":          "path",
			"required":    true,
			"schema":      map[string]string{"type": "string"},
			"description": "The " + paramName + " parameter",
		})
	}

	return params
}

// buildSchemaFromStruct builds an OpenAPI schema from a Go struct
func buildSchemaFromStruct(input interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	required := make([]string, 0)
	properties := schema["properties"].(map[string]interface{})

	inputType := reflect.TypeOf(input)
	if inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}

	if inputType.Kind() != reflect.Struct {
		return schema
	}

	for i := 0; i < inputType.NumField(); i++ {
		field := inputType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse json tag (might have options like "name,omitempty")
		fieldName := strings.Split(jsonTag, ",")[0]

		// Determine field type
		fieldSchema := map[string]interface{}{
			"type": getJSONType(field.Type),
		}

		// Check if required
		validateTag := field.Tag.Get("validate")
		if strings.Contains(validateTag, "required") {
			required = append(required, fieldName)
		}

		// Add min/max constraints for strings
		if field.Type.Kind() == reflect.String && validateTag != "" {
			if strings.Contains(validateTag, "min=") {
				minRe := regexp.MustCompile(`min=(\d+)`)
				if matches := minRe.FindStringSubmatch(validateTag); len(matches) > 1 {
					fieldSchema["minLength"] = matches[1]
				}
			}
			if strings.Contains(validateTag, "max=") {
				maxRe := regexp.MustCompile(`max=(\d+)`)
				if matches := maxRe.FindStringSubmatch(validateTag); len(matches) > 1 {
					fieldSchema["maxLength"] = matches[1]
				}
			}
			if strings.Contains(validateTag, "email") {
				fieldSchema["format"] = "email"
			}
		}

		properties[fieldName] = fieldSchema
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// getJSONType converts Go type to JSON schema type
func getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}

// buildSwaggerResponses builds standard OpenAPI response definitions
func buildSwaggerResponses() map[string]interface{} {
	errorSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]string{"type": "string"},
		},
	}

	return map[string]interface{}{
		"200": map[string]interface{}{
			"description": "successful operation",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{},
				},
			},
		},
		"400": map[string]interface{}{
			"description": "Invalid input",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": errorSchema,
				},
			},
		},
		"404": map[string]interface{}{
			"description": "Not Found",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": errorSchema,
				},
			},
		},
		"422": map[string]interface{}{
			"description": "Missing or invalid params",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": errorSchema,
				},
			},
		},
		"500": map[string]interface{}{
			"description": "Server error",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": errorSchema,
				},
			},
		},
	}
}
