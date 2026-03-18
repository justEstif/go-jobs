package scrapers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// AshbyAdapter implements ports.JobScraper for the Ashby ATS.
//
// Ashby exposes a public posting API:
//
//	POST https://api.ashbyhq.com/posting-api/job-board/{board_name}
//	Body: {"includeCompensation": false}
//
// The board_name is the identifier from the Ashby careers URL.
type AshbyAdapter struct {
	client *http.Client
}

// NewAshbyAdapter constructs an AshbyAdapter with a conservative timeout.
func NewAshbyAdapter() *AshbyAdapter {
	return &AshbyAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Scrape fetches all open jobs for company from the Ashby job board API.
func (a *AshbyAdapter) Scrape(ctx context.Context, company domain.Company) ([]domain.RawJob, error) {
	url := fmt.Sprintf("https://api.ashbyhq.com/posting-api/job-board/%s", company.BoardToken)

	body, _ := json.Marshal(map[string]bool{"includeCompensation": false})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ashby build request for %s: %w", company.BoardToken, err)
	}
	req.Header.Set("Content-Type", "application/json")
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

	var payload struct {
		Jobs []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			PublishedAt string `json:"publishedAt"`
			Location    struct {
				City    string `json:"city"`
				Region  string `json:"region"`
				Country string `json:"country"`
			} `json:"location"`
			LocationType   string `json:"locationRequirement"` // "RemoteGlobal" | "RemoteCountry" | "Onsite" | "Hybrid"
			Department     string `json:"departmentName"`
			JobURL         string `json:"jobUrl"`
			Description    string `json:"descriptionHtml"`
			EmploymentType string `json:"employmentType"` // "FullTime" | "PartTime" | "Contract" | "Intern"
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("ashby decode %s: %w", company.BoardToken, err)
	}

	jobs := make([]domain.RawJob, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		location := buildAshbyLocation(j.Location.City, j.Location.Region, j.Location.Country)

		raw := domain.RawJob{
			ExternalID:     j.ID,
			Title:          j.Title,
			URL:            j.JobURL,
			Location:       location,
			Description:    stripHTML(j.Description),
			RawHTML:        j.Description,
			FirstSeen:      time.Now(),
			Department:     j.Department,
			Country:        strings.ToUpper(j.Location.Country),
			WorkplaceType:  ashbyWorkplaceType(j.LocationType),
			EmploymentType: ashbyEmploymentType(j.EmploymentType),
		}

		if t, err := time.Parse(time.RFC3339, j.PublishedAt); err == nil {
			raw.FirstSeen = t
		}

		jobs = append(jobs, raw)
	}

	return jobs, nil
}

// buildAshbyLocation assembles a human-readable location string from components.
func buildAshbyLocation(city, region, country string) string {
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

// ashbyWorkplaceType maps Ashby's locationRequirement to domain.WorkplaceType.
func ashbyWorkplaceType(s string) domain.WorkplaceType {
	switch s {
	case "RemoteGlobal", "RemoteCountry", "Remote":
		return domain.WorkplaceRemote
	case "Hybrid":
		return domain.WorkplaceHybrid
	case "Onsite", "OnSite":
		return domain.WorkplaceOnsite
	default:
		return ""
	}
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
