package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

type fakeJobRepo struct {
	activeByExternalID map[string]bool
}

func (f *fakeJobRepo) Upsert(_ context.Context, _ domain.CompanyID, job domain.RawJob, _ domain.ATSType) (domain.JobID, error) {
	if f.activeByExternalID == nil {
		f.activeByExternalID = make(map[string]bool)
	}
	f.activeByExternalID[job.ExternalID] = true
	return uuid.New(), nil
}

func (f *fakeJobRepo) MarkInactive(_ context.Context, _ domain.CompanyID, activeExternalIDs []string) (int, error) {
	activeNow := make(map[string]bool, len(activeExternalIDs))
	for _, id := range activeExternalIDs {
		activeNow[id] = true
	}

	deactivated := 0
	for externalID, active := range f.activeByExternalID {
		if active && !activeNow[externalID] {
			f.activeByExternalID[externalID] = false
			deactivated++
		}
	}

	return deactivated, nil
}

func (f *fakeJobRepo) GetByID(context.Context, domain.JobID) (domain.Job, error) {
	return domain.Job{}, nil
}

func (f *fakeJobRepo) GetByIDs(context.Context, []domain.JobID) ([]domain.Job, error) {
	return nil, nil
}

func (f *fakeJobRepo) Search(context.Context, domain.SearchFilters, *domain.UserSearchContext) ([]domain.Job, error) {
	return nil, nil
}

func (f *fakeJobRepo) ListUnenriched(context.Context, int) ([]domain.Job, error) {
	return nil, nil
}

func (f *fakeJobRepo) SaveTags(context.Context, domain.JobTags) error {
	return nil
}

type fakeScraper struct {
	jobs []domain.RawJob
}

func (f *fakeScraper) Scrape(context.Context, domain.Company) ([]domain.RawJob, error) {
	return f.jobs, nil
}

var _ ports.JobRepository = (*fakeJobRepo)(nil)

func TestScrapeCompany_MarksMissingJobsInactive(t *testing.T) {
	companyID := uuid.New()
	repo := &fakeJobRepo{}
	svc := &scrapeService{jobs: repo}
	company := domain.Company{ID: companyID, Name: "Acme", ATSType: domain.ATSGreenhouse}

	firstBatch := &fakeScraper{jobs: []domain.RawJob{
		{ExternalID: "job-1", Title: "Backend", URL: "https://example.com/1", FirstSeen: time.Now()},
		{ExternalID: "job-2", Title: "Frontend", URL: "https://example.com/2", FirstSeen: time.Now()},
	}}
	_, _, removed, err := svc.scrapeCompany(context.Background(), company, firstBatch)
	if err != nil {
		t.Fatalf("first scrape failed: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected no removed jobs on first scrape, got %d", removed)
	}

	secondBatch := &fakeScraper{jobs: []domain.RawJob{
		{ExternalID: "job-1", Title: "Backend", URL: "https://example.com/1", FirstSeen: time.Now()},
	}}
	_, _, removed, err = svc.scrapeCompany(context.Background(), company, secondBatch)
	if err != nil {
		t.Fatalf("second scrape failed: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed job, got %d", removed)
	}
	if repo.activeByExternalID["job-2"] {
		t.Fatalf("expected job-2 to be inactive after disappearing from source")
	}
}

func TestScrapeCompany_ReactivatesJobWhenItReappears(t *testing.T) {
	companyID := uuid.New()
	repo := &fakeJobRepo{}
	svc := &scrapeService{jobs: repo}
	company := domain.Company{ID: companyID, Name: "Acme", ATSType: domain.ATSGreenhouse}

	_, _, _, err := svc.scrapeCompany(context.Background(), company, &fakeScraper{jobs: []domain.RawJob{
		{ExternalID: "job-1", Title: "Backend", URL: "https://example.com/1", FirstSeen: time.Now()},
		{ExternalID: "job-2", Title: "Frontend", URL: "https://example.com/2", FirstSeen: time.Now()},
	}})
	if err != nil {
		t.Fatalf("initial scrape failed: %v", err)
	}

	_, _, _, err = svc.scrapeCompany(context.Background(), company, &fakeScraper{jobs: []domain.RawJob{
		{ExternalID: "job-1", Title: "Backend", URL: "https://example.com/1", FirstSeen: time.Now()},
	}})
	if err != nil {
		t.Fatalf("second scrape failed: %v", err)
	}

	_, _, removed, err := svc.scrapeCompany(context.Background(), company, &fakeScraper{jobs: []domain.RawJob{
		{ExternalID: "job-1", Title: "Backend", URL: "https://example.com/1", FirstSeen: time.Now()},
		{ExternalID: "job-2", Title: "Frontend", URL: "https://example.com/2", FirstSeen: time.Now()},
	}})
	if err != nil {
		t.Fatalf("third scrape failed: %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected no removals when job reappears, got %d", removed)
	}
	if !repo.activeByExternalID["job-2"] {
		t.Fatalf("expected job-2 to be active again after reappearing")
	}
}
