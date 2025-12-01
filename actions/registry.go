package actions

import (
	"sync"

	"github.com/evantahler/go-actionhero/internal/api"
)

var (
	registry   []func() api.Action
	registryMu sync.RWMutex
)

// Register adds an action constructor to the registry
// This should be called from init() functions in action files
func Register(constructor func() api.Action) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, constructor)
}

// GetAll returns all registered actions
func GetAll() []api.Action {
	registryMu.RLock()
	defer registryMu.RUnlock()

	actions := make([]api.Action, 0, len(registry))
	for _, constructor := range registry {
		actions = append(actions, constructor())
	}
	return actions
}
