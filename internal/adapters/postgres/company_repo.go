package postgres

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// CompanyRepo implements ports.CompanyRepository against PostgreSQL.
type CompanyRepo struct {
	q *queries.Queries
}

// NewCompanyRepo constructs a CompanyRepo backed by the given Queries handle.
func NewCompanyRepo(q *queries.Queries) *CompanyRepo {
	return &CompanyRepo{q: q}
}

// Upsert inserts or updates a company by (ats_type, board_token).
func (r *CompanyRepo) Upsert(ctx context.Context, company domain.Company) (domain.CompanyID, error) {
	normalized := company.NormalizedName
	if normalized == "" {
		normalized = domain.NormalizeCompanyName(company.Name)
	}
	row, err := r.q.UpsertCompany(ctx, queries.UpsertCompanyParams{
		Name:           company.Name,
		CareersUrl:     company.CareersURL,
		AtsType:        string(company.ATSType),
		ScrapeType:     string(company.ScrapeType),
		BoardToken:     company.BoardToken,
		Active:         company.Active,
		NormalizedName: normalized,
	})
	if err != nil {
		return domain.CompanyID{}, fmt.Errorf("upsert company %q: %w", company.Name, err)
	}
	return pgToUUID(row.ID), nil
}

// ListActive returns all companies with active=true, ordered by name.
func (r *CompanyRepo) ListActive(ctx context.Context) ([]domain.Company, error) {
	rows, err := r.q.ListActiveCompanies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active companies: %w", err)
	}
	companies := make([]domain.Company, len(rows))
	for i, row := range rows {
		companies[i] = domainCompanyFromDB(row)
	}
	return companies, nil
}

// GetByID retrieves a company by its primary key.
func (r *CompanyRepo) GetByID(ctx context.Context, id domain.CompanyID) (domain.Company, error) {
	row, err := r.q.GetCompanyByID(ctx, uuidToPg(id))
	if err != nil {
		return domain.Company{}, fmt.Errorf("get company by id %s: %w", id, err)
	}
	return domainCompanyFromDB(row), nil
}

// GetByBoardToken retrieves a company by (ats_type, board_token).
func (r *CompanyRepo) GetByBoardToken(ctx context.Context, atsType domain.ATSType, boardToken string) (domain.Company, error) {
	row, err := r.q.GetCompanyByBoardToken(ctx, queries.GetCompanyByBoardTokenParams{
		AtsType:    string(atsType),
		BoardToken: boardToken,
	})
	if err != nil {
		return domain.Company{}, fmt.Errorf("get company by board token %q (%s): %w", boardToken, atsType, err)
	}
	return domainCompanyFromDB(row), nil
}
