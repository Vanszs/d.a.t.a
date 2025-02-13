package data

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	"github.com/carv-protocol/d.a.t.a/src/plugins/core"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"
	"go.uber.org/zap"
)

// init registers the plugin factory
func init() {
	core.RegisterPlugin("d.a.t.a", NewPlugin)
}

// Required configuration keys
const (
	ConfigKeyAPIURL    = "api_url"    // maps to CarvConfig.BaseURL
	ConfigKeyAuthToken = "auth_token" // maps to CarvConfig.APIKey
	ConfigKeyChain     = "chain"      // maps to Token.Network
	ConfigKeyLLM       = "llm"        // LLM configuration section
)

// Plugin implements the core.Plugin interface for data functionality
type Plugin struct {
	config     *core.PluginConfig
	llmClient  llm.Client
	logger     *zap.SugaredLogger
	actions    []core.Action
	providers  []core.Provider
	evaluators []core.Evaluator
	services   []core.Service
	clients    []core.Client
}

// NewPlugin creates a new data plugin
func NewPlugin(llmClient llm.Client) core.Plugin {
	return &Plugin{
		llmClient: llmClient,
		logger:    logger.GetLogger().With(zap.String("plugin", "d.a.t.a")),
		config: &core.PluginConfig{
			Metadata: core.PluginMetadata{
				Name:        "d.a.t.a",
				Description: "Data interaction plugin",
				Version:     "1.0.0",
				Author:      "CARV Protocol",
				License:     "MIT",
				Homepage:    "https://github.com/carv-protocol/d.a.t.a",
				Repository:  "https://github.com/carv-protocol/d.a.t.a",
			},
			Options: make(map[string]interface{}),
		},
	}
}

// Name implements core.Plugin interface
func (p *Plugin) Name() string {
	return p.config.Metadata.Name
}

// Description implements core.Plugin interface
func (p *Plugin) Description() string {
	return p.config.Metadata.Description
}

// Version implements core.Plugin interface
func (p *Plugin) Version() string {
	return p.config.Metadata.Version
}

// Actions implements core.Plugin interface
func (p *Plugin) Actions() []core.Action {
	return p.actions
}

// Providers implements core.Plugin interface
func (p *Plugin) Providers() []core.Provider {
	return p.providers
}

// Evaluators implements core.Plugin interface
func (p *Plugin) Evaluators() []core.Evaluator {
	return p.evaluators
}

// Services implements core.Plugin interface
func (p *Plugin) Services() []core.Service {
	return p.services
}

// Clients implements core.Plugin interface
func (p *Plugin) Clients() []core.Client {
	return p.clients
}

// validateConfig validates the plugin configuration
func (p *Plugin) validateConfig(opts map[string]interface{}) error {
	required := []string{ConfigKeyAPIURL, ConfigKeyAuthToken, ConfigKeyChain, ConfigKeyLLM}
	for _, key := range required {
		val, ok := opts[key]
		if !ok {
			return fmt.Errorf("missing required configuration: %s", key)
		}
		if key == ConfigKeyLLM {
			// Try both map[interface{}]interface{} and map[string]interface{}
			llmConfig, ok := val.(map[string]interface{})
			if !ok {
				// Try the old type as fallback
				if llmConfig2, ok := val.(map[interface{}]interface{}); ok {
					// Convert to map[string]interface{}
					llmConfig = make(map[string]interface{})
					for k, v := range llmConfig2 {
						if kStr, ok := k.(string); ok {
							llmConfig[kStr] = v
						}
					}
				} else {
					return fmt.Errorf("invalid configuration value for %s: must be a map", key)
				}
			}
			if model, ok := llmConfig["model"].(string); !ok || model == "" {
				return fmt.Errorf("invalid or missing model in LLM configuration")
			}
		} else if strVal, ok := val.(string); !ok || strVal == "" {
			return fmt.Errorf("invalid configuration value for %s: must be a non-empty string", key)
		}
	}
	return nil
}

// DatabaseProviderAdapter adapts DatabaseProvider to core.Provider interface
type DatabaseProviderAdapter struct {
	provider actions.DatabaseProvider
}

func (a *DatabaseProviderAdapter) Name() string {
	return "ethereum_database_provider"
}

func (a *DatabaseProviderAdapter) GetData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return a.provider.ProcessQuery(ctx, params)
}

// FetchTransactionActionAdapter adapts FetchTransactionAction to core.Action interface
type FetchTransactionActionAdapter struct {
	action *actions.FetchTransactionAction
	logger *zap.SugaredLogger
}

