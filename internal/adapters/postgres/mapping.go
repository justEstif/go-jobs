package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// uuidToPg converts a google/uuid to pgtype.UUID.
func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// pgToUUID converts a pgtype.UUID to google/uuid.UUID.
// Returns uuid.Nil if the pgtype.UUID is not valid.
func pgToUUID(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return id.Bytes
}

// timeToPg converts a time.Time to pgtype.Timestamptz.
func timeToPg(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// pgToTime converts a pgtype.Timestamptz to time.Time.
// Returns zero time if the pgtype value is not valid.
func pgToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// pgToTimePtr converts a pgtype.Timestamptz to *time.Time.
// Returns nil if the pgtype value is not valid.
func pgToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

// timePtrToPg converts a *time.Time to pgtype.Timestamptz.
func timePtrToPg(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// domainCompanyFromDB maps a queries.Company to domain.Company.
func domainCompanyFromDB(c queries.Company) domain.Company {
	return domain.Company{
		ID:         pgToUUID(c.ID),
		Name:       c.Name,
		CareersURL: c.CareersUrl,
		ATSType:    domain.ATSType(c.AtsType),
		ScrapeType: domain.ScrapeType(c.ScrapeType),
		BoardToken: c.BoardToken,
		Active:     c.Active,
		CreatedAt:  pgToTime(c.CreatedAt),
	}
}

// domainJobFromJobRow maps a queries.GetJobByIDRow to domain.Job.
func domainJobFromJobRow(r queries.GetJobByIDRow) domain.Job {
	return domain.Job{
		ID:          pgToUUID(r.ID),
		CompanyID:   pgToUUID(r.CompanyID),
		CompanyName: r.CompanyName,
		ExternalID:  r.ExternalID,
		Title:       r.Title,
		URL:         r.Url,
		Location:    r.Location,
		Description: r.Description,
		RawHTML:     r.RawHtml,
		Source:      domain.ATSType(r.Source),
		FirstSeen:   pgToTime(r.FirstSeen),
		LastSeen:    pgToTime(r.LastSeen),
		Active:      r.Active,
	}
}

// domainJobFromGetByIDsRow maps a queries.GetJobsByIDsRow to domain.Job.
func domainJobFromGetByIDsRow(r queries.GetJobsByIDsRow) domain.Job {
	return domain.Job{
		ID:          pgToUUID(r.ID),
		CompanyID:   pgToUUID(r.CompanyID),
		CompanyName: r.CompanyName,
		ExternalID:  r.ExternalID,
		Title:       r.Title,
		URL:         r.Url,
		Location:    r.Location,
		Description: r.Description,
		RawHTML:     r.RawHtml,
		Source:      domain.ATSType(r.Source),
		FirstSeen:   pgToTime(r.FirstSeen),
		LastSeen:    pgToTime(r.LastSeen),
		Active:      r.Active,
	}
}

// domainJobFromUnenrichedRow maps a queries.ListUnenrichedJobsRow to domain.Job.
func domainJobFromUnenrichedRow(r queries.ListUnenrichedJobsRow) domain.Job {
	return domain.Job{
		ID:          pgToUUID(r.ID),
		CompanyID:   pgToUUID(r.CompanyID),
		CompanyName: r.CompanyName,
		ExternalID:  r.ExternalID,
		Title:       r.Title,
		URL:         r.Url,
		Location:    r.Location,
		Description: r.Description,
		RawHTML:     r.RawHtml,
		Source:      domain.ATSType(r.Source),
		FirstSeen:   pgToTime(r.FirstSeen),
		LastSeen:    pgToTime(r.LastSeen),
		Active:      r.Active,
	}
}

// domainUserFromDB maps a queries.User to domain.User.
func domainUserFromDB(u queries.User) domain.User {
	return domain.User{
		ID:            pgToUUID(u.ID),
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		LLMAPIKey:     u.LlmApiKey,
		LLMProvider:   domain.LLMProvider(u.LlmProvider),
		LastVisitedAt: pgToTimePtr(u.LastVisitedAt),
		CreatedAt:     pgToTime(u.CreatedAt),
	}
}

// domainUserJobFromDB maps a queries.UserJob to domain.UserJob.
func domainUserJobFromDB(uj queries.UserJob) domain.UserJob {
	return domain.UserJob{
		UserID:    pgToUUID(uj.UserID),
		JobID:     pgToUUID(uj.JobID),
		Status:    domain.ApplicationStatus(uj.Status),
		StatusAt:  pgToTime(uj.StatusAt),
		AppliedAt: pgToTimePtr(uj.AppliedAt),
		Notes:     uj.Notes,
	}
}

// domainScrapeRunFromDB maps a queries.ScrapeRun to domain.ScrapeRun.
func domainScrapeRunFromDB(r queries.ScrapeRun) domain.ScrapeRun {
	return domain.ScrapeRun{
		ID:          pgToUUID(r.ID),
		StartedAt:   pgToTime(r.StartedAt),
		FinishedAt:  pgToTimePtr(r.FinishedAt),
		Status:      domain.ScrapeStatus(r.Status),
		JobsAdded:   int(r.JobsAdded),
		JobsUpdated: int(r.JobsUpdated),
		JobsRemoved: int(r.JobsRemoved),
		Error:       r.Error,
	}
}
