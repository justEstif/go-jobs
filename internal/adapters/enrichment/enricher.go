package enrichment

import (
	"context"
	"errors"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// TieredEnricher implements ports.JobEnricher via a three-tier pipeline:
//
//  1. ATS metadata extraction (tier 1) — free, instant, no external calls
//  2. Rule-based keyword/regex matching (tier 2) — free, fast, no external calls
//  3. LLM structured output (tier 3) — optional, user-provided API key required
//
// Tiers run in order. Each tier fills only the fields left zero by previous
// tiers. EnrichmentSource is set to the highest tier that contributed.
type TieredEnricher struct {
	llm *llmEnricher
}

// NewTieredEnricher constructs a TieredEnricher.
//
// provider and apiKey configure the optional LLM tier. Pass an empty apiKey
// to disable LLM enrichment — tiers 1 and 2 still run.
func NewTieredEnricher(provider domain.LLMProvider, apiKey string) ports.JobEnricher {
	return &TieredEnricher{
		llm: newLLMEnricher(provider, apiKey),
	}
}

// Enrich runs the three-tier pipeline against job and returns populated JobTags.
//
// EnrichmentSource records the highest tier that contributed:
//   - "ats"   — tier 1 alone covered all required fields
//   - "rules" — tier 2 was needed to fill remaining gaps
//   - "llm"   — tier 3 was invoked (and had a valid API key)
func (e *TieredEnricher) Enrich(ctx context.Context, job domain.Job) (domain.JobTags, error) {
	// Tier 1: extract from ATS metadata already in the job record.
	tags := extractFromATS(job)
	tags.EnrichmentSource = domain.EnrichmentATS

	// If department is available (stored in description prefix by some scrapers),
	// map it to a RoleType. We don't have a dedicated field on domain.Job for
	// department after persistence, so derive from title when needed (tier 2).

	// Tier 2: rule-based extraction.
	tags = applyRules(job, tags)
	// Only upgrade the source label if rules actually contributed something
	// beyond what ATS gave us. We check for non-zero Seniority as a proxy —
	// rules always set seniority (at minimum to Mid), so if it was empty
	// before rules ran, rules contributed.
	if tags.Seniority != "" || len(tags.TechStack) > 0 {
		tags.EnrichmentSource = domain.EnrichmentRules
	}

	// Tier 3: LLM enrichment for any remaining gaps (optional).
	llmTags, err := e.llm.enrich(ctx, job, tags)
	if err == nil {
		tags = llmTags
		tags.EnrichmentSource = domain.EnrichmentLLM
	} else if !errors.Is(err, ErrNoLLMKey) {
		// LLM call failed for a reason other than missing key — log-worthy
		// but not fatal; return what tiers 1–2 produced.
		return finalize(tags), nil
	}
	// ErrNoLLMKey is expected — silently skip LLM tier.

	return finalize(tags), nil
}

// finalize sets EnrichedAt and ensures required fields have sensible defaults.
func finalize(tags domain.JobTags) domain.JobTags {
	tags.EnrichedAt = time.Now()

	// Default RoleType to Other when nothing matched.
	if tags.RoleType == "" {
		tags.RoleType = domain.RoleOther
	}

	// Default Seniority to Mid when nothing matched.
	if tags.Seniority == "" {
		tags.Seniority = domain.SeniorityMid
	}

	// Ensure TechStack is never nil (makes SQL array handling easier).
	if tags.TechStack == nil {
		tags.TechStack = []string{}
	}

	return tags
}
