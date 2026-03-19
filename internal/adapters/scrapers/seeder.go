package scrapers

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// Simplify README sources that embed ATS apply links in markdown tables.
var simplifyREADMESources = []string{
	"https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/dev/README.md",
	"https://raw.githubusercontent.com/SimplifyJobs/New-Grad-Positions/dev/README.md",
}

// URL patterns for each ATS platform — captures the board token.
var (
	// Greenhouse: https://boards.greenhouse.io/{board_token}/... or
	//             https://job-boards.greenhouse.io/{board_token}/...
	reGreenhouse = regexp.MustCompile(`https://(?:boards|job-boards)\.greenhouse\.io/([^/\s)]+)`)

	// Lever: https://jobs.lever.co/{company_slug}/... or
	//        https://jobs.lever.co/{company_slug}
	reLever = regexp.MustCompile(`https://jobs\.lever\.co/([^/\s)]+)`)

	// Ashby: https://jobs.ashbyhq.com/{board_name}/... or
	//        https://app.ashbyhq.com/jobs/.../{board_name}
	reAshby = regexp.MustCompile(`https://(?:jobs\.ashbyhq\.com|app\.ashbyhq\.com/jobs/[^/]+)/([^/\s)?#]+)`)
)

// SimplifySeeder implements ports.CompanySeeder by parsing the Simplify README files.
//
// Company slugs/tokens are extracted by URL pattern per platform and returned as
// domain.Company values ready for upsert. Parsing is idempotent — safe to call on
// every scrape cycle; new companies appear automatically as the READMEs update.
type SimplifySeeder struct {
	client *http.Client
}

// NewSimplifySeeder constructs a SimplifySeeder.
func NewSimplifySeeder() *SimplifySeeder {
	return &SimplifySeeder{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Seed fetches the Simplify README files and extracts unique company tokens for
// Greenhouse, Lever, and Ashby. Returns one Company per unique (ats_type, board_token).
func (s *SimplifySeeder) Seed(ctx context.Context) ([]domain.Company, error) {
	// Use a set keyed by "atsType:boardToken" to deduplicate across README sources.
	seen := make(map[string]domain.Company)

	for _, url := range simplifyREADMESources {
		companies, err := s.parseREADME(ctx, url)
		if err != nil {
			// Log and continue — a single failing source doesn't abort seeding.
			fmt.Printf("seeder: skipping %s: %v\n", url, err)
			continue
		}
		for _, c := range companies {
			key := string(c.ATSType) + ":" + c.BoardToken
			if _, exists := seen[key]; !exists {
				seen[key] = c
			}
		}
	}

	// Merge static curated companies — skipped if already present from README sources.
	for _, c := range staticCompanies {
		key := string(c.ATSType) + ":" + c.BoardToken
		if _, exists := seen[key]; !exists {
			seen[key] = c
		}
	}

	result := make([]domain.Company, 0, len(seen))
	for _, c := range seen {
		result = append(result, c)
	}
	return result, nil
}

// parseREADME fetches a single README and extracts Company values from ATS URLs.
func (s *SimplifySeeder) parseREADME(ctx context.Context, url string) ([]domain.Company, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "go-jobs/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s returned HTTP %d", url, resp.StatusCode)
	}

	var companies []domain.Company
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract company name from markdown table cells — first cell is usually company
		// e.g.: | **[Stripe](https://stripe.com)** | Software Engineer | ...
		name := extractCompanyName(line)

		// Greenhouse matches
		if m := reGreenhouse.FindStringSubmatch(line); m != nil {
			token := cleanToken(m[1])
			if token != "" {
				companies = append(companies, domain.Company{
					Name:       name,
					ATSType:    domain.ATSGreenhouse,
					ScrapeType: domain.ScrapeHTTP,
					BoardToken: token,
					Active:     true,
				})
			}
		}

		// Lever matches
		if m := reLever.FindStringSubmatch(line); m != nil {
			token := cleanToken(m[1])
			if token != "" {
				companies = append(companies, domain.Company{
					Name:       name,
					ATSType:    domain.ATSLever,
					ScrapeType: domain.ScrapeHTTP,
					BoardToken: token,
					Active:     true,
				})
			}
		}

		// Ashby matches
		if m := reAshby.FindStringSubmatch(line); m != nil {
			token := cleanToken(m[1])
			if token != "" {
				companies = append(companies, domain.Company{
					Name:       name,
					ATSType:    domain.ATSAshby,
					ScrapeType: domain.ScrapeHTTP,
					BoardToken: token,
					Active:     true,
				})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", url, err)
	}

	return companies, nil
}

// reCompanyName matches the first link text in a markdown table row.
var reCompanyName = regexp.MustCompile(`\*?\*?\[([^\]]+)\]`)

// extractCompanyName extracts the first link label from a markdown table row.
// Returns an empty string if no match is found.
func extractCompanyName(line string) string {
	if m := reCompanyName.FindStringSubmatch(line); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// cleanToken strips trailing punctuation and query params from a captured token.
func cleanToken(s string) string {
	// Strip anything after '?' or '#'
	if i := strings.IndexAny(s, "?#"); i != -1 {
		s = s[:i]
	}
	// Strip trailing slashes and common markdown punctuation
	s = strings.TrimRight(s, "/)")
	s = strings.TrimSpace(s)
	return s
}
