package llm

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/config"
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

type clientImpl struct {
}

func (c *clientImpl) CreateCompletion(ctx context.Context, request CompletionRequest) (string, error) {
	return "", nil
}

func NewClient(conf *config.Config) Client {
	return &clientImpl{}
}
