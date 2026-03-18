package postgres

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// UserJobRepo implements ports.UserJobRepository against PostgreSQL.
type UserJobRepo struct {
	q *queries.Queries
}

// NewUserJobRepo constructs a UserJobRepo backed by the given Queries handle.
func NewUserJobRepo(q *queries.Queries) *UserJobRepo {
	return &UserJobRepo{q: q}
}

// Upsert writes all fields from userJob to the store.
//
// Applied_at is managed by the DB: set on first applied transition and never
// overwritten, regardless of the value in userJob.AppliedAt.
func (r *UserJobRepo) Upsert(ctx context.Context, userJob domain.UserJob) error {
	err := r.q.UpsertUserJob(ctx, queries.UpsertUserJobParams{
		UserID:   uuidToPg(userJob.UserID),
		JobID:    uuidToPg(userJob.JobID),
		Status:   string(userJob.Status),
		StatusAt: timeToPg(userJob.StatusAt),
		Column5:  timePtrToPg(userJob.AppliedAt), // applied_at — protected by DB logic
		Notes:    userJob.Notes,
	})
	if err != nil {
		return fmt.Errorf("upsert user_job (user=%s, job=%s): %w", userJob.UserID, userJob.JobID, err)
	}
	return nil
}

// GetUserJob retrieves a user's pipeline state for a specific job.
func (r *UserJobRepo) GetUserJob(ctx context.Context, userID domain.UserID, jobID domain.JobID) (domain.UserJob, error) {
	row, err := r.q.GetUserJob(ctx, queries.GetUserJobParams{
		UserID: uuidToPg(userID),
		JobID:  uuidToPg(jobID),
	})
	if err != nil {
		return domain.UserJob{}, fmt.Errorf("get user_job (user=%s, job=%s): %w", userID, jobID, err)
	}
	return domainUserJobFromDB(row), nil
}

// ListByStatus returns job IDs for a user's jobs in a given status.
func (r *UserJobRepo) ListByStatus(ctx context.Context, userID domain.UserID, status domain.ApplicationStatus) ([]domain.JobID, error) {
	rows, err := r.q.ListUserJobsByStatus(ctx, queries.ListUserJobsByStatusParams{
		UserID: uuidToPg(userID),
		Status: string(status),
	})
	if err != nil {
		return nil, fmt.Errorf("list user_jobs by status %s for user %s: %w", status, userID, err)
	}
	ids := make([]domain.JobID, len(rows))
	for i, row := range rows {
		ids[i] = pgToUUID(row)
	}
	return ids, nil
}

// ListAll returns all UserJob records for a user.
func (r *UserJobRepo) ListAll(ctx context.Context, userID domain.UserID) ([]domain.UserJob, error) {
	rows, err := r.q.ListAllUserJobs(ctx, uuidToPg(userID))
	if err != nil {
		return nil, fmt.Errorf("list all user_jobs for user %s: %w", userID, err)
	}
	jobs := make([]domain.UserJob, len(rows))
	for i, row := range rows {
		jobs[i] = domainUserJobFromDB(row)
	}
	return jobs, nil
}
