package memory

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
)

type UserInput struct {
}

type Memory struct {
	memoryID  string
	createdAt time.Time
	content   string
}

type Manager interface {
	FetchContext(ctx context.Context, userInput UserInput) ([]*Memory, error)
	GetMemory(ctx context.Context, memoryID string) (*Memory, error)
}

type managerImpl struct {
	workingMemory  WorkingMemory
	longTermMemory LongTermMemory
	store          database.Store
}

func NewManager(store database.Store) *managerImpl {
	return &managerImpl{
		store: store,
	}
}

func (m *managerImpl) Add(ctx context.Context, entry Entry) error {
	return m.workingMemory.Add(ctx, entry)
}

func (m *managerImpl) FetchContext(ctx context.Context, userInput UserInput) ([]*Memory, error) {
	return nil, nil
}

func (m *managerImpl) StoreFailure(ctx context.Context, entry Entry) error {
	return m.longTermMemory.Store(ctx, entry)
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) GetState(ctx context.Context, id string, result interface{}) error {
	return nil
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) GetMemory(ctx context.Context, memoryID string) (*Memory, error) {
	return nil, nil
}

// SaveState is a generic function to save any type of state
func (m *managerImpl) SaveState(ctx context.Context, id string, state interface{}) error {
	return nil
}
