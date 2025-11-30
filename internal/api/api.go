package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
)

// API is the main singleton that manages the entire ActionHero application
type API struct {
	// Configuration
	Config *config.Config

	// Logger
	Logger *util.Logger

	// Actions registry
	actions map[string]Action
	actionsMu sync.RWMutex

	// Servers
	servers []Server
	serversMu sync.RWMutex

	// Initializers
	initializers []Initializer
	initializersMu sync.RWMutex

	// Lifecycle state
	running bool
	mu sync.RWMutex

	// Context for graceful shutdown
	ctx context.Context
	cancel context.CancelFunc
}

// Initializer represents a plugin-like component that needs initialization
type Initializer interface {
	// Name returns the unique name of the initializer
	Name() string

	// Priority returns the initialization priority (lower runs first)
	Priority() int

	// Initialize sets up the initializer
	Initialize(api *API) error

	// Start starts the initializer (called after all initializers are initialized)
	Start(api *API) error

	// Stop stops the initializer gracefully
	Stop(api *API) error
}

// New creates a new API instance
func New(cfg *config.Config, logger *util.Logger) *API {
	ctx, cancel := context.WithCancel(context.Background())

	return &API{
		Config:       cfg,
		Logger:       logger,
		actions:      make(map[string]Action),
		servers:      make([]Server, 0),
		initializers: make([]Initializer, 0),
		running:      false,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// RegisterAction registers an action in the API
func (a *API) RegisterAction(action Action) error {
	a.actionsMu.Lock()
	defer a.actionsMu.Unlock()

	name := action.Name()
	if _, exists := a.actions[name]; exists {
		return fmt.Errorf("action '%s' is already registered", name)
	}

	a.actions[name] = action
	a.Logger.Debugf("Registered action: %s", name)
	return nil
}

// GetAction retrieves an action by name
func (a *API) GetAction(name string) (Action, bool) {
	a.actionsMu.RLock()
	defer a.actionsMu.RUnlock()

	action, exists := a.actions[name]
	return action, exists
}

// GetActions returns all registered actions
func (a *API) GetActions() []Action {
	a.actionsMu.RLock()
	defer a.actionsMu.RUnlock()

	actions := make([]Action, 0, len(a.actions))
	for _, action := range a.actions {
		actions = append(actions, action)
	}
	return actions
}

// RegisterServer registers a server in the API
func (a *API) RegisterServer(server Server) {
	a.serversMu.Lock()
	defer a.serversMu.Unlock()

	a.servers = append(a.servers, server)
	a.Logger.Debugf("Registered server: %s", server.Name())
}

// GetServers returns all registered servers
func (a *API) GetServers() []Server {
	a.serversMu.RLock()
	defer a.serversMu.RUnlock()

	servers := make([]Server, len(a.servers))
	copy(servers, a.servers)
	return servers
}

// RegisterInitializer registers an initializer in the API
func (a *API) RegisterInitializer(initializer Initializer) {
	a.initializersMu.Lock()
	defer a.initializersMu.Unlock()

	a.initializers = append(a.initializers, initializer)
	a.Logger.Debugf("Registered initializer: %s", initializer.Name())
}

// GetInitializers returns all registered initializers sorted by priority
func (a *API) GetInitializers() []Initializer {
	a.initializersMu.RLock()
	defer a.initializersMu.RUnlock()

	// Create a copy
	initializers := make([]Initializer, len(a.initializers))
	copy(initializers, a.initializers)

	// Sort by priority (lower priority runs first)
	for i := 0; i < len(initializers); i++ {
		for j := i + 1; j < len(initializers); j++ {
			if initializers[i].Priority() > initializers[j].Priority() {
				initializers[i], initializers[j] = initializers[j], initializers[i]
			}
		}
	}

	return initializers
}

// Initialize initializes all components in the proper order
func (a *API) Initialize() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("API is already running")
	}
	a.mu.Unlock()

	a.Logger.Info("Initializing ActionHero...")

	// Initialize all initializers in priority order
	initializers := a.GetInitializers()
	for _, initializer := range initializers {
		a.Logger.Infof("Initializing: %s", initializer.Name())
		if err := initializer.Initialize(a); err != nil {
			return fmt.Errorf("failed to initialize %s: %w", initializer.Name(), err)
		}
	}

	// Initialize all servers
	servers := a.GetServers()
	for _, server := range servers {
		a.Logger.Infof("Initializing server: %s", server.Name())
		if err := server.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize server %s: %w", server.Name(), err)
		}
	}

	a.Logger.Info("ActionHero initialized successfully")
	return nil
}

// Start starts all components
func (a *API) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("API is already running")
	}
	a.running = true
	a.mu.Unlock()

	a.Logger.Info("Starting ActionHero...")

	// Start all initializers in priority order
	initializers := a.GetInitializers()
	for _, initializer := range initializers {
		a.Logger.Infof("Starting: %s", initializer.Name())
		if err := initializer.Start(a); err != nil {
			return fmt.Errorf("failed to start %s: %w", initializer.Name(), err)
		}
	}

	// Start all servers
	servers := a.GetServers()
	for _, server := range servers {
		a.Logger.Infof("Starting server: %s", server.Name())
		if err := server.Start(); err != nil {
			return fmt.Errorf("failed to start server %s: %w", server.Name(), err)
		}
	}

	a.Logger.Info("ActionHero started successfully")
	return nil
}

// Stop stops all components gracefully
func (a *API) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("API is not running")
	}
	a.running = false
	a.mu.Unlock()

	a.Logger.Info("Stopping ActionHero...")

	// Cancel context to signal shutdown
	a.cancel()

	// Stop all servers (in reverse order)
	servers := a.GetServers()
	for i := len(servers) - 1; i >= 0; i-- {
		server := servers[i]
		a.Logger.Infof("Stopping server: %s", server.Name())
		if err := server.Stop(); err != nil {
			a.Logger.Errorf("Error stopping server %s: %v", server.Name(), err)
		}
	}

	// Stop all initializers (in reverse order)
	initializers := a.GetInitializers()
	for i := len(initializers) - 1; i >= 0; i-- {
		initializer := initializers[i]
		a.Logger.Infof("Stopping: %s", initializer.Name())
		if err := initializer.Stop(a); err != nil {
			a.Logger.Errorf("Error stopping %s: %v", initializer.Name(), err)
		}
	}

	a.Logger.Info("ActionHero stopped successfully")
	return nil
}

// IsRunning returns whether the API is currently running
func (a *API) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

// Context returns the API's context (for graceful shutdown)
func (a *API) Context() context.Context {
	return a.ctx
}
