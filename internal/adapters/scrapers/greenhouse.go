package scrapers

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext" // for jsontext.Value (raw JSON)

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

// greenhouseMetadata is the per-job custom metadata field returned by Greenhouse.
// The value field is polymorphic — it can be a string, bool, number, array, or
// object depending on the field type. We capture it as raw JSON and ignore it;
// we only use standard structured fields (departments, offices) for ATS metadata.
type greenhouseMetadata struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Value     jsontext.Value `json:"value"`
	ValueType string         `json:"value_type"`
}

// greenhousePayload is the top-level response from the Greenhouse boards API.
type greenhousePayload struct {
	Jobs []struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		AbsoluteURL string `json:"absolute_url"`
		Location    struct {
			Name string `json:"name"`
		} `json:"location"`
		Content     string               `json:"content"` // HTML description
		Metadata    []greenhouseMetadata `json:"metadata"`
		Departments []struct {
			Name string `json:"name"`
		} `json:"departments"`
		Offices []struct {
			Name      string `json:"name"`
			CountryID string `json:"country_id"`
		} `json:"offices"`
	} `json:"jobs"`
}

// Scrape fetches all open jobs for company from the Greenhouse boards API.
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
		return []domain.RawJob{}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("greenhouse %s returned HTTP %d", company.BoardToken, resp.StatusCode)
	}

	var payload greenhousePayload
	if err := json.UnmarshalRead(resp.Body, &payload); err != nil {
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

		if len(j.Departments) > 0 {
			raw.Department = j.Departments[0].Name
		}
		if len(j.Offices) > 0 && j.Offices[0].CountryID != "" {
			raw.Country = strings.ToUpper(j.Offices[0].CountryID)
		}

		jobs = append(jobs, raw)
	}

	return jobs, nil
}

// stripHTML removes HTML tags from a string and decodes HTML entities,
// returning plain text. Entity decoding runs after tag removal so that
// entities embedded in attribute values or tag bodies are both handled.
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
	result := strings.ReplaceAll(b.String(), "\n\n\n", "\n\n")
	return strings.TrimSpace(html.UnescapeString(result))
}
