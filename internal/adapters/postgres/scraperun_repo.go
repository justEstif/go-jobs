package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// ScrapeRunRepo implements ports.ScrapeRunRepository against PostgreSQL.
type ScrapeRunRepo struct {
	q *queries.Queries
}

// NewScrapeRunRepo constructs a ScrapeRunRepo backed by the given Queries handle.
func NewScrapeRunRepo(q *queries.Queries) *ScrapeRunRepo {
	return &ScrapeRunRepo{q: q}
}

// Create inserts a new ScrapeRun record.
func (r *ScrapeRunRepo) Create(ctx context.Context, run domain.ScrapeRun) error {
	err := r.q.CreateScrapeRun(ctx, queries.CreateScrapeRunParams{
		ID:        uuidToPg(run.ID),
		StartedAt: timeToPg(run.StartedAt),
		Status:    string(run.Status),
	})
	if err != nil {
		return fmt.Errorf("create scrape_run %s: %w", run.ID, err)
	}
	return nil
}

// Update writes progress fields (finished_at, status, job counts, error) back to the DB.
func (r *ScrapeRunRepo) Update(ctx context.Context, run domain.ScrapeRun) error {
	finishedAt := timePtrToPg(run.FinishedAt)
	if run.FinishedAt == nil && run.Status != domain.ScrapeStatusRunning {
		now := time.Now()
		finishedAt = timeToPg(now)
	}
	err := r.q.UpdateScrapeRun(ctx, queries.UpdateScrapeRunParams{
		ID:          uuidToPg(run.ID),
		FinishedAt:  finishedAt,
		Status:      string(run.Status),
		JobsAdded:   int32(run.JobsAdded),
		JobsUpdated: int32(run.JobsUpdated),
		JobsRemoved: int32(run.JobsRemoved),
		Error:       run.Error,
	})
	if err != nil {
		return fmt.Errorf("update scrape_run %s: %w", run.ID, err)
	}
	return nil
}

// Latest returns the most recent ScrapeRun, ordered by started_at DESC.
func (r *ScrapeRunRepo) Latest(ctx context.Context) (domain.ScrapeRun, error) {
	row, err := r.q.GetLatestScrapeRun(ctx)
	if err != nil {
		return domain.ScrapeRun{}, fmt.Errorf("get latest scrape_run: %w", err)
	}
	return domainScrapeRunFromDB(row), nil
}
