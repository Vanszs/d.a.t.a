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
	CreateMemory(ctx context.Context, memory Memory) error
	GetMemory(ctx context.Context, id string) (*Memory, error)
	ListMemories(ctx context.Context, filter MemoryFilter) ([]*Memory, error)
	DeleteMemory(ctx context.Context, id string) error
	SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]*Memory, error)
}

type managerImpl struct {
	store database.Store
}

func NewManager(store database.Store) *managerImpl {
	return &managerImpl{
		store: store,
	}
}

func (m *managerImpl) CreateMemory(ctx context.Context, memory Memory) error {
	return m.store.Insert(ctx, "memory", map[string]interface{}{})
}

func (m *managerImpl) GetMemory(ctx context.Context, id string) (*Memory, error) {
	return nil, nil
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) ListMemories(ctx context.Context, filter MemoryFilter) ([]*Memory, error) {
	return nil, nil
}

// GetState is a generic function to retrieve any type of state
func (m *managerImpl) DeleteMemory(ctx context.Context, id string) error {
	return nil
}

// SaveState is a generic function to save any type of state
func (m *managerImpl) SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]*Memory, error) {
	return nil, nil
}
