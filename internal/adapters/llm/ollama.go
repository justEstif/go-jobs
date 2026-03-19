// Package llm provides LLM client adapters for the Job Coach feature.
//
// Each adapter implements ports.LLMClient. Clients are constructed per-call
// with user-specific credentials — no global state is retained.
package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// ollamaClient implements ports.LLMClient via Ollama's OpenAI-compatible API.
// No API key is required — Ollama runs locally.
type ollamaClient struct {
	client openai.Client
	model  string
}

// NewOllamaClient creates an LLMClient targeting an Ollama instance.
//
// baseURL is the Ollama server URL (e.g. "http://localhost:11434").
// The "/v1" path is appended automatically for OpenAI compatibility.
// model is the Ollama model name (e.g. "llama3.1", "qwen3").
func NewOllamaClient(baseURL, model string) ports.LLMClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1"
	}
	client := openai.NewClient(
		option.WithBaseURL(baseURL+"/v1"),
		option.WithAPIKey("ollama"), // Ollama ignores the key but the SDK requires one
	)
	return &ollamaClient{client: client, model: model}
}

func (c *ollamaClient) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("ollama complete: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("ollama: no choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}
