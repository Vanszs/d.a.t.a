package core

import (
	"fmt"
	"sync"
	"time"
)

// Action represents an action waiting to be executed
type Action struct {
	ID           string
	Name         string
	Type         string
	Parameters   map[string]interface{}
	Priority     float64
	Deadline     time.Time
	Dependencies []string // IDs of actions that must complete first
}

type ActionRegistry struct {
	actions map[string]Action
	mu      sync.RWMutex
}

func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]Action),
	}
}

func (r *ActionRegistry) Register(action Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := action.Name
	if _, exists := r.actions[name]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	r.actions[name] = action
	return nil
}

func (r *ActionRegistry) Get(name string) (Action, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	action, ok := r.actions[name]
	return action, ok
}

func (r *ActionRegistry) List() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions := make([]Action, 0, len(r.actions))
	for _, action := range r.actions {
		actions = append(actions, action)
	}
	return actions
}
