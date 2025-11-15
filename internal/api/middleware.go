package api

// MiddlewareResponse allows middleware to modify params and responses
type MiddlewareResponse struct {
	UpdatedParams  interface{}
	UpdatedResponse interface{}
}

// Middleware defines hooks that run before and/or after action execution
type Middleware interface {
	// RunBefore is called before the action runs
	// Can modify params or return an error to halt execution
	RunBefore(params interface{}, conn *Connection) (*MiddlewareResponse, error)

	// RunAfter is called after the action runs
	// Can modify the response
	RunAfter(params interface{}, conn *Connection) (*MiddlewareResponse, error)
}

