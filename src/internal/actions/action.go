package actions

import (
	"fmt"
)

type Action interface {
	Name() string
	Description() string
	Execute() error
	Type() string
}

type Manager interface {
	Register(action Action) error
	GetAvailableActions() []Action
}

// ActionType defines the type of action to be executed
type ActionType string

// // Action represents an action waiting to be executed
// type Action struct {
// 	ID           string
// 	Name         string
// 	Type         ActionType
// 	Parameters   map[string]interface{}
// 	Priority     float64
// 	Deadline     time.Time
// 	Dependencies []string // IDs of actions that must complete first
// }

// func (a *Action) Execute(ctx context.Context) error {
// 	// TODO: Implement me
// 	return nil
// }

type ManagerImpl struct {
	actions map[string]Action
}

func NewManager() *ManagerImpl {
	return &ManagerImpl{
		actions: make(map[string]Action),
	}
}

func (m *ManagerImpl) Register(action Action) error {
	name := action.Name
	if _, exists := m.actions[action.Name()]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	m.actions[action.Name()] = action
	return nil
}

func (m *ManagerImpl) GetAvailableActions() []Action {
	return nil
}
