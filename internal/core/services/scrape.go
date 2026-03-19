package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// defaultWorkers is the number of companies scraped concurrently.
const defaultWorkers = 20

// scrapeService implements ports.ScrapeService.
type scrapeService struct {
	companies ports.CompanyRepository
	jobs      ports.JobRepository
	scrapers  map[domain.ATSType]ports.JobScraper
	enricher  ports.JobEnricher // may be nil
	runs      ports.ScrapeRunRepository
	seeder    ports.CompanySeeder
}

// NewScrapeService constructs a ScrapeService.
//
// enricher may be nil — the pipeline runs without enrichment if no enricher is
// configured. scrapers is keyed by ATSType.
func NewScrapeService(
	companies ports.CompanyRepository,
	jobs ports.JobRepository,
	scrapers map[domain.ATSType]ports.JobScraper,
	enricher ports.JobEnricher,
	runs ports.ScrapeRunRepository,
	seeder ports.CompanySeeder,
) ports.ScrapeService {
	return &scrapeService{
		companies: companies,
		jobs:      jobs,
		scrapers:  scrapers,
		enricher:  enricher,
		runs:      runs,
		seeder:    seeder,
	}
}

// Run executes the full scrape pipeline:
//  1. SeedCompanies (idempotent — picks up any new companies from README sources)
//  2. For each active company whose scrape_type is "http":
//     a. Scrape via the appropriate JobScraper
//     b. Upsert each raw job
//     c. MarkInactive for jobs not in this batch
//  3. Persist a ScrapeRun record with counters
//
// Per-company errors are logged and skipped — a single failing scraper does not
// abort the pipeline. Headless companies are skipped with a warning.
func (s *scrapeService) Run(ctx context.Context) error {
	// Seed before scraping so new companies discovered in the README are included.
	if err := s.SeedCompanies(ctx); err != nil {
		log.Printf("scrape: seed companies failed (continuing): %v", err)
	}

	run := domain.ScrapeRun{
		ID:        uuid.New(),
		StartedAt: time.Now(),
		Status:    domain.ScrapeStatusRunning,
	}
	if err := s.runs.Create(ctx, run); err != nil {
		// Non-fatal — record might not persist but scrape still runs.
		log.Printf("scrape: failed to create run record: %v", err)
	}

	companies, err := s.companies.ListActive(ctx)
	if err != nil {
		run.Status = domain.ScrapeStatusFailed
		run.Error = fmt.Sprintf("list companies: %v", err)
		_ = s.runs.Update(ctx, run)
		return fmt.Errorf("scrape: list active companies: %w", err)
	}

	var (
		totalAdded   atomic.Int64
		totalUpdated atomic.Int64
		totalRemoved atomic.Int64
		wg           sync.WaitGroup
		sem          = semaphore.NewWeighted(defaultWorkers)
	)

	for _, company := range companies {
		if company.ScrapeType == domain.ScrapeHeadless {
			log.Printf("scrape: skipping headless company %q (post-MVP)", company.Name)
			continue
		}

		scraper, ok := s.scrapers[company.ATSType]
		if !ok {
			log.Printf("scrape: no scraper for ATS type %q (company %q)", company.ATSType, company.Name)
			continue
		}

		// Acquire a worker slot before launching the goroutine.
		if err := sem.Acquire(ctx, 1); err != nil {
			// Context cancelled — stop launching new work.
			break
		}

		wg.Add(1)
		go func(c domain.Company, sc ports.JobScraper) {
			defer wg.Done()
			defer sem.Release(1)

			added, updated, removed, err := s.scrapeCompany(ctx, c, sc)
			if err != nil {
				log.Printf("scrape: company %q failed: %v", c.Name, err)
				return
			}
			totalAdded.Add(int64(added))
			totalUpdated.Add(int64(updated))
			totalRemoved.Add(int64(removed))
		}(company, scraper)
	}

	wg.Wait()

	run.JobsAdded = int(totalAdded.Load())
	run.JobsUpdated = int(totalUpdated.Load())
	run.JobsRemoved = int(totalRemoved.Load())

	now := time.Now()
	run.FinishedAt = &now
	run.Status = domain.ScrapeStatusCompleted
	if err := s.runs.Update(ctx, run); err != nil {
		log.Printf("scrape: failed to update run record: %v", err)
	}

	log.Printf("scrape: complete — added=%d updated=%d removed=%d", run.JobsAdded, run.JobsUpdated, run.JobsRemoved)
	return nil
}

// scrapeCompany runs the scrape pipeline for a single company.
// Returns (added, updated, removed) counts.
func (s *scrapeService) scrapeCompany(ctx context.Context, company domain.Company, scraper ports.JobScraper) (added, updated, removed int, err error) {
	rawJobs, err := scraper.Scrape(ctx, company)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("scrape %q: %w", company.Name, err)
	}

	seenIDs := make([]string, 0, len(rawJobs))
	for _, raw := range rawJobs {
		seenIDs = append(seenIDs, raw.ExternalID)

		_, err := s.jobs.Upsert(ctx, company.ID, raw, company.ATSType)
		if err != nil {
			log.Printf("scrape: upsert job %q for %q: %v", raw.ExternalID, company.Name, err)
			continue
		}

		// For M1, count every upsert as "updated". A future enhancement could
		// differentiate inserts from updates by comparing row counts.
		updated++
	}

	// Mark jobs not in this batch as inactive.
	deactivated, err := s.jobs.MarkInactive(ctx, company.ID, seenIDs)
	if err != nil {
		log.Printf("scrape: mark inactive for %q: %v", company.Name, err)
	} else {
		removed += deactivated
	}

	return added, updated, removed, nil
}

// SeedCompanies fetches company slugs from the Simplify README sources and
// upserts them into the company list. Safe to call on every scrape cycle.
func (s *scrapeService) SeedCompanies(ctx context.Context) error {
	companies, err := s.seeder.Seed(ctx)
	if err != nil {
		return fmt.Errorf("seed: fetch companies: %w", err)
	}

	for _, c := range companies {
		if _, err := s.companies.Upsert(ctx, c); err != nil {
			log.Printf("seed: upsert company %q (%s): %v", c.Name, c.BoardToken, err)
		}
	}

	log.Printf("seed: upserted %d companies", len(companies))
	return nil
}

// LatestRun returns the most recent ScrapeRun record.
func (s *scrapeService) LatestRun(ctx context.Context) (domain.ScrapeRun, error) {
	run, err := s.runs.Latest(ctx)
	if err != nil {
		return domain.ScrapeRun{}, fmt.Errorf("latest run: %w", err)
	}
	return run, nil
}
