package services

import (
	"context"
	"fmt"
	"log"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// BackfillTagsFn returns a function that enriches un-tagged jobs in bulk.
// Used by the hidden `backfill-tags` CLI command to clear the existing backlog.
func BackfillTagsFn(jobs ports.JobRepository, enricher ports.JobEnricher) func(ctx context.Context, limit int) (enriched, failed int, err error) {
	return func(ctx context.Context, limit int) (enriched, failed int, err error) {
		unenriched, err := jobs.ListUnenriched(ctx, limit)
		if err != nil {
			return 0, 0, fmt.Errorf("list unenriched jobs: %w", err)
		}

		for _, job := range unenriched {
			tags, enrichErr := enricher.Enrich(ctx, job)
			if enrichErr != nil {
				log.Printf("backfill-tags: job %s (%q): %v", job.ID, job.Title, enrichErr)
				failed++
				continue
			}
			if saveErr := jobs.SaveTags(ctx, tags); saveErr != nil {
				log.Printf("backfill-tags: save tags for job %s (%q): %v", job.ID, job.Title, saveErr)
				failed++
				continue
			}
			enriched++
		}

		log.Printf("backfill-tags: complete — enriched=%d failed=%d total=%d", enriched, failed, len(unenriched))
		return enriched, failed, nil
	}
}
