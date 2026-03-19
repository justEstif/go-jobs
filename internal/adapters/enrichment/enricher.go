package enrichment

import (
	"context"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// TieredEnricher implements ports.JobEnricher via a two-tier pipeline:
//
//  1. ATS metadata extraction (tier 1) — free, instant, no external calls
//  2. Rule-based keyword/regex matching (tier 2) — free, fast, no external calls
//
// LLM-based enrichment has been removed from the background pipeline.
// LLM is now used interactively via the Job Coach feature (per-user, on-demand).
type TieredEnricher struct{}

// NewTieredEnricher constructs a TieredEnricher.
func NewTieredEnricher() ports.JobEnricher {
	return &TieredEnricher{}
}

// Enrich runs the two-tier pipeline against job and returns populated JobTags.
//
// EnrichmentSource records the highest tier that contributed:
//   - "ats"   — tier 1 alone covered all required fields
//   - "rules" — tier 2 was needed to fill remaining gaps
func (e *TieredEnricher) Enrich(_ context.Context, job domain.Job) (domain.JobTags, error) {
	// Tier 1: extract from ATS metadata already in the job record.
	tags := extractFromATS(job)
	tags.EnrichmentSource = domain.EnrichmentATS

	// Tier 2: rule-based extraction. Rules always contribute at minimum
	// Seniority (defaulting to Mid) and a TechStack scan, so the source
	// label is always advanced to EnrichmentRules after this tier.
	tags = applyRules(job, tags)
	tags.EnrichmentSource = domain.EnrichmentRules

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
