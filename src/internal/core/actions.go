package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ActionType defines the type of action to be executed
type ActionType string

// Action represents an action waiting to be executed
type Action struct {
	ID           string
	Name         string
	Type         ActionType
	Parameters   map[string]interface{}
	Priority     float64
	Deadline     time.Time
	Dependencies []string // IDs of actions that must complete first
}

func (a *Action) Execute(ctx context.Context) error {
	// TODO: Implement me
	return nil
}

type ActionGeneration struct {
	Chain   *ThoughtChain
	Actions []Action
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

const (
	ActionTypeStandard     ActionType = "standard"
	ActionTypeCompound     ActionType = "compound"
	ActionTypeConditional  ActionType = "conditional"
	ActionTypeVerification ActionType = "verification"
)

// convertThoughtChainToActions converts a thought chain into executable actions
func convertThoughtChainToActions(chain *ThoughtChain) ([]Action, error) {
	var actions []Action

	// Track dependencies between actions
	// dependencies := make(map[string][]string)

	var relevantSteps []*ThoughtStep

	// 1. Collect only contributing steps
	for _, step := range chain.Steps {
		if step.ContributesToOutcome {
			relevantSteps = append(relevantSteps, step)
		}
	}

	// Process each thought step
	// for i, step := range relevantSteps {

	// }

	// Update actions with dependencies
	// actions = updateActionDependencies(actions, dependencies)

	return actions, nil
}
