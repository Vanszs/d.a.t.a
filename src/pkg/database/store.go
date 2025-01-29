package database

import (
	"context"
	"time"
)

type Store interface {
	Connect(ctx context.Context) error
	Close() error
}

type MemoryStore interface {
	Store
	CreateMemory(ctx context.Context, memory Memory) error
	GetMemory(ctx context.Context, id string) (Memory, error)
	ListMemories(ctx context.Context, filter MemoryFilter) ([]Memory, error)
	DeleteMemory(ctx context.Context, id string) error
	SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]Memory, error)
}

type Memory struct {
	ID        string
	Type      string
	Content   []byte
	Embedding []float64
	Metadata  map[string]interface{}
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemoryFilter struct {
	Types    []string
	Tags     []string
	FromTime time.Time
	ToTime   time.Time
	Limit    int
	Offset   int
}

type SearchOptions struct {
	Limit          int
	MinSimilarity  float64
	IncludeContent bool
}
