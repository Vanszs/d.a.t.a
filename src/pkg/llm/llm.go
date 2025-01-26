package llm

import "context"

type State struct {
	Prompt string
}

type Client interface {
	GenerateText(ctx context.Context, state State) (string, error)
}

type clientImpl struct{}

func NewClient() Client {
	return &clientImpl{}
}
