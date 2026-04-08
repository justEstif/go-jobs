package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

const fuzzyMatchThreshold = 0.5

// atsProbeEndpoints maps ATS types to their API URL templates.
// The placeholder {slug} is replaced with the slugified company name.
var atsProbeEndpoints = []struct {
	atsType  domain.ATSType
	urlTmpl  string
	apiTmpl  string
}{
	{domain.ATSGreenhouse, "https://boards-api.greenhouse.io/v1/boards/{slug}/jobs", "https://boards-api.greenhouse.io/v1/boards/{slug}/jobs"},
	{domain.ATSLever, "https://api.lever.co/v0/postings/{slug}", "https://api.lever.co/v0/postings/{slug}?mode=json"},
	{domain.ATSAshby, "https://api.ashbyhq.com/posting-api/job-board/{slug}", "https://api.ashbyhq.com/posting-api/job-board/{slug}"},
}

type contactService struct {
	contacts  ports.ContactRepository
	matcher   ports.CompanyMatcher
	companies ports.CompanyRepository
	client    *http.Client
}

// NewContactService constructs a ContactService.
func NewContactService(
	contacts ports.ContactRepository,
	matcher ports.CompanyMatcher,
	companies ports.CompanyRepository,
) ports.ContactService {
	return &contactService{
		contacts:  contacts,
		matcher:   matcher,
		companies: companies,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

func (s *contactService) ImportCSV(ctx context.Context, userID domain.UserID, r io.Reader) (domain.ImportResult, error) {
	contacts, err := parseLinkedInCSV(r)
	if err != nil {
		return domain.ImportResult{}, fmt.Errorf("parse CSV: %w", err)
	}

	var result domain.ImportResult

	// Collect unique normalized company names for batch matching.
	companyMap := make(map[string]struct{})
	for i := range contacts {
		contacts[i].UserID = userID
		contacts[i].NormalizedCompanyName = domain.NormalizeCompanyName(contacts[i].CompanyName)
		companyMap[contacts[i].NormalizedCompanyName] = struct{}{}
	}

	// Phase 1: Match companies using DB lookups (fast).
	nameToCompany := make(map[string]*domain.CompanyID)
	var unmatchedNames []string

	for normalized := range companyMap {
		if normalized == "" {
			continue
		}

		// Exact match.
		company, err := s.matcher.GetByNormalizedName(ctx, normalized)
		if err == nil {
			id := company.ID
			nameToCompany[normalized] = &id
			result.CompaniesLinked++
			continue
		}

		// Fuzzy match.
		company, score, err := s.matcher.FuzzyMatch(ctx, normalized)
		if err == nil && score >= fuzzyMatchThreshold {
			id := company.ID
			nameToCompany[normalized] = &id
			result.CompaniesLinked++
			continue
		}

		unmatchedNames = append(unmatchedNames, normalized)
	}

	// Phase 2: Upsert contacts with company IDs from DB matches.
	for i := range contacts {
		if cID, ok := nameToCompany[contacts[i].NormalizedCompanyName]; ok {
			contacts[i].CompanyID = cID
		}
		if _, err := s.contacts.Upsert(ctx, contacts[i]); err != nil {
			log.Printf("contact import: failed to upsert contact %q: %v", contacts[i].FirstName+" "+contacts[i].LastName, err)
			continue
		}
		result.ContactsImported++
	}

	result.CompaniesUnmatched = len(unmatchedNames)

	// Phase 3: ATS probing for unmatched companies (slow — runs in background).
	// Contacts are already saved; this phase discovers new companies and links them.
	if len(unmatchedNames) > 0 {
		go s.probeAndLinkCompanies(unmatchedNames)
	}

	return result, nil
}

// probeAndLinkCompanies runs ATS probes for unmatched company names and links
// contacts to any newly discovered companies. Runs in a background goroutine.
func (s *contactService) probeAndLinkCompanies(names []string) {
	ctx := context.Background()
	for _, normalized := range names {
		slug := domain.Slugify(normalized)
		if slug == "" {
			continue
		}

		atsType, found := s.probeATS(ctx, slug)
		if !found {
			continue
		}

		newCompany := domain.Company{
			Name:       normalized,
			ATSType:    atsType,
			ScrapeType: domain.ScrapeHTTP,
			BoardToken: slug,
			Active:     true,
		}
		companyID, err := s.companies.Upsert(ctx, newCompany)
		if err != nil {
			log.Printf("background ATS probe: failed to register company %q: %v", normalized, err)
			continue
		}

		linked, err := s.contacts.LinkToCompany(ctx, normalized, companyID)
		if err != nil {
			log.Printf("background ATS probe: failed to link contacts for %q: %v", normalized, err)
			continue
		}
		log.Printf("background ATS probe: registered %q (%s) and linked %d contacts", normalized, atsType, linked)
	}
	log.Printf("background ATS probe: finished processing %d companies", len(names))
}

func (s *contactService) ContactsAtCompany(ctx context.Context, userID domain.UserID, companyID domain.CompanyID) ([]domain.Contact, error) {
	return s.contacts.ListByCompanyID(ctx, userID, companyID)
}

func (s *contactService) ContactsAtCompanies(ctx context.Context, userID domain.UserID, companyIDs []domain.CompanyID) (map[domain.CompanyID][]domain.Contact, error) {
	if len(companyIDs) == 0 {
		return nil, nil
	}
	contacts, err := s.contacts.ListByCompanyIDs(ctx, userID, companyIDs)
	if err != nil {
		return nil, err
	}
	grouped := make(map[domain.CompanyID][]domain.Contact)
	for _, c := range contacts {
		if c.CompanyID != nil {
			grouped[*c.CompanyID] = append(grouped[*c.CompanyID], c)
		}
	}
	return grouped, nil
}

func (s *contactService) LinkedCompanyIDs(ctx context.Context, userID domain.UserID) ([]domain.CompanyID, error) {
	return s.contacts.ListLinkedCompanyIDs(ctx, userID)
}

func (s *contactService) DeleteContacts(ctx context.Context, userID domain.UserID) error {
	return s.contacts.DeleteAllForUser(ctx, userID)
}

func (s *contactService) Stats(ctx context.Context, userID domain.UserID) (total int64, linked int64, companies int64, err error) {
	total, err = s.contacts.CountForUser(ctx, userID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count contacts: %w", err)
	}
	linked, err = s.contacts.CountLinkedForUser(ctx, userID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count linked contacts: %w", err)
	}
	companies, err = s.contacts.CountDistinctCompaniesForUser(ctx, userID)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("count companies: %w", err)
	}
	return total, linked, companies, nil
}

// probeATS tries HEAD requests against known ATS API endpoints for the given slug.
// Returns the ATS type and true if any responds with 200.
func (s *contactService) probeATS(ctx context.Context, slug string) (domain.ATSType, bool) {
	for _, ep := range atsProbeEndpoints {
		url := strings.ReplaceAll(ep.urlTmpl, "{slug}", slug)
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			continue
		}
		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return ep.atsType, true
		}
	}
	return "", false
}

