package actions

import "time"

// IAction is an interface for actions that can be executed by the agent
type IAction interface {
	Name() string
	Description() string
	Execute() error
	Type() string
}

// ActionManager is an interface for managing actions
type ActionManager interface {
	Register(action IAction) error
	GetAvailableActions() []IAction
	GetAction(name string) IAction
}

// ActionKind defines the type of action to be executed
type ActionKind string

// ActionInfo represents an action waiting to be executed
type ActionInfo struct {
	ID           string
	Name         string
	Kind         ActionKind
	Parameters   map[string]interface{}
	Priority     float64
	Deadline     time.Time
	Dependencies []string // IDs of actions that must complete first
}
