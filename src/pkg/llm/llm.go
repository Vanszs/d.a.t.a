package llm

import (
	"context"
)

type State struct {
	Prompt string
}

type Message struct {
	Role    string
	Content string
}

type CompletionRequest struct {
	Model    string
	Messages []Message
}

type Client interface {
	CreateCompletion(ctx context.Context, request CompletionRequest) (string, error)
}
