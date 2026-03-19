package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/justestif/go-jobs/internal/adapters/crypto"
	"github.com/justestif/go-jobs/internal/adapters/enrichment"
	"github.com/justestif/go-jobs/internal/adapters/httpclient"
	"github.com/justestif/go-jobs/internal/adapters/llm"
	"github.com/justestif/go-jobs/internal/adapters/postgres"
	"github.com/justestif/go-jobs/internal/adapters/scrapers"
	"github.com/justestif/go-jobs/internal/cli"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
	"github.com/justestif/go-jobs/internal/core/services"
)

// defaultBaseURL is the hosted server used when neither --base-url nor
// BASE_URL is set. This lets `npm install -g @justestif/go-jobs && go-jobs search`
// work out of the box without a local PostgreSQL database.
const defaultBaseURL = "https://jobs.estifanos.cc"

// resolveBaseURL checks for --base-url flag or BASE_URL env var before cobra runs.
// Falls back to the hosted server so the CLI works without a local database.
func resolveBaseURL() string {
	args := os.Args
	for i := 1; i < len(args); i++ {
		if args[i] == "--base-url" && i+1 < len(args) {
			return args[i+1]
		}
		if len(args[i]) > 12 && args[i][:12] == "--base-url=" {
			return args[i][12:]
		}
	}
	if env := os.Getenv("BASE_URL"); env != "" {
		return env
	}
	return defaultBaseURL
}

func main() {
	baseURL := resolveBaseURL()

	// Remote mode is the default. Local mode (direct DB access) is only used
	// when the special sentinel value "local" is passed via --base-url or
	// BASE_URL. This keeps the npm-installed CLI working without PostgreSQL.
	if baseURL == "local" {
		ctx := context.Background()
		svc, err := setupLocalServices(ctx)
		if err != nil {
			log.Fatalf("Failed to setup services: %v", err)
		}
		rootCmd := cli.NewRootCmd(svc)
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	runRemoteCLI(baseURL)
}

// runRemoteCLI wires httpclient adapters targeting baseURL and executes the
// cobra command tree. The DB is never initialised in this path.
func runRemoteCLI(baseURL string) {
	// Read the stored token for authenticated commands. An empty token is fine
	// for public commands (register, login, search).
	token, err := cli.ReadStoredToken()
	if err != nil {
		log.Fatalf("Failed to read stored token: %v", err)
	}

	c := httpclient.NewClient(baseURL, token)
	authClient := httpclient.NewAuthClient(c)

	cliServices := cli.Services{
		Auth:        authClient,
		Session:     authClient,
		Search:      httpclient.NewSearchClient(c),
		Application: httpclient.NewApplicationClient(c),
		// Scrape, Enrich, Serve are not available in remote mode.
	}

	rootCmd := cli.NewRootCmd(cliServices)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// setupLocalServices wires up all the local services (DB, scrapers, enrichment)
func setupLocalServices(ctx context.Context) (cli.Services, error) {
	// Initialize database
	if err := postgres.InitDB(); err != nil {
		return cli.Services{}, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Driven adapters (repos)
	companyRepo := postgres.NewCompanyRepo(postgres.DB)
	jobRepo := postgres.NewJobRepo(postgres.DB)
	userRepo := postgres.NewUserRepo(postgres.DB)
	userJobRepo := postgres.NewUserJobRepo(postgres.DB)
	userCompanyRepo := postgres.NewUserCompanyRepo(postgres.DB)
	scrapeRunRepo := postgres.NewScrapeRunRepo(postgres.DB)

	// Scraper adapters
	scraperMap := map[domain.ATSType]ports.JobScraper{
		domain.ATSGreenhouse: scrapers.NewGreenhouseAdapter(),
		domain.ATSLever:      scrapers.NewLeverAdapter(),
		domain.ATSAshby:      scrapers.NewAshbyAdapter(),
	}
	seeder := scrapers.NewSimplifySeeder()

	// Enrichment adapter (two-tier: ATS → rules)
	enricher := enrichment.NewTieredEnricher()

	// Encryption for user API keys at rest.
	encryptor, err := crypto.NewKeyEncryptor()
	if err != nil {
		log.Printf("Warning: %v — LLM API key encryption disabled (Ollama still works)", err)
	}
	encryptFn := func(plaintext string) (string, error) {
		if encryptor == nil {
			return plaintext, nil
		}
		return encryptor.Encrypt(plaintext)
	}
	decryptFn := func(ciphertext string) (string, error) {
		if encryptor == nil {
			return ciphertext, nil
		}
		return encryptor.Decrypt(ciphertext)
	}

	// Core services
	scrapeService := services.NewScrapeService(
		companyRepo,
		jobRepo,
		scraperMap,
		enricher,
		scrapeRunRepo,
		seeder,
	)
	enrichService := services.NewEnrichService(jobRepo, enricher)
	searchService := services.NewJobSearchService(jobRepo)
	applicationService := services.NewApplicationService(userJobRepo, jobRepo)
	authService := services.NewAuthService(userRepo, userRepo)
	userService := services.NewUserService(userRepo, encryptFn)
	_ = services.NewCompanyService(companyRepo, userCompanyRepo) // not used in CLI
	coachCacheRepo := postgres.NewCoachCacheRepo(postgres.DB)
	coachService := services.NewJobCoachService(userRepo, jobRepo, companyRepo, coachCacheRepo, llm.NewClient, decryptFn)

	serve := func(ctx context.Context) error {
		return fmt.Errorf("server mode not available in CLI binary")
	}

	return cli.Services{
		Scrape:      scrapeService,
		Enrich:      enrichService,
		Search:      searchService,
		Application: applicationService,
		Session:     userRepo,
		Auth:        authService,
		User:        userService,
		Coach:       coachService,
		Serve:       serve,
	}, nil
}
