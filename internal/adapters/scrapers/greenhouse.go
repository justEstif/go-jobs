package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// GreenhouseAdapter implements ports.JobScraper for the Greenhouse ATS.
//
// Greenhouse exposes a public JSON API:
//
//	GET https://boards-api.greenhouse.io/v1/boards/{board_token}/jobs?content=true
//
// Each company's board_token is the slug from their Greenhouse careers URL.
type GreenhouseAdapter struct {
	client *http.Client
}

// NewGreenhouseAdapter constructs a GreenhouseAdapter with a conservative timeout.
func NewGreenhouseAdapter() *GreenhouseAdapter {
	return &GreenhouseAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// Scrape fetches all open jobs for company from the Greenhouse boards API.
// headless companies are the caller's responsibility to skip.
func (a *GreenhouseAdapter) Scrape(ctx context.Context, company domain.Company) ([]domain.RawJob, error) {
	url := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", company.BoardToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("greenhouse build request for %s: %w", company.BoardToken, err)
	}
	req.Header.Set("User-Agent", "go-jobs/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("greenhouse fetch %s: %w", company.BoardToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Board no longer exists — return empty, not an error
		return []domain.RawJob{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("greenhouse %s returned HTTP %d", company.BoardToken, resp.StatusCode)
	}

	var payload struct {
		Jobs []struct {
			ID        int    `json:"id"`
			Title     string `json:"title"`
			UpdatedAt string `json:"updated_at"`
			Location  struct {
				Name string `json:"name"`
			} `json:"location"`
			AbsoluteURL string `json:"absolute_url"`
			Content     string `json:"content"` // HTML description
			Metadata    []struct {
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"metadata"`
			Departments []struct {
				Name string `json:"name"`
			} `json:"departments"`
			Offices []struct {
				Name    string `json:"name"`
				Country string `json:"country_id"`
			} `json:"offices"`
		} `json:"jobs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("greenhouse decode %s: %w", company.BoardToken, err)
	}

	jobs := make([]domain.RawJob, 0, len(payload.Jobs))
	for _, j := range payload.Jobs {
		raw := domain.RawJob{
			ExternalID:  fmt.Sprintf("%d", j.ID),
			Title:       j.Title,
			URL:         j.AbsoluteURL,
			Location:    j.Location.Name,
			Description: stripHTML(j.Content),
			RawHTML:     j.Content,
			FirstSeen:   time.Now(),
		}

		// ATS metadata — department → role hint, offices → country
		if len(j.Departments) > 0 {
			raw.Department = j.Departments[0].Name
		}
		if len(j.Offices) > 0 && j.Offices[0].Country != "" {
			// Greenhouse country_id is a 2-letter ISO code
			raw.Country = strings.ToUpper(j.Offices[0].Country)
		}

		jobs = append(jobs, raw)
	}

	return jobs, nil
}

// stripHTML removes HTML tags from a string, returning plain text.
// This is a simple approach suitable for job descriptions.
func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	// Collapse multiple blank lines
	result := strings.ReplaceAll(b.String(), "\n\n\n", "\n\n")
	return strings.TrimSpace(result)
}
