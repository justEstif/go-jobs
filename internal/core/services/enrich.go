package services

import (
	"context"
	"fmt"
	"log"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// enrichService implements ports.EnrichService.
type enrichService struct {
	jobs     ports.JobRepository
	enricher ports.JobEnricher
}

// NewEnrichService constructs an EnrichService.
//
// jobs is used to fetch un-enriched jobs and persist the resulting tags.
// enricher is the tiered enrichment adapter (ATS → rules → LLM).
func NewEnrichService(jobs ports.JobRepository, enricher ports.JobEnricher) ports.EnrichService {
	return &enrichService{
		jobs:     jobs,
		enricher: enricher,
	}
}

// Run fetches up to limit un-enriched jobs, runs each through the enricher,
// and persists the tags. Per-job errors are logged and skipped.
func (s *enrichService) Run(ctx context.Context, limit int) (enriched, failed int, err error) {
	unenriched, err := s.jobs.ListUnenriched(ctx, limit)
	if err != nil {
		return 0, 0, fmt.Errorf("enrich: list unenriched jobs: %w", err)
	}

	for _, job := range unenriched {
		tags, enrichErr := s.enricher.Enrich(ctx, job)
		if enrichErr != nil {
			log.Printf("enrich: job %s (%q): %v", job.ID, job.Title, enrichErr)
			failed++
			continue
		}

		if saveErr := s.jobs.SaveTags(ctx, tags); saveErr != nil {
			log.Printf("enrich: save tags for job %s (%q): %v", job.ID, job.Title, saveErr)
			failed++
			continue
		}

		enriched++
	}

	log.Printf("enrich: complete — enriched=%d failed=%d total=%d", enriched, failed, len(unenriched))
	return enriched, failed, nil
}
