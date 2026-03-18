package scrapers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-json-experiment/json"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// AshbyAdapter implements ports.JobScraper for the Ashby ATS.
//
// Ashby exposes a public GET endpoint:
//
//	GET https://api.ashbyhq.com/posting-api/job-board/{board_name}
//
// The board_name is the identifier from the Ashby careers URL.
// Note: the older POST form of this endpoint now returns 401 for many boards.
type AshbyAdapter struct {
	client *http.Client
}

// NewAshbyAdapter constructs an AshbyAdapter with a conservative timeout.
func NewAshbyAdapter() *AshbyAdapter {
	return &AshbyAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ashbyPayload is the top-level response from the Ashby job board API.
type ashbyPayload struct {
	Jobs []struct {
		ID             string `json:"id"`
		Title          string `json:"title"`
		Department     string `json:"department"`
		EmploymentType string `json:"employmentType"`
		Location       string `json:"location"`
		PublishedAt    string `json:"publishedAt"`
		IsRemote       bool   `json:"isRemote"`
		WorkplaceType  string `json:"workplaceType"` // "Remote" | "Hybrid" | "OnSite"
		Address        struct {
			PostalAddress struct {
				Region  string `json:"addressRegion"`
				Country string `json:"addressCountry"`
				City    string `json:"addressLocality"`
			} `json:"postalAddress"`
		} `json:"address"`
		JobURL           string `json:"jobUrl"`
		DescriptionHTML  string `json:"descriptionHtml"`
		DescriptionPlain string `json:"descriptionPlain"`
	} `json:"jobs"`
}

// Scrape fetches all open jobs for company from the Ashby job board API.
func (a *AshbyAdapter) Scrape(ctx context.Context, company domain.Company) ([]domain.RawJob, error) {
	url := fmt.Sprintf("https://api.ashbyhq.com/posting-api/job-board/%s", company.BoardToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ashby build request for %s: %w", company.BoardToken, err)
	}
	req.Header.Set("User-Agent", "go-jobs/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ashby fetch %s: %w", company.BoardToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []domain.RawJob{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ashby %s returned HTTP %d", company.BoardToken, resp.StatusCode)
	}

	var payload ashbyPayload
	if err := json.UnmarshalRead(resp.Body, &payload); err != nil {
		return nil, fmt.Errorf("ashby decode %s: %w", company.BoardToken, err)
	}

	jobs := make([]domain.RawJob, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		// Prefer descriptionPlain; fall back to stripping the HTML.
		desc := j.DescriptionPlain
		if desc == "" {
			desc = stripHTML(j.DescriptionHTML)
		}

		country := j.Address.PostalAddress.Country
		location := buildAshbyLocation(j.Location, j.Address.PostalAddress.City, j.Address.PostalAddress.Region, country)

		raw := domain.RawJob{
			ExternalID:     j.ID,
			Title:          j.Title,
			URL:            j.JobURL,
			Location:       location,
			Description:    desc,
			RawHTML:        j.DescriptionHTML,
			FirstSeen:      time.Now(),
			Department:     j.Department,
			Country:        normaliseCountry(country),
			WorkplaceType:  ashbyWorkplaceType(j.WorkplaceType, j.IsRemote),
			EmploymentType: ashbyEmploymentType(j.EmploymentType),
		}

		if t, err := time.Parse(time.RFC3339, j.PublishedAt); err == nil {
			raw.FirstSeen = t
		}

		jobs = append(jobs, raw)
	}

	return jobs, nil
}

// buildAshbyLocation assembles a human-readable location string.
// Prefers the top-level location field (already formatted by Ashby),
// falling back to address components.
func buildAshbyLocation(topLevel, city, region, country string) string {
	if topLevel != "" {
		return topLevel
	}
	parts := []string{}
	if city != "" {
		parts = append(parts, city)
	}
	if region != "" {
		parts = append(parts, region)
	}
	if country != "" {
		parts = append(parts, country)
	}
	return strings.Join(parts, ", ")
}

// normaliseCountry normalises Ashby's country string to an ISO 2-letter code where possible.
// Ashby returns values like "USA", "United States", "Germany" etc.
func normaliseCountry(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "USA", "UNITED STATES", "UNITED STATES OF AMERICA", "US":
		return "US"
	case "UK", "UNITED KINGDOM", "GREAT BRITAIN", "GB":
		return "GB"
	case "CANADA", "CA":
		return "CA"
	case "GERMANY", "DE", "DEUTSCHLAND":
		return "DE"
	case "FRANCE", "FR":
		return "FR"
	case "INDIA", "IN":
		return "IN"
	case "AUSTRALIA", "AU":
		return "AU"
	case "BRAZIL", "BR":
		return "BR"
	default:
		// Return as-is if we don't recognise it; enrichment can normalise further.
		return s
	}
}

// ashbyWorkplaceType maps Ashby's workplaceType string to domain.WorkplaceType.
func ashbyWorkplaceType(wt string, isRemote bool) domain.WorkplaceType {
	switch strings.ToLower(wt) {
	case "remote":
		return domain.WorkplaceRemote
	case "hybrid":
		return domain.WorkplaceHybrid
	case "onsite", "on-site", "on_site":
		return domain.WorkplaceOnsite
	}
	// Fall back to isRemote flag.
	if isRemote {
		return domain.WorkplaceRemote
	}
	return ""
}

// ashbyEmploymentType maps Ashby's employmentType to domain.EmploymentType.
func ashbyEmploymentType(s string) domain.EmploymentType {
	switch s {
	case "FullTime":
		return domain.EmploymentFullTime
	case "PartTime":
		return domain.EmploymentPartTime
	case "Contract":
		return domain.EmploymentContract
	case "Intern":
		return domain.EmploymentIntern
	default:
		return ""
	}
}
