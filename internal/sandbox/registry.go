package sandbox

import (
	"fmt"
	"sync"
)

// Registry manages available sandbox drivers
type Registry struct {
	mu      sync.RWMutex
	drivers map[string]func() (Sandbox, error)
}

var globalRegistry = &Registry{
	drivers: make(map[string]func() (Sandbox, error)),
}

// Register adds a sandbox driver to the global registry
// This is typically called from driver packages' init() functions
func Register(name string, factory func() (Sandbox, error)) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if factory == nil {
		panic("sandbox: Register factory is nil")
	}
	if _, dup := globalRegistry.drivers[name]; dup {
		panic("sandbox: Register called twice for driver " + name)
	}

	globalRegistry.drivers[name] = factory
}

// Get creates a new instance of the named sandbox driver
func Get(name string) (Sandbox, error) {
	globalRegistry.mu.RLock()
	factory, exists := globalRegistry.drivers[name]
	globalRegistry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sandbox driver %q not found", name)
	}

	return factory()
}

// Available returns a list of registered driver names
func Available() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.drivers))
	for name := range globalRegistry.drivers {
		names = append(names, name)
	}

	return names
}

// IsRegistered checks if a driver is registered
func IsRegistered(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, exists := globalRegistry.drivers[name]
	return exists
}
