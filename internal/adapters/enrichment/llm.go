package enrichment

import (
	"context"
	"errors"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// ErrNoLLMKey is returned when tier 3 enrichment is attempted without a
// configured LLM API key.
var ErrNoLLMKey = errors.New("no LLM API key configured")

// llmEnricher is the tier 3 enrichment adapter. It calls an LLM provider
// (OpenAI, Anthropic, or Google) with a structured-output prompt to fill fields
// that tiers 1 and 2 could not determine.
//
// Currently a stub — full implementation is deferred to M5 when per-user
// LLM key management (via UserService) is in place.
type llmEnricher struct {
	provider domain.LLMProvider
	apiKey   string
}

// newLLMEnricher creates an llmEnricher. apiKey may be empty — calls to
// enrich will return ErrNoLLMKey until a key is provided.
func newLLMEnricher(provider domain.LLMProvider, apiKey string) *llmEnricher {
	return &llmEnricher{
		provider: provider,
		apiKey:   apiKey,
	}
}

// enrich fills remaining zero-value fields in tags using an LLM structured
// output call. Returns ErrNoLLMKey when no API key is configured.
//
// The prompt instructs the model to return JSON with the fields that are still
// empty — only pay for tokens we actually need.
func (l *llmEnricher) enrich(_ context.Context, _ domain.Job, tags domain.JobTags) (domain.JobTags, error) {
	if l.apiKey == "" {
		// No key — return tags as-is; tiers 1 and 2 still ran.
		return tags, ErrNoLLMKey
	}

	// TODO(M5): implement per-provider structured output calls.
	// Providers: openai-go, anthropic-sdk-go, google/generative-ai-go.
	// Use exponential backoff on 429 responses.
	// Only call for fields still zero after tiers 1–2.
	return tags, nil
}
