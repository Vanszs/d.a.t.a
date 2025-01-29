package memory

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
)

type UserInput struct {
}

type Memory struct {
}

type Manager interface {
	FetchContext(ctx context.Context, userInput UserInput) ([]Memory, error)
}

type managerImpl struct {
	workingMemory  WorkingMemory
	longTermMemory LongTermMemory
}

func NewManager(store database.Store) *managerImpl {
	return &managerImpl{}
}

func (m *managerImpl) Add(ctx context.Context, entry Entry) error {
	return m.workingMemory.Add(ctx, entry)
}

func (m *managerImpl) FetchContext(ctx context.Context, userInput UserInput) ([]Memory, error) {
	return nil, nil
}

func (m *managerImpl) StoreFailure(ctx context.Context, entry Entry) error {
	return m.longTermMemory.Store(ctx, entry)
}
