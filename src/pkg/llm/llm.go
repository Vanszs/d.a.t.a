package llm

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/openai"
)

type State struct {
	Prompt string
}

type Client interface {
	CreateCompletion(ctx context.Context, request openai.CompletionRequest) (string, error)
}
