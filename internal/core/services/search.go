package services

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// jobSearchService implements ports.JobSearchService.
type jobSearchService struct {
	jobs ports.JobRepository
}

// NewJobSearchService constructs a JobSearchService backed by the given repository.
func NewJobSearchService(jobs ports.JobRepository) ports.JobSearchService {
	return &jobSearchService{jobs: jobs}
}

// Search returns jobs matching the given filters. If userCtx is non-nil,
// results may be annotated with the user's pipeline state.
func (s *jobSearchService) Search(ctx context.Context, filters domain.SearchFilters, userCtx *domain.UserSearchContext) ([]domain.Job, error) {
	jobs, err := s.jobs.Search(ctx, filters, userCtx)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return jobs, nil
}

// GetByID retrieves a single job by its primary key.
func (s *jobSearchService) GetByID(ctx context.Context, id domain.JobID) (domain.Job, error) {
	job, err := s.jobs.GetByID(ctx, id)
	if err != nil {
		return domain.Job{}, fmt.Errorf("get job by id %s: %w", id, err)
	}
	return job, nil
}