func NewFetchTransactionActionAdapter(action *actions.FetchTransactionAction, logger *zap.SugaredLogger) *FetchTransactionActionAdapter {
	return &FetchTransactionActionAdapter{
		action: action,
		logger: logger,
	}
}

func (a *FetchTransactionActionAdapter) Name() string {
	return a.action.Name()
}

func (a *FetchTransactionActionAdapter) Description() string {
	return a.action.Description()
}

func (a *FetchTransactionActionAdapter) Type() string {
	return a.action.Type()
}

func (a *FetchTransactionActionAdapter) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	a.logger.Infow("Executing fetch transaction action", "params", params)
	result, err := a.action.Execute(ctx, params)
	if err != nil {
		a.logger.Errorw("Failed to execute fetch transaction action", "error", err)
		return nil, err
	}
	a.logger.Infow("Fetch transaction action executed successfully", "result", result)
	return result, nil
}

// GetAction returns the underlying FetchTransactionAction
func (a *FetchTransactionActionAdapter) GetAction() core.FetchTransactionAction {
	return a.action
}

// Init implements core.Plugin interface
func (p *Plugin) Init(ctx context.Context, opts map[string]interface{}) error {
	// Store configuration
	if p.config == nil {
		p.config = &core.PluginConfig{
			Options: make(map[string]interface{}),
		}
	}

	// Merge options
	for k, v := range opts {
		p.config.Options[k] = v
	}

	// Validate configuration
	if err := p.validateConfig(p.config.Options); err != nil {
		p.logger.Errorw("Failed to validate configuration", "error", err)
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize provider
	llmConfig, ok := p.config.Options[ConfigKeyLLM].(map[string]interface{})
	if !ok {
		// Try converting from map[interface{}]interface{}
		if llmConfig2, ok := p.config.Options[ConfigKeyLLM].(map[interface{}]interface{}); ok {
			llmConfig = make(map[string]interface{})
			for k, v := range llmConfig2 {
				if kStr, ok := k.(string); ok {
					llmConfig[kStr] = v
				}
			}
		} else {
			return fmt.Errorf("invalid LLM configuration type")
		}
	}

	model, ok := llmConfig["model"].(string)
	if !ok {
		return fmt.Errorf("invalid or missing model in LLM configuration")
	}

	provider := providers.NewDatabaseProvider(
		p.config.Options[ConfigKeyAPIURL].(string),
		p.config.Options[ConfigKeyAuthToken].(string),
		p.config.Options[ConfigKeyChain].(string),
		p.llmClient,
		model,
	)
	providerAdapter := &DatabaseProviderAdapter{provider: provider}
	p.providers = append(p.providers, providerAdapter)

	// Initialize action
	action := actions.NewFetchTransactionAction(provider)
	actionAdapter := NewFetchTransactionActionAdapter(action, p.logger)
	p.actions = append(p.actions, actionAdapter)

	return nil
}

// Start implements core.Plugin interface
func (p *Plugin) Start(ctx context.Context) error {
	// Start all services
	for _, service := range p.services {
		if err := service.Start(ctx); err != nil {
			p.logger.Errorw("Failed to start service",
				"service", service.Name(),
				"error", err,
			)
			return fmt.Errorf("failed to start service %s: %w", service.Name(), err)
		}
	}

	// Connect all clients
	for _, client := range p.clients {
		if err := client.Connect(ctx); err != nil {
			p.logger.Errorw("Failed to connect client",
				"client", client.Name(),
				"error", err,
			)
			return fmt.Errorf("failed to connect client %s: %w", client.Name(), err)
		}
	}

	p.logger.Info("d.a.t.a plugin started successfully")
	return nil
}

// Stop implements core.Plugin interface
func (p *Plugin) Stop(ctx context.Context) error {
	p.logger.Info("Stopping data plugin")

	var errs []error

	// Stop all services
	for _, service := range p.services {
		if err := service.Stop(ctx); err != nil {
			p.logger.Errorw("Failed to stop service",
				"service", service.Name(),
				"error", err,
			)
			errs = append(errs, fmt.Errorf("failed to stop service %s: %w", service.Name(), err))
		}
	}

	// Close all clients
	for _, client := range p.clients {
		if err := client.Close(ctx); err != nil {
			p.logger.Errorw("Failed to close client",
				"client", client.Name(),
				"error", err,
			)
			errs = append(errs, fmt.Errorf("failed to close client %s: %w", client.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop plugin cleanly: %v", errs)
	}

	p.logger.Info("Data plugin stopped successfully")
	return nil
}
