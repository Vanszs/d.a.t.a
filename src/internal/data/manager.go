package data

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

type UserInput struct {
	Content string
}

type Manager interface {
	Register(ctx context.Context, source DataSource) error
	FetchData(ctx context.Context, dataType string, input interface{}) ([]DataOutput, error)
}

type DataSource interface {
	Name() string
	Initialize(ctx context.Context) error
	Fetch(ctx context.Context, dataType string, input interface{}) (DataOutput, error)
}

type managerImpl struct {
	sources   map[string]DataSource
	llmClient llm.Client
}

type DataOutput struct {
}

func NewManager(llmClient llm.Client) *managerImpl {
	return &managerImpl{
		sources:   make(map[string]DataSource),
		llmClient: llmClient,
	}
}

// Register a plugin
func (m *managerImpl) Register(ctx context.Context, source DataSource) error {
	if _, exists := m.sources[source.Name()]; exists {
		return fmt.Errorf("data source %s already registered", source.Name())
	}

	if err := source.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize source %s: %w", source.Name(), err)
	}

	m.sources[source.Name()] = source

	return nil
}

func (m *managerImpl) FetchData(ctx context.Context, dataType string, input interface{}) ([]DataOutput, error) {
	dataOutput := []DataOutput{}
	for _, source := range m.sources {
		data, err := source.Fetch(ctx, dataType, input)
		if err != nil {
			return nil, err
		}
		dataOutput = append(dataOutput, data)
	}
	// Initialize other providers...
	return dataOutput, nil
}
