package api

import (
	"context"
	"errors"
	"testing"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
)

// Mock Action
type mockAction struct {
	name        string
	description string
}

func (m *mockAction) Name() string             { return m.name }
func (m *mockAction) Description() string      { return m.description }
func (m *mockAction) Inputs() interface{}      { return nil }
func (m *mockAction) Middleware() []Middleware { return nil }
func (m *mockAction) Web() *WebConfig          { return nil }
func (m *mockAction) Task() *TaskConfig        { return nil }
func (m *mockAction) Run(ctx context.Context, params interface{}, conn *Connection) (interface{}, error) {
	return nil, nil
}

// Mock Server
type mockServer struct {
	name             string
	initializeCalled bool
	startCalled      bool
	stopCalled       bool
	shouldFailInit   bool
	shouldFailStart  bool
	shouldFailStop   bool
}

func (m *mockServer) Name() string { return m.name }

func (m *mockServer) Initialize() error {
	m.initializeCalled = true
	if m.shouldFailInit {
		return errors.New("init failed")
	}
	return nil
}

func (m *mockServer) Start() error {
	m.startCalled = true
	if m.shouldFailStart {
		return errors.New("start failed")
	}
	return nil
}

func (m *mockServer) Stop() error {
	m.stopCalled = true
	if m.shouldFailStop {
		return errors.New("stop failed")
	}
	return nil
}

// Mock Initializer
type mockInitializer struct {
	name             string
	priority         int
	initializeCalled bool
	startCalled      bool
	stopCalled       bool
	shouldFailInit   bool
	shouldFailStart  bool
	shouldFailStop   bool
}

func (m *mockInitializer) Name() string  { return m.name }
func (m *mockInitializer) Priority() int { return m.priority }

func (m *mockInitializer) Initialize(api *API) error {
	m.initializeCalled = true
	if m.shouldFailInit {
		return errors.New("init failed")
	}
	return nil
}

func (m *mockInitializer) Start(api *API) error {
	m.startCalled = true
	if m.shouldFailStart {
		return errors.New("start failed")
	}
	return nil
}

func (m *mockInitializer) Stop(api *API) error {
	m.stopCalled = true
	if m.shouldFailStop {
		return errors.New("stop failed")
	}
	return nil
}

func TestNew(t *testing.T) {
	cfg := &config.Config{}
	logger := util.NewLogger(config.DefaultLoggerConfig())

	api := New(cfg, logger)

	if api == nil {
		t.Fatal("Expected API instance, got nil")
	}

	if api.Config != cfg {
		t.Error("Expected Config to be set")
	}

	if api.Logger != logger {
		t.Error("Expected Logger to be set")
	}

	if api.IsRunning() {
		t.Error("Expected API to not be running initially")
	}

	if api.Context() == nil {
		t.Error("Expected Context to be set")
	}
}

func TestRegisterAction(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	action := &mockAction{name: "test:action", description: "Test action"}

	// Register action
	err := api.RegisterAction(action)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to register duplicate
	err = api.RegisterAction(action)
	if err == nil {
		t.Error("Expected error when registering duplicate action")
	}

	// Get action
	retrieved, exists := api.GetAction("test:action")
	if !exists {
		t.Error("Expected action to exist")
	}
	if retrieved != action {
		t.Error("Expected retrieved action to match registered action")
	}

	// Get non-existent action
	_, exists = api.GetAction("nonexistent")
	if exists {
		t.Error("Expected action to not exist")
	}
}

func TestGetActions(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	action1 := &mockAction{name: "action:one"}
	action2 := &mockAction{name: "action:two"}

	api.RegisterAction(action1)
	api.RegisterAction(action2)

	actions := api.GetActions()
	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}
}

func TestRegisterServer(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	server := &mockServer{name: "test-server"}

	api.RegisterServer(server)

	servers := api.GetServers()
	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	if servers[0].Name() != "test-server" {
		t.Errorf("Expected server name 'test-server', got %s", servers[0].Name())
	}
}

func TestRegisterInitializer(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	init1 := &mockInitializer{name: "init1", priority: 10}
	init2 := &mockInitializer{name: "init2", priority: 5}
	init3 := &mockInitializer{name: "init3", priority: 15}

	api.RegisterInitializer(init1)
	api.RegisterInitializer(init2)
	api.RegisterInitializer(init3)

	initializers := api.GetInitializers()
	if len(initializers) != 3 {
		t.Errorf("Expected 3 initializers, got %d", len(initializers))
	}

	// Check that they're sorted by priority
	if initializers[0].Priority() != 5 {
		t.Error("Expected first initializer to have priority 5")
	}
	if initializers[1].Priority() != 10 {
		t.Error("Expected second initializer to have priority 10")
	}
	if initializers[2].Priority() != 15 {
		t.Error("Expected third initializer to have priority 15")
	}
}

func TestInitialize(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	initializer := &mockInitializer{name: "test-init", priority: 1}
	server := &mockServer{name: "test-server"}

	api.RegisterInitializer(initializer)
	api.RegisterServer(server)

	err := api.Initialize()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !initializer.initializeCalled {
		t.Error("Expected initializer Initialize to be called")
	}

	if !server.initializeCalled {
		t.Error("Expected server Initialize to be called")
	}
}

func TestInitializeError(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	initializer := &mockInitializer{name: "failing-init", priority: 1, shouldFailInit: true}
	api.RegisterInitializer(initializer)

	err := api.Initialize()
	if err == nil {
		t.Error("Expected error from Initialize")
	}
}

func TestStartStop(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	initializer := &mockInitializer{name: "test-init", priority: 1}
	server := &mockServer{name: "test-server"}

	api.RegisterInitializer(initializer)
	api.RegisterServer(server)

	// Initialize first
	if err := api.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Start
	err := api.Start()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !api.IsRunning() {
		t.Error("Expected API to be running")
	}

	if !initializer.startCalled {
		t.Error("Expected initializer Start to be called")
	}

	if !server.startCalled {
		t.Error("Expected server Start to be called")
	}

	// Stop
	err = api.Stop()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if api.IsRunning() {
		t.Error("Expected API to not be running")
	}

	if !initializer.stopCalled {
		t.Error("Expected initializer Stop to be called")
	}

	if !server.stopCalled {
		t.Error("Expected server Stop to be called")
	}
}

func TestStartWithoutInitialize(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	// Try to start without initializing
	err := api.Start()
	if err != nil {
		t.Fatalf("Start should succeed even without Initialize: %v", err)
	}

	if !api.IsRunning() {
		t.Error("Expected API to be running")
	}

	api.Stop()
}

func TestDoubleStart(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	api.Start()

	err := api.Start()
	if err == nil {
		t.Error("Expected error when starting already running API")
	}

	api.Stop()
}

func TestStopNotRunning(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	err := api.Stop()
	if err == nil {
		t.Error("Expected error when stopping API that's not running")
	}
}

func TestContext(t *testing.T) {
	api := New(&config.Config{}, util.NewLogger(config.DefaultLoggerConfig()))

	ctx := api.Context()
	if ctx == nil {
		t.Fatal("Expected context to be set")
	}

	// Start and stop API
	api.Start()
	api.Stop()

	// Check that context was cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Expected context to be cancelled after Stop")
	}
}
