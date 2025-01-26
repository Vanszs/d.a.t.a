package memory

import (
	"context"
	"time"
)

type Entry struct {
	ID         string
	Type       string
	Content    interface{}
	Metadata   map[string]interface{}
	Timestamp  time.Time
	Importance float64
	Tags       []string
}

type WorkingMemory interface {
	Add(ctx context.Context, entry Entry) error
	Get(ctx context.Context, id string) (Entry, error)
	GetRecent(ctx context.Context, n int) ([]Entry, error)
	Clear(ctx context.Context) error
}

type LongTermMemory interface {
	Store(ctx context.Context, entry Entry) error
	Retrieve(ctx context.Context, query Query) ([]Entry, error)
	Update(ctx context.Context, id string, entry Entry) error
	Delete(ctx context.Context, id string) error
}

type Query struct {
	Tags       []string
	TimeRange  *TimeRange
	Type       string
	Importance float64
	Limit      int
}
