package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// applicationService implements ports.ApplicationService.
type applicationService struct {
	userJobs ports.UserJobRepository
	jobs     ports.JobRepository
}

// NewApplicationService constructs an ApplicationService.
//
// userJobs manages pipeline state rows; jobs is used to hydrate full Job
// values when listing by status or pipeline.
func NewApplicationService(userJobs ports.UserJobRepository, jobs ports.JobRepository) ports.ApplicationService {
	return &applicationService{
		userJobs: userJobs,
		jobs:     jobs,
	}
}

// SetStatus transitions a job to the given pipeline status.
//
// Business rules enforced here:
//   - Setting StatusApplied automatically ensures StatusInterested is recorded
//     first (if not already set). The DB captures AppliedAt on first applied
//     transition and never overwrites it.
//   - Setting StatusInterested on an already-Applied (or later) job is a no-op
//     to prevent moving backwards in the pipeline.
func (s *applicationService) SetStatus(ctx context.Context, userID domain.UserID, jobID domain.JobID, status domain.ApplicationStatus) error {
	// Read existing state. Not-found means first time tracking this job.
	// Propagate genuine infrastructure errors immediately.
	existing, err := s.userJobs.GetUserJob(ctx, userID, jobID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return fmt.Errorf("set status: read existing user_job: %w", err)
	}
	notFound := errors.Is(err, ports.ErrNotFound)

	// Business rule: setting Applied must also set Interested first.
	if status == domain.StatusApplied && (notFound || existing.Status != domain.StatusInterested) {
		// Only auto-set Interested if we haven't already passed it.
		if notFound {
			intJob := domain.UserJob{
				UserID:   userID,
				JobID:    jobID,
				Status:   domain.StatusInterested,
				StatusAt: time.Now(),
			}
			if upsertErr := s.userJobs.Upsert(ctx, intJob); upsertErr != nil {
				return fmt.Errorf("set status: auto-interested: %w", upsertErr)
			}
			// Re-read to get correct state for the applied upsert below.
			existing, err = s.userJobs.GetUserJob(ctx, userID, jobID)
			if err != nil {
				return fmt.Errorf("set status: re-read after auto-interested: %w", err)
			}
			notFound = false
		}
		// If already past interested (applied/interviewing/etc), skip auto-set.
	}

	// Business rule: cannot move backwards to Interested if already Applied or later.
	if status == domain.StatusInterested && !notFound {
		pipeline := []domain.ApplicationStatus{
			domain.StatusInterested,
			domain.StatusApplied,
			domain.StatusInterviewing,
			domain.StatusOffer,
			domain.StatusRejected,
			domain.StatusWithdrawn,
		}
		existingRank := rankOf(existing.Status, pipeline)
		targetRank := rankOf(status, pipeline)
		if existingRank >= rankOf(domain.StatusApplied, pipeline) && targetRank <= existingRank {
			// Already applied or later — no-op.
			return nil
		}
	}

	// Preserve existing notes on status-only transitions.
	notes := ""
	if !notFound {
		notes = existing.Notes
	}

	userJob := domain.UserJob{
		UserID:   userID,
		JobID:    jobID,
		Status:   status,
		StatusAt: time.Now(),
		Notes:    notes,
	}
	if err := s.userJobs.Upsert(ctx, userJob); err != nil {
		return fmt.Errorf("set status %s for job %s: %w", status, jobID, err)
	}
	return nil
}

// SetNotes updates the notes field for a tracked job.
//
// Reads the existing UserJob first (read-modify-write) to preserve status.
// If the job has not been tracked yet, it is implicitly set to StatusInterested.
func (s *applicationService) SetNotes(ctx context.Context, userID domain.UserID, jobID domain.JobID, notes string) error {
	existing, err := s.userJobs.GetUserJob(ctx, userID, jobID)
	if err != nil {
		// Not tracked yet — implicitly track as interested.
		existing = domain.UserJob{
			UserID:   userID,
			JobID:    jobID,
			Status:   domain.StatusInterested,
			StatusAt: time.Now(),
		}
	}
	existing.Notes = notes
	if err := s.userJobs.Upsert(ctx, existing); err != nil {
		return fmt.Errorf("set notes for job %s: %w", jobID, err)
	}
	return nil
}

// GetUserJob returns a user's pipeline state for a specific job.
func (s *applicationService) GetUserJob(ctx context.Context, userID domain.UserID, jobID domain.JobID) (domain.UserJob, error) {
	uj, err := s.userJobs.GetUserJob(ctx, userID, jobID)
	if err != nil {
		return domain.UserJob{}, fmt.Errorf("get user job (user=%s, job=%s): %w", userID, jobID, err)
	}
	return uj, nil
}

// ListByStatus returns full Job values for a user's jobs in a given status,
// hydrated from the job repository.
func (s *applicationService) ListByStatus(ctx context.Context, userID domain.UserID, status domain.ApplicationStatus) ([]domain.Job, error) {
	ids, err := s.userJobs.ListByStatus(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("list by status %s: %w", status, err)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	jobs, err := s.jobs.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("hydrate jobs for status %s: %w", status, err)
	}
	return jobs, nil
}

// ListPipeline returns all tracked jobs for a user, grouped by status.
func (s *applicationService) ListPipeline(ctx context.Context, userID domain.UserID) (map[domain.ApplicationStatus][]domain.Job, error) {
	userJobs, err := s.userJobs.ListAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list pipeline: %w", err)
	}
	if len(userJobs) == 0 {
		return map[domain.ApplicationStatus][]domain.Job{}, nil
	}

	// Collect all job IDs for a single batch fetch.
	ids := make([]domain.JobID, len(userJobs))
	for i, uj := range userJobs {
		ids[i] = uj.JobID
	}
	jobSlice, err := s.jobs.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("hydrate pipeline jobs: %w", err)
	}

	// Index jobs by ID.
	byID := make(map[domain.JobID]domain.Job, len(jobSlice))
	for _, j := range jobSlice {
		byID[j.ID] = j
	}

	// Group by status.
	result := make(map[domain.ApplicationStatus][]domain.Job)
	for _, uj := range userJobs {
		j, ok := byID[uj.JobID]
		if !ok {
			continue // job deleted from DB — skip
		}
		result[uj.Status] = append(result[uj.Status], j)
	}
	return result, nil
}

// ResetTracker deletes all pipeline state (user_jobs) for the user.
func (s *applicationService) ResetTracker(ctx context.Context, userID domain.UserID) error {
	if err := s.userJobs.DeleteAll(ctx, userID); err != nil {
		return fmt.Errorf("reset tracker for user %s: %w", userID, err)
	}
	return nil
}

// rankOf returns the position of status in the ordered pipeline slice,
// or -1 if not found.
func rankOf(status domain.ApplicationStatus, order []domain.ApplicationStatus) int {
	for i, s := range order {
		if s == status {
			return i
		}
	}
	return -1
}
