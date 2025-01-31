package data

import (
	"context"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

type DataSources interface {
	Initialize(ctx context.Context) error
	Fetch(ctx context.Context, input UserInput, llm llm.Client) error
}

type Data struct {
	Source  string
	Content string
}

type IdentityInfo struct {
	ID            string
	Web2IDs       map[string]string // platform -> id
	Web3Addresses []string
	Reputation    float64
	VerifiedAt    time.Time
}

type Provider interface {
	Name() string
	Initialize(ctx context.Context) error
	// GetData(ctx context.Context, query Query) (interface{}, error)
}
