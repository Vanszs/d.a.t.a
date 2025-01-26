package data

import (
	"context"

	"data-agent/pkg/llm"
)

type UserInput struct {
}

type Manager interface {
	Initialize(ctx context.Context) error
	FetchFetch(ctx context.Context, input UserInput, llm llm.Client) ([]DataOutput, error)
}

type DataSource interface {
	Initialize(ctx context.Context) error
	Fetch(ctx context.Context) (DataOutput, error)
}

type managerImpl struct {
	sources []DataSource
}

type DataOutput struct {
}

func (m *managerImpl) Initialize(ctx context.Context) error {
	// Initialize all data providers
	for _, source := range m.sources {
		if err := source.Initialize(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (m *managerImpl) Fetch(ctx context.Context) ([]DataOutput, error) {
	dataOutput := []DataOutput{}
	for _, source := range m.sources {
		data, err := source.Fetch(ctx)
		if err != nil {
			return nil, err
		}
		dataOutput = append(dataOutput, data)
	}
	// Initialize other providers...
	return dataOutput, nil
}
