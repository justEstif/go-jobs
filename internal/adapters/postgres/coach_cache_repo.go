package postgres

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// CoachCacheRepo implements ports.CoachCacheRepository.
type CoachCacheRepo struct {
	q *queries.Queries
}

// NewCoachCacheRepo constructs a CoachCacheRepo.
func NewCoachCacheRepo(q *queries.Queries) *CoachCacheRepo {
	return &CoachCacheRepo{q: q}
}

// Get returns the cached result for the (userID, jobID, kind) triple.
func (r *CoachCacheRepo) Get(ctx context.Context, userID domain.UserID, jobID domain.JobID, kind domain.CoachCacheKind) (domain.CoachCache, error) {
	row, err := r.q.GetCoachCache(ctx, queries.GetCoachCacheParams{
		UserID: uuidToPg(userID),
		JobID:  uuidToPg(jobID),
		Kind:   string(kind),
	})
	if err != nil {
		return domain.CoachCache{}, fmt.Errorf("get coach cache: %w", err)
	}
	return domain.CoachCache{
		UserID:    pgToUUID(row.UserID),
		JobID:     pgToUUID(row.JobID),
		Kind:      domain.CoachCacheKind(row.Kind),
		Result:    row.Result,
		ModelUsed: row.ModelUsed,
		CreatedAt: pgToTime(row.CreatedAt),
	}, nil
}

// Upsert writes or overwrites a cache entry.
func (r *CoachCacheRepo) Upsert(ctx context.Context, entry domain.CoachCache) error {
	err := r.q.UpsertCoachCache(ctx, queries.UpsertCoachCacheParams{
		UserID:    uuidToPg(entry.UserID),
		JobID:     uuidToPg(entry.JobID),
		Kind:      string(entry.Kind),
		Result:    entry.Result,
		ModelUsed: entry.ModelUsed,
	})
	if err != nil {
		return fmt.Errorf("upsert coach cache: %w", err)
	}
	return nil
}
