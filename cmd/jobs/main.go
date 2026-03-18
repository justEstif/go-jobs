package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/justestif/go-jobs/internal/adapters/enrichment"
	httphandlers "github.com/justestif/go-jobs/internal/adapters/http"
	"github.com/justestif/go-jobs/internal/adapters/http/api"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/adapters/httpclient"
	"github.com/justestif/go-jobs/internal/adapters/postgres"
	"github.com/justestif/go-jobs/internal/adapters/scrapers"
	"github.com/justestif/go-jobs/internal/cli"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
	"github.com/justestif/go-jobs/internal/core/services"
)

func main() {
	// ----------------------------------------------------------------
	// Remote mode detection
	//
	// Check --base-url flag and BASE_URL env before booting any adapters.
	// Commands that require direct DB access (serve, scrape, enrich) always
	// run in local mode regardless of base URL.
	// ----------------------------------------------------------------
	baseURL := resolveBaseURL(os.Args[1:])
	cmd := firstCommand(os.Args[1:])
	localOnlyCmd := cmd == "serve" || cmd == "scrape" || cmd == "enrich"
	remoteMode := baseURL != "" && !localOnlyCmd

	if remoteMode {
		runRemoteCLI(baseURL)
		return
	}

	// ----------------------------------------------------------------
	// Local mode: boot the full stack (DB + scrapers + enrichment)
	// ----------------------------------------------------------------
	if err := postgres.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer postgres.Close()

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

	// Enrichment adapter (tiered: ATS → rules → LLM)
	// LLM tier is disabled until the user configures an API key (M5).
	enricher := enrichment.NewTieredEnricher(domain.LLMProvider(""), "")

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
	userService := services.NewUserService(userRepo)
	companyService := services.NewCompanyService(companyRepo, userCompanyRepo)
	serve := func(ctx context.Context) error {
		return runHTTPServer(
			ctx,
			scrapeService,
			enrichService,
			authService,
			searchService,
			applicationService,
			userService,
			companyService,
		)
	}

	// No CLI args → start the web server directly.
	if len(os.Args) == 1 {
		if err := serve(context.Background()); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		return
	}

	cliServices := cli.Services{
		Scrape:      scrapeService,
		Enrich:      enrichService,
		Search:      searchService,
		Application: applicationService,
		Session:     userRepo,
		Auth:        authService,
		Serve:       serve,
	}
	rootCmd := cli.NewRootCmd(cliServices)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
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

// resolveBaseURL returns the effective base URL from args or environment.
//
// Precedence:
//  1. --base-url=<value> or --base-url <value> in args
//  2. BASE_URL environment variable
//  3. Empty string (local/in-process mode)
func resolveBaseURL(args []string) string {
	for i, arg := range args {
		if len(arg) > 11 && arg[:11] == "--base-url=" {
			return arg[11:]
		}
		if arg == "--base-url" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return os.Getenv("BASE_URL")
}

// firstCommand returns the first non-flag argument in args, which cobra treats
// as the subcommand name. Returns an empty string if none is found.
// It skips the value argument that follows --base-url so it is not mistaken
// for a subcommand name.
func firstCommand(args []string) string {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--base-url" {
			skipNext = true
			continue
		}
		// Skip --base-url=value form and any other --flag=value or --flag forms.
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}
		return arg
	}
	return ""
}

