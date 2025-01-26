package plugins

import (
	"context"
	"fmt"
	"sync"
)

type PluginRegistry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Execute a specific event on all plugins
func (r *PluginRegistry) Execute(ctx context.Context, event string, data interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.plugins {
		if err := plugin.Execute(ctx, event, data); err != nil {
			return fmt.Errorf("plugin %s execution failed: %w", plugin.Name(), err)
		}
	}
	return nil
}

// Register a plugin
func (r *PluginRegistry) Register(ctx context.Context, plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", plugin.Name())
	}

	if err := plugin.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
	}

	r.plugins[plugin.Name()] = plugin
	return nil
}
