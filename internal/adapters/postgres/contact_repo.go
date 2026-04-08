package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// ContactRepo implements ports.ContactRepository and ports.CompanyMatcher.
type ContactRepo struct {
	q *queries.Queries
}

func NewContactRepo(q *queries.Queries) *ContactRepo {
	return &ContactRepo{q: q}
}

func (r *ContactRepo) Upsert(ctx context.Context, c domain.Contact) (domain.ContactID, error) {
	var connectedOn pgtype.Date
	if c.ConnectedOn != nil {
		connectedOn = pgtype.Date{Time: *c.ConnectedOn, Valid: true}
	}
	var companyID pgtype.UUID
	if c.CompanyID != nil {
		companyID = uuidToPg(*c.CompanyID)
	}

	row, err := r.q.UpsertContact(ctx, queries.UpsertContactParams{
		UserID:                uuidToPg(c.UserID),
		FirstName:             c.FirstName,
		LastName:              c.LastName,
		Email:                 c.Email,
		Title:                 c.Title,
		LinkedinUrl:           c.LinkedInURL,
		ConnectedOn:           connectedOn,
		CompanyName:           c.CompanyName,
		NormalizedCompanyName: c.NormalizedCompanyName,
		CompanyID:             companyID,
	})
	if err != nil {
		return domain.ContactID{}, fmt.Errorf("upsert contact %q: %w", c.FirstName+" "+c.LastName, err)
	}
	return pgToUUID(row.ID), nil
}

func (r *ContactRepo) LinkToCompany(ctx context.Context, normalizedCompanyName string, companyID domain.CompanyID) (int64, error) {
	n, err := r.q.LinkContactsToCompany(ctx, queries.LinkContactsToCompanyParams{
		CompanyID:             uuidToPg(companyID),
		NormalizedCompanyName: normalizedCompanyName,
	})
	if err != nil {
		return 0, fmt.Errorf("link contacts to company: %w", err)
	}
	return n, nil
}

func (r *ContactRepo) ListByCompanyID(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) ([]domain.Contact, error) {
	rows, err := r.q.ListContactsByCompanyID(ctx, queries.ListContactsByCompanyIDParams{
		UserID:    uuidToPg(userID),
		CompanyID: uuidToPg(companyID),
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts by company: %w", err)
	}
	return mapContacts(rows), nil
}

func (r *ContactRepo) ListByCompanyIDs(ctx context.Context, userID domain.UserID, companyIDs []domain.CompanyID) ([]domain.Contact, error) {
	pgIDs := make([]pgtype.UUID, len(companyIDs))
	for i, id := range companyIDs {
		pgIDs[i] = uuidToPg(id)
	}
	rows, err := r.q.ListContactsByCompanyIDs(ctx, queries.ListContactsByCompanyIDsParams{
		UserID: uuidToPg(userID),
		Column2: pgIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts by company IDs: %w", err)
	}
	return mapContacts(rows), nil
}

func (r *ContactRepo) DeleteAllForUser(ctx context.Context, userID domain.UserID) error {
	return r.q.DeleteContactsForUser(ctx, uuidToPg(userID))
}

func (r *ContactRepo) CountForUser(ctx context.Context, userID domain.UserID) (int64, error) {
	return r.q.CountContactsForUser(ctx, uuidToPg(userID))
}

func (r *ContactRepo) CountLinkedForUser(ctx context.Context, userID domain.UserID) (int64, error) {
	return r.q.CountLinkedContactsForUser(ctx, uuidToPg(userID))
}

func (r *ContactRepo) CountDistinctCompaniesForUser(ctx context.Context, userID domain.UserID) (int64, error) {
	return r.q.CountDistinctCompaniesForUser(ctx, uuidToPg(userID))
}

func (r *ContactRepo) ListLinkedCompanyIDs(ctx context.Context, userID domain.UserID) ([]domain.CompanyID, error) {
	pgIDs, err := r.q.ListLinkedCompanyIDsForUser(ctx, uuidToPg(userID))
	if err != nil {
		return nil, fmt.Errorf("list linked company IDs: %w", err)
	}
	ids := make([]domain.CompanyID, len(pgIDs))
	for i, pgID := range pgIDs {
		ids[i] = pgToUUID(pgID)
	}
	return ids, nil
}

func (r *ContactRepo) ListUnlinkedCompanyNames(ctx context.Context, userID domain.UserID) ([]string, error) {
	return r.q.ListUnlinkedCompanyNames(ctx, uuidToPg(userID))
}

// --- CompanyMatcher ---

func (r *ContactRepo) GetByNormalizedName(ctx context.Context, normalizedName string) (domain.Company, error) {
	row, err := r.q.GetCompanyByNormalizedName(ctx, normalizedName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Company{}, pgx.ErrNoRows
		}
		return domain.Company{}, fmt.Errorf("get company by normalized name: %w", err)
	}
	return domainCompanyFromDB(row), nil
}

func (r *ContactRepo) FuzzyMatch(ctx context.Context, normalizedName string) (domain.Company, float64, error) {
	row, err := r.q.FuzzyMatchCompany(ctx, normalizedName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Company{}, 0, pgx.ErrNoRows
		}
		return domain.Company{}, 0, fmt.Errorf("fuzzy match company: %w", err)
	}
	company := domain.Company{
		ID:             pgToUUID(row.ID),
		Name:           row.Name,
		NormalizedName: row.NormalizedName,
		CareersURL:     row.CareersUrl,
		ATSType:        domain.ATSType(row.AtsType),
		ScrapeType:     domain.ScrapeType(row.ScrapeType),
		BoardToken:     row.BoardToken,
		Active:         row.Active,
		CreatedAt:      pgToTime(row.CreatedAt),
	}
	score := float64(row.Score)
	return company, score, nil
}

// mapContacts converts a slice of queries.Contact to domain.Contact.
func mapContacts(rows []queries.Contact) []domain.Contact {
	contacts := make([]domain.Contact, len(rows))
	for i, row := range rows {
		contacts[i] = domainContactFromDB(row)
	}
	return contacts
}

func domainContactFromDB(c queries.Contact) domain.Contact {
	contact := domain.Contact{
		ID:                    pgToUUID(c.ID),
		UserID:                pgToUUID(c.UserID),
		FirstName:             c.FirstName,
		LastName:              c.LastName,
		FullName:              c.FullName.String,
		Email:                 c.Email,
		Title:                 c.Title,
		LinkedInURL:           c.LinkedinUrl,
		CompanyName:           c.CompanyName,
		NormalizedCompanyName: c.NormalizedCompanyName,
		CreatedAt:             pgToTime(c.CreatedAt),
	}
	if c.ConnectedOn.Valid {
		t := c.ConnectedOn.Time
		contact.ConnectedOn = &t
	}
	if c.CompanyID.Valid {
		id := pgToUUID(c.CompanyID)
		contact.CompanyID = &id
	}
	return contact
}
