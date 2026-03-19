package services

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// companyService implements ports.CompanyService.
type companyService struct {
	companies     ports.CompanyRepository
	userCompanies ports.UserCompanyRepository
}

// NewCompanyService constructs a CompanyService.
func NewCompanyService(companies ports.CompanyRepository, userCompanies ports.UserCompanyRepository) ports.CompanyService {
	return &companyService{
		companies:     companies,
		userCompanies: userCompanies,
	}
}

// HideCompany marks a company as hidden for the given user.
func (s *companyService) HideCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error {
	if err := s.userCompanies.SetHidden(ctx, userID, companyID, true); err != nil {
		return fmt.Errorf("hide company %s for user %s: %w", companyID, userID, err)
	}
	return nil
}

// ShowCompany removes the hidden flag for a company for the given user.
func (s *companyService) ShowCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) error {
	if err := s.userCompanies.SetHidden(ctx, userID, companyID, false); err != nil {
		return fmt.Errorf("show company %s for user %s: %w", companyID, userID, err)
	}
	return nil
}

// ListCompanies returns all active companies that the user has not hidden.
func (s *companyService) ListCompanies(ctx context.Context, userID domain.UserID) ([]domain.Company, error) {
	all, err := s.companies.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	hiddenIDs, err := s.userCompanies.ListHidden(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list hidden companies: %w", err)
	}
	hiddenSet := make(map[domain.CompanyID]bool, len(hiddenIDs))
	for _, id := range hiddenIDs {
		hiddenSet[id] = true
	}
	visible := all[:0]
	for _, c := range all {
		if !hiddenSet[c.ID] {
			visible = append(visible, c)
		}
	}
	return visible, nil
}

// ListHiddenCompanies returns the full Company objects for companies the user
// has hidden. Excludes companies with placeholder names (just the ATS provider
// name, e.g. "greenhouse").
func (s *companyService) ListHiddenCompanies(ctx context.Context, userID domain.UserID) ([]domain.Company, error) {
	hiddenIDs, err := s.userCompanies.ListHidden(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list hidden company IDs: %w", err)
	}
	if len(hiddenIDs) == 0 {
		return nil, nil
	}

	all, err := s.companies.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active companies: %w", err)
	}

	hiddenSet := make(map[domain.CompanyID]bool, len(hiddenIDs))
	for _, id := range hiddenIDs {
		hiddenSet[id] = true
	}

	var result []domain.Company
	for _, c := range all {
		if hiddenSet[c.ID] && !c.IsPlaceholderName() {
			result = append(result, c)
		}
	}
	return result, nil
}
