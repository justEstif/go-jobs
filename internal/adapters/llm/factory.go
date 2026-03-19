package llm

import (
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// NewClient constructs the appropriate LLMClient for the given provider.
//
// For Ollama, apiKey is ignored and baseURL defaults to localhost:11434.
// For cloud providers, apiKey must be the decrypted user key.
func NewClient(provider domain.LLMProvider, apiKey, model, baseURL string) (ports.LLMClient, error) {
	switch provider {
	case domain.LLMOllama:
		return NewOllamaClient(baseURL, model), nil
	case domain.LLMOpenAI:
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI requires an API key")
		}
		return NewOpenAIClient(apiKey, model), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q", provider)
	}
}
