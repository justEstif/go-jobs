package scrapers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-json-experiment/json"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// LeverAdapter implements ports.JobScraper for the Lever ATS.
//
// Lever exposes a public v0 API:
//
//	GET https://api.lever.co/v0/postings/{company_slug}?mode=json
//
// The company_slug is the identifier from the Lever careers URL.
type LeverAdapter struct {
	client *http.Client
}

// NewLeverAdapter constructs a LeverAdapter with a conservative timeout.
func NewLeverAdapter() *LeverAdapter {
	return &LeverAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// leverPosting is a single job posting returned by the Lever v0 API.
type leverPosting struct {
	ID         string `json:"id"`
	Text       string `json:"text"` // job title
	HostedURL  string `json:"hostedUrl"`
	Categories struct {
		Team          string `json:"team"`
		Department    string `json:"department"`
		Location      string `json:"location"`
		WorkplaceType string `json:"workplaceType"`
	} `json:"categories"`
	DescriptionPlain string `json:"descriptionPlain"`
	Description      string `json:"description"` // HTML fallback
	CreatedAt        int64  `json:"createdAt"`   // unix ms
}

// Scrape fetches all open jobs for company from the Lever postings API.
func (a *LeverAdapter) Scrape(ctx context.Context, company domain.Company) ([]domain.RawJob, error) {
	url := fmt.Sprintf("https://api.lever.co/v0/postings/%s?mode=json", company.BoardToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("lever build request for %s: %w", company.BoardToken, err)
	}
	req.Header.Set("User-Agent", "go-jobs/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lever fetch %s: %w", company.BoardToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []domain.RawJob{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lever %s returned HTTP %d", company.BoardToken, resp.StatusCode)
	}

	var postings []leverPosting
	// json v2 UnmarshalRead: reads until EOF, case-sensitive field matching by default.
	// Lever returns camelCase keys matching our struct tags exactly.
	if err := json.UnmarshalRead(resp.Body, &postings); err != nil {
		return nil, fmt.Errorf("lever decode %s: %w", company.BoardToken, err)
	}

	jobs := make([]domain.RawJob, 0, len(postings))
	for _, j := range postings {
		desc := j.DescriptionPlain
		if desc == "" {
			desc = stripHTML(j.Description)
		}

		raw := domain.RawJob{
			ExternalID:  j.ID,
			Title:       j.Text,
			URL:         j.HostedURL,
			Location:    j.Categories.Location,
			Description: desc,
			RawHTML:     j.Description,
			FirstSeen:   time.Now(),
		}

		if j.Categories.Department != "" {
			raw.Department = j.Categories.Department
		} else if j.Categories.Team != "" {
			raw.Department = j.Categories.Team
		}

		if wt := leverWorkplaceType(j.Categories.WorkplaceType); wt != "" {
			raw.WorkplaceType = wt
		}

		if j.CreatedAt > 0 {
			raw.FirstSeen = time.UnixMilli(j.CreatedAt)
		}

		jobs = append(jobs, raw)
	}

	return jobs, nil
}

// leverWorkplaceType maps Lever's workplace type string to domain.WorkplaceType.
func leverWorkplaceType(s string) domain.WorkplaceType {
	switch s {
	case "remote", "Remote":
		return domain.WorkplaceRemote
	case "hybrid", "Hybrid":
		return domain.WorkplaceHybrid
	case "on-site", "On-site", "onsite", "Onsite":
		return domain.WorkplaceOnsite
	default:
		return ""
	}
}
