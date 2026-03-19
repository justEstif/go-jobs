package postgres

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// UserCompanyRepo implements ports.UserCompanyRepository against PostgreSQL.
type UserCompanyRepo struct {
	q *queries.Queries
}

// NewUserCompanyRepo constructs a UserCompanyRepo backed by the given Queries handle.
func NewUserCompanyRepo(q *queries.Queries) *UserCompanyRepo {
	return &UserCompanyRepo{q: q}
}

// SetHidden upserts the hidden preference for (user, company).
func (r *UserCompanyRepo) SetHidden(ctx context.Context, userID domain.UserID, companyID domain.CompanyID, hidden bool) error {
	err := r.q.SetUserCompanyHidden(ctx, queries.SetUserCompanyHiddenParams{
		UserID:    uuidToPg(userID),
		CompanyID: uuidToPg(companyID),
		Hidden:    hidden,
	})
	if err != nil {
		return fmt.Errorf("set hidden=%v for user %s company %s: %w", hidden, userID, companyID, err)
	}
	return nil
}

// IsHidden returns true if the user has hidden the specified company.
// Returns false if no row exists for the (user, company) pair.
func (r *UserCompanyRepo) IsHidden(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) (bool, error) {
	hidden, err := r.q.IsCompanyHidden(ctx, queries.IsCompanyHiddenParams{
		UserID:    uuidToPg(userID),
		CompanyID: uuidToPg(companyID),
	})
	if err != nil {
		return false, fmt.Errorf("check hidden for user %s company %s: %w", userID, companyID, err)
	}
	return hidden, nil
}

// ListHidden returns the IDs of companies hidden by the given user.
func (r *UserCompanyRepo) ListHidden(ctx context.Context, userID domain.UserID) ([]domain.CompanyID, error) {
	rows, err := r.q.ListHiddenCompanies(ctx, uuidToPg(userID))
	if err != nil {
		return nil, fmt.Errorf("list hidden companies for user %s: %w", userID, err)
	}
	ids := make([]domain.CompanyID, len(rows))
	for i, row := range rows {
		ids[i] = pgToUUID(row)
	}
	return ids, nil
}
