package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// openaiClient implements ports.LLMClient using the OpenAI API.
type openaiClient struct {
	client openai.Client
	model  string
}

// NewOpenAIClient creates an LLMClient targeting the OpenAI API.
//
// apiKey is the user's OpenAI API key (decrypted before calling this).
// model defaults to "gpt-4o-mini" if empty — cheapest model with native
// JSON schema support.
func NewOpenAIClient(apiKey, model string) ports.LLMClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &openaiClient{client: client, model: model}
}

func (c *openaiClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai complete: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}
