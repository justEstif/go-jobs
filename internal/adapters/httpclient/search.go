package httpclient

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// SearchClient implements ports.JobSearchService against the /api/v1/jobs endpoints.
type SearchClient struct {
	c *Client
}

// NewSearchClient constructs a SearchClient.
func NewSearchClient(c *Client) *SearchClient {
	return &SearchClient{c: c}
}

// Search calls GET /api/v1/jobs with filters mapped to query parameters.
// userCtx is used only to determine whether to send the Bearer token so that
// the server can annotate results with pipeline state. OnlyNew is not
// supported in remote mode.
func (s *SearchClient) Search(ctx context.Context, filters domain.SearchFilters, userCtx *domain.UserSearchContext) ([]domain.Job, error) {
	q := filtersToQuery(filters)
	bearer := userCtx != nil && s.c.token != ""

	var jobs []domain.Job
	if err := s.c.get(ctx, "jobs", q, bearer, &jobs); err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	if jobs == nil {
		jobs = []domain.Job{}
	}
	return jobs, nil
}

// GetByID is not currently exposed by the JSON API and is not used by the CLI.
func (s *SearchClient) GetByID(_ context.Context, _ domain.JobID) (domain.Job, error) {
	return domain.Job{}, fmt.Errorf("GetByID: not supported in remote mode")
}

// filtersToQuery converts domain.SearchFilters to URL query parameters that
// match the /api/v1/jobs query param schema.
func filtersToQuery(f domain.SearchFilters) url.Values {
	q := url.Values{}

	if f.Query != "" {
		q.Set("q", f.Query)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Offset > 0 {
		q.Set("offset", strconv.Itoa(f.Offset))
	}
	if f.PostedWithinDays > 0 {
		q.Set("posted_within", strconv.Itoa(f.PostedWithinDays)+"d")
	}
	for _, v := range f.RoleTypes {
		q.Add("role", string(v))
	}
	for _, v := range f.Seniorities {
		q.Add("seniority", string(v))
	}
	for _, v := range f.RemotePolicy {
		q.Add("remote", string(v))
	}
	for _, v := range f.Countries {
		q.Add("country", v)
	}
	for _, v := range f.TechStack {
		q.Add("tech", v)
	}
	for _, id := range f.CompanyIDs {
		q.Add("company", id.String())
	}

	return q
}
