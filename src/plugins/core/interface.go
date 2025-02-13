package core

import (
	"context"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Basic information
	Name() string
	Description() string
	Version() string

	// Components
	Actions() []Action
	Providers() []Provider
	Evaluators() []Evaluator
	Services() []Service
	Clients() []Client

	// Lifecycle
	Init(ctx context.Context, opts map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Action defines the interface for plugin actions
type Action interface {
	// Name returns the unique name of the action
	Name() string

	// Description returns the description of the action
	Description() string

	// Execute executes the action with given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// FetchTransactionAction defines the interface for fetch transaction action
type FetchTransactionAction interface {
	Action
	ExecuteWithParams(ctx context.Context, query string, params map[string]interface{}) (interface{}, error)
}

// FetchTransactionActionAdapter defines the interface for fetch transaction action adapter
type FetchTransactionActionAdapter interface {
	Action
	GetAction() FetchTransactionAction
}

// Provider defines the interface for plugin data providers
type Provider interface {
	// Name returns the unique name of the provider
	Name() string

	// GetData retrieves data with given parameters
	GetData(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Evaluator defines the interface for plugin evaluators
type Evaluator interface {
	// Name returns the unique name of the evaluator
	Name() string

	// Evaluate performs evaluation with given parameters
	Evaluate(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Service defines the interface for plugin services
type Service interface {
	// Name returns the unique name of the service
	Name() string

	// Start starts the service
	Start(ctx context.Context) error

	// Stop stops the service
	Stop(ctx context.Context) error
}

// Client defines the interface for plugin clients
type Client interface {
	// Name returns the unique name of the client
	Name() string

	// Connect establishes connection
	Connect(ctx context.Context) error

	// Close closes the connection
	Close(ctx context.Context) error
}

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name        string
	Description string
	Version     string
	Author      string
	License     string
	Homepage    string
	Repository  string
}

// PluginConfig contains plugin configuration
type PluginConfig struct {
	// Plugin metadata
	Metadata PluginMetadata

	// Plugin options
	Options map[string]interface{}

	// Dependencies
	Dependencies []string
}
