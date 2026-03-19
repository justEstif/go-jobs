package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// JobRepo implements ports.JobRepository against PostgreSQL.
type JobRepo struct {
	q *queries.Queries
}

// NewJobRepo constructs a JobRepo backed by the given Queries handle.
func NewJobRepo(q *queries.Queries) *JobRepo {
	return &JobRepo{q: q}
}

// Upsert inserts or updates a job by (company_id, external_id).
// source identifies which ATS platform the job came from.
func (r *JobRepo) Upsert(ctx context.Context, companyID domain.CompanyID, raw domain.RawJob, source domain.ATSType) (domain.JobID, error) {
	return r.upsertWithSource(ctx, companyID, raw, source)
}

// upsertWithSource is the internal implementation used by Upsert.
func (r *JobRepo) upsertWithSource(ctx context.Context, companyID domain.CompanyID, raw domain.RawJob, atsType domain.ATSType) (domain.JobID, error) {
	firstSeen := raw.FirstSeen
	if firstSeen.IsZero() {
		firstSeen = time.Now()
	}
	row, err := r.q.UpsertJob(ctx, queries.UpsertJobParams{
		CompanyID:   uuidToPg(companyID),
		ExternalID:  raw.ExternalID,
		Title:       raw.Title,
		Url:         raw.URL,
		Location:    raw.Location,
		Description: raw.Description,
		RawHtml:     raw.RawHTML,
		Source:      string(atsType),
		FirstSeen:   timeToPg(firstSeen),
		LastSeen:    timeToPg(time.Now()),
	})
	if err != nil {
		return domain.JobID{}, fmt.Errorf("upsert job %q for company %s: %w", raw.ExternalID, companyID, err)
	}
	return pgToUUID(row.ID), nil
}

// GetByID retrieves a job by its primary key, including company name.

func (r *JobRepo) GetByID(ctx context.Context, id domain.JobID) (domain.Job, error) {
	row, err := r.q.GetJobByID(ctx, uuidToPg(id))
	if err != nil {
		return domain.Job{}, fmt.Errorf("get job by id %s: %w", id, err)
	}
	return domainJobFromJobRow(row), nil
}

// GetByIDs fetches multiple jobs by ID in a single query.
func (r *JobRepo) GetByIDs(ctx context.Context, ids []domain.JobID) ([]domain.Job, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = uuidToPg(id)
	}
	rows, err := r.q.GetJobsByIDs(ctx, pgIDs)
	if err != nil {
		return nil, fmt.Errorf("get jobs by ids: %w", err)
	}
	jobs := make([]domain.Job, len(rows))
	for i, row := range rows {
		jobs[i] = domainJobFromGetByIDsRow(row)
	}
	return jobs, nil
}

// Search returns jobs matching filters.
//
// Applies multi-dimensional filtering via the SearchJobs query. All slice
// fields use OR semantics; TechStack uses AND semantics. userCtx is reserved
// for future user-context features (not yet used).
func (r *JobRepo) Search(ctx context.Context, filters domain.SearchFilters, userCtx *domain.UserSearchContext) ([]domain.Job, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	// Convert domain types to the query parameter types.
	// Use nil (not empty slice) so array_length($n, 1) IS NULL disables the filter.
	var roleTypes []string
	for _, rt := range filters.RoleTypes {
		roleTypes = append(roleTypes, string(rt))
	}
	var seniorities []string
	for _, s := range filters.Seniorities {
		seniorities = append(seniorities, string(s))
	}
	var remotePolicies []string
	for _, rp := range filters.RemotePolicy {
		remotePolicies = append(remotePolicies, string(rp))
	}
	var companyIDs []pgtype.UUID
	for _, id := range filters.CompanyIDs {
		companyIDs = append(companyIDs, uuidToPg(id))
	}

	rows, err := r.q.SearchJobs(ctx, queries.SearchJobsParams{
		Column1: filters.Query,
		Column2: roleTypes,
		Column3: seniorities,
		Column4: remotePolicies,
		Column5: filters.Countries,
		Column6: filters.TechStack,
		Column7: companyIDs,
		Column8: int32(filters.PostedWithinDays),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("search jobs: %w", err)
	}
	jobs := make([]domain.Job, len(rows))
	for i, row := range rows {
		jobs[i] = domainJobFromSearchRow(row)
	}
	return jobs, nil
}

// ListUnenriched returns jobs without a job_tags row, up to limit.
func (r *JobRepo) ListUnenriched(ctx context.Context, limit int) ([]domain.Job, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.q.ListUnenrichedJobs(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("list unenriched jobs: %w", err)
	}
	jobs := make([]domain.Job, len(rows))
	for i, row := range rows {
		jobs[i] = domainJobFromUnenrichedRow(row)
	}
	return jobs, nil
}

// MarkInactive sets active=false for jobs no longer present at source.
func (r *JobRepo) MarkInactive(ctx context.Context, companyID domain.CompanyID, activeExternalIDs []string) (int, error) {
	rows, err := r.q.MarkJobsInactive(ctx, queries.MarkJobsInactiveParams{
		CompanyID: uuidToPg(companyID),
		Column2:   activeExternalIDs,
	})
	if err != nil {
		return 0, fmt.Errorf("mark inactive jobs for company %s: %w", companyID, err)
	}
	return int(rows), nil
}

// SaveTags upserts enrichment tags for a job.
func (r *JobRepo) SaveTags(ctx context.Context, tags domain.JobTags) error {
	err := r.q.UpsertJobTags(ctx, queries.UpsertJobTagsParams{
		JobID:            uuidToPg(tags.JobID),
		RoleType:         string(tags.RoleType),
		Seniority:        string(tags.Seniority),
		RemotePolicy:     string(tags.RemotePolicy),
		LocationNorm:     tags.LocationNorm,
		Country:          tags.Country,
		TechStack:        tags.TechStack,
		EnrichmentSource: string(tags.EnrichmentSource),
		EnrichedAt:       timeToPg(tags.EnrichedAt),
	})
	if err != nil {
		return fmt.Errorf("save tags for job %s: %w", tags.JobID, err)
	}
	return nil
}
