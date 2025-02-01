package llm

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/openai"
)

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
}

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
	provider     string
	openaiClient *openai.Client
}

func (c *clientImpl) CreateCompletion(ctx context.Context, request CompletionRequest) (string, error) {
	switch c.provider {
	case "openai":
		return c.openaiClient.CreateCompletion(ctx, openai.CompletionRequest{
			Model:    request.Model,
			Messages: toOpenAIMessage(request.Messages),
		})
	}
	return "", nil
}

func NewClient(conf *LLMConfig) Client {
	if conf.Provider == "openai" {
		return &clientImpl{
			provider:     conf.Provider,
			openaiClient: openai.NewClient(conf.APIKey),
		}
	}
	return &clientImpl{}
}

func toOpenAIMessage(messages []Message) []openai.Message {
	var openAIMessages []openai.Message
	for _, message := range messages {
		openAIMessages = append(openAIMessages, openai.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	return openAIMessages
}