func runHTTPServer(
	ctx context.Context,
	scrapeService ports.ScrapeService,
	enrichService ports.EnrichService,
	authService ports.AuthService,
	searchService ports.JobSearchService,
	applicationService ports.ApplicationService,
	userService ports.UserService,
	companyService ports.CompanyService,
) error {
	sessionSecret := []byte(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) == 0 {
		return errors.New("SESSION_SECRET environment variable not set")
	}
	sm := middleware.NewSessionManager(postgres.Pool, sessionSecret)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(sm.LoadAndSave)

	csrfKey := []byte(os.Getenv("CSRF_KEY"))
	if len(csrfKey) != 32 {
		return errors.New("CSRF_KEY must be exactly 32 bytes long")
	}
	csrfMw := middleware.SetupCSRF(csrfKey, false)

	r.Use(middleware.OptionalAuth(sm))
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	authH := httphandlers.NewAuthHandler(authService, sm)
	jobsH := httphandlers.NewJobSearchHandler(searchService, applicationService, userService)
	trackerH := httphandlers.NewTrackerHandler(applicationService, searchService)
	pipelineH := httphandlers.NewPipelineHandler(applicationService)
	companyH := httphandlers.NewCompanyHandler(companyService)
	apiH := api.New(authService, searchService, applicationService)

	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Get("/", jobsH.List)
		r.Get("/about", httphandlers.About)
		r.Get("/contact", httphandlers.ContactForm)
		r.Post("/contact", httphandlers.ContactSubmit)
		r.Get("/register", authH.ShowRegister)
		r.Post("/register", authH.Register)
		r.Get("/login", authH.ShowLogin)
		r.Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.Get("/jobs/{id}", jobsH.Detail)
	})

	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Use(middleware.RequireAuth(sm))
		r.Post("/jobs/{id}/interested", trackerH.Interested)
		r.Post("/jobs/{id}/apply", trackerH.Apply)
		r.Post("/jobs/{id}/status", trackerH.SetStatus)
		r.Post("/jobs/{id}/notes", trackerH.SetNotes)
		r.Get("/pipeline", pipelineH.List)
		r.Get("/companies", companyH.List)
		r.Post("/companies/{id}/hide", companyH.Hide)
		r.Post("/companies/{id}/show", companyH.Show)
	})

	// ----------------------------------------------------------------
	// /api/v1/ — JSON driving adapter (CLI, external clients)
	// Auth: Authorization: Bearer <token>. No CSRF required.
	// ----------------------------------------------------------------
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints.
		r.Post("/auth/register", apiH.Register)
		r.Post("/auth/login", apiH.Login)
		r.Get("/jobs", apiH.Search)

		// Protected endpoints — require valid Bearer token.
		r.Group(func(r chi.Router) {
			r.Use(middleware.BearerAuth(authService))
			r.Get("/auth/me", apiH.Me)
			r.Post("/auth/logout", apiH.Logout)
			r.Get("/jobs/interested", apiH.Interested)
			r.Get("/jobs/applied", apiH.Applied)
			r.Post("/jobs/{id}/interested", apiH.MarkInterested)
			r.Post("/jobs/{id}/apply", apiH.MarkApplied)
			r.Post("/jobs/{id}/status", apiH.SetStatus)
			r.Post("/jobs/{id}/notes", apiH.SetNotes)
			r.Get("/pipeline", apiH.Pipeline)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	scrapeInterval := parseDurationEnv("SCRAPE_INTERVAL", 6*time.Hour)
	enrichLimit := parseIntEnv("ENRICH_LIMIT", 1000)

	schedulerCtx, schedulerCancel := context.WithCancel(ctx)
	defer schedulerCancel()
	go runScheduledPipeline(schedulerCtx, scrapeService, enrichService, scrapeInterval, enrichLimit)

	server := &http.Server{Addr: ":" + port, Handler: r}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Printf("Server starting on http://localhost:%s", port)
	log.Printf("Scheduler enabled: scrape+enrich every %s (enrich_limit=%d)", scrapeInterval, enrichLimit)

	select {
	case <-ctx.Done():
		log.Printf("Shutdown signal received: context canceled")
	case sig := <-sigCh:
		log.Printf("Shutdown signal received: %s", sig)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	return nil
}

func runScheduledPipeline(ctx context.Context, scrape ports.ScrapeService, enrich ports.EnrichService, interval time.Duration, enrichLimit int) {
	run := func() {
		if err := scrape.Run(ctx); err != nil {
			log.Printf("scheduler: scrape failed: %v", err)
			return
		}
		enriched, failed, err := enrich.Run(ctx, enrichLimit)
		if err != nil {
			log.Printf("scheduler: enrich failed: %v", err)
			return
		}
		log.Printf("scheduler: enrich complete enriched=%d failed=%d", enriched, failed)
	}

	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("scheduler: stopped")
			return
		case <-ticker.C:
			run()
		}
	}
}

func parseDurationEnv(name string, fallback time.Duration) time.Duration {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		log.Printf("Invalid %s=%q; using default %s", name, raw, fallback)
		return fallback
	}
	return d
}

func parseIntEnv(name string, fallback int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		log.Printf("Invalid %s=%q; using default %d", name, raw, fallback)
		return fallback
	}
	return v
}