// parseLinkedInCSV reads a LinkedIn Connections.csv export.
// Expected columns: First Name, Last Name, Email Address, Company, Position, Connected On.
// LinkedIn prepends a notes preamble before the actual CSV header — we skip it.
func parseLinkedInCSV(r io.Reader) ([]domain.Contact, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // allow variable field count (preamble lines)

	// Skip preamble lines until we find the actual header row containing "First Name".
	var header []string
	for {
		row, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		// Look for the real header row.
		for _, col := range row {
			if strings.EqualFold(strings.TrimSpace(col), "first name") {
				header = row
				break
			}
		}
		if header != nil {
			break
		}
	}

	// Build column index map (case-insensitive, trimmed).
	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	getCol := func(row []string, name string) string {
		if idx, ok := colIdx[name]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	var contacts []domain.Contact
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // skip malformed rows
		}

		company := getCol(row, "company")
		if company == "" {
			continue
		}

		firstName := getCol(row, "first name")
		lastName := getCol(row, "last name")

		// Build LinkedIn URL from name if not provided.
		linkedinURL := getCol(row, "url")
		if linkedinURL == "" {
			linkedinURL = getCol(row, "profile url")
		}
		if linkedinURL == "" && firstName != "" && lastName != "" {
			slug := domain.Slugify(firstName + " " + lastName)
			if slug != "" {
				linkedinURL = "https://www.linkedin.com/in/" + slug
			}
		}

		c := domain.Contact{
			FirstName:   firstName,
			LastName:    lastName,
			Email:       getCol(row, "email address"),
			Title:       getCol(row, "position"),
			LinkedInURL: linkedinURL,
			CompanyName: company,
		}

		// Parse Connected On date (format: "DD Mon YYYY" or "DD/MM/YYYY").
		if connStr := getCol(row, "connected on"); connStr != "" {
			if t, err := parseLinkedInDate(connStr); err == nil {
				c.ConnectedOn = &t
			}
		}

		contacts = append(contacts, c)
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("no contacts with company found in CSV")
	}
	return contacts, nil
}

// parseLinkedInDate handles LinkedIn's date format "02 Jan 2006".
func parseLinkedInDate(s string) (time.Time, error) {
	// LinkedIn uses "02 Jan 2006" format.
	t, err := time.Parse("02 Jan 2006", s)
	if err != nil {
		// Try alternative format.
		t, err = time.Parse("2 Jan 2006", s)
	}
	if err != nil {
		// Try DD/MM/YYYY.
		t, err = time.Parse("02/01/2006", s)
	}
	return t, err
}

// Ensure pgx.ErrNoRows is used for matching.
var _ = pgx.ErrNoRows
