package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/core"
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

type ManagerImpl struct {
	actions map[string]core.Action
}

func NewManager() *ManagerImpl {
	return &ManagerImpl{
		actions: make(map[string]core.Action),
	}
}

func (m *ManagerImpl) Register(action core.Action) error {
	name := action.Name()
	if _, exists := m.actions[action.Name()]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	m.actions[action.Name()] = action
	return nil
}

func (m *ManagerImpl) GetAvailableActions() []core.Action {
	return nil
}
