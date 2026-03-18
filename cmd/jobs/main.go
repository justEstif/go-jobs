package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/justestif/go-jobs/internal/adapters/enrichment"
	httphandlers "github.com/justestif/go-jobs/internal/adapters/http"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/adapters/postgres"
	"github.com/justestif/go-jobs/internal/adapters/scrapers"
	"github.com/justestif/go-jobs/internal/cli"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
	"github.com/justestif/go-jobs/internal/core/services"
)

func main() {
	// Initialize database
	if err := postgres.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer postgres.Close()

	// ----------------------------------------------------------------
	// Driven adapters (repos)
	// ----------------------------------------------------------------
	companyRepo := postgres.NewCompanyRepo(postgres.DB)
	jobRepo := postgres.NewJobRepo(postgres.DB)
	userRepo := postgres.NewUserRepo(postgres.DB)
	userJobRepo := postgres.NewUserJobRepo(postgres.DB)
	userCompanyRepo := postgres.NewUserCompanyRepo(postgres.DB)
	scrapeRunRepo := postgres.NewScrapeRunRepo(postgres.DB)

	// ----------------------------------------------------------------
	// Scraper adapters
	// ----------------------------------------------------------------
	scraperMap := map[domain.ATSType]ports.JobScraper{
		domain.ATSGreenhouse: scrapers.NewGreenhouseAdapter(),
		domain.ATSLever:      scrapers.NewLeverAdapter(),
		domain.ATSAshby:      scrapers.NewAshbyAdapter(),
	}
	seeder := scrapers.NewSimplifySeeder()

	// ----------------------------------------------------------------
	// Enrichment adapter (tiered: ATS → rules → LLM)
	// LLM tier is disabled until the user configures an API key (M5).
	// ----------------------------------------------------------------
	enricher := enrichment.NewTieredEnricher(domain.LLMProvider(""), "")

	// ----------------------------------------------------------------
	// Core services
	// ----------------------------------------------------------------
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

	// ----------------------------------------------------------------
	// CLI — check if we're running a CLI command (any arg present)
	// ----------------------------------------------------------------
	if len(os.Args) > 1 {
		cliServices := cli.Services{
			Scrape:      scrapeService,
			Enrich:      enrichService,
			Search:      searchService,
			Application: applicationService,
			Session:     userRepo,
			Auth:        authService,
		}
		rootCmd := cli.NewRootCmd(cliServices)
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	// ----------------------------------------------------------------
	// HTTP server
	// ----------------------------------------------------------------
	sessionSecret := []byte(os.Getenv("SESSION_SECRET"))
	if len(sessionSecret) == 0 {
		log.Fatal("SESSION_SECRET environment variable not set")
	}
	sm := middleware.NewSessionManager(postgres.Pool, sessionSecret)

	r := chi.NewRouter()

	// Standard middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// scs session middleware — must come before any handler that reads/writes sessions
	r.Use(sm.LoadAndSave)

	// CSRF protection — set secure=true in production
	csrfKey := []byte(os.Getenv("CSRF_KEY"))
	if len(csrfKey) != 32 {
		log.Fatal("CSRF_KEY must be exactly 32 bytes long")
	}
	csrfMw := middleware.SetupCSRF(csrfKey, false)

	// Optional auth on all routes (loads user ID from session into context)
	r.Use(middleware.OptionalAuth(sm))

	// Static files (no CSRF needed)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// HTTP handler structs
	authH := httphandlers.NewAuthHandler(authService, sm)
	jobsH := httphandlers.NewJobSearchHandler(searchService, applicationService, userService)
	trackerH := httphandlers.NewTrackerHandler(applicationService, searchService)
	pipelineH := httphandlers.NewPipelineHandler(applicationService)
	companyH := httphandlers.NewCompanyHandler(companyService)

	// Public routes with CSRF
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
		// Job detail — public (tracker actions within are auth-gated per handler)
		r.Get("/jobs/{id}", jobsH.Detail)
	})

	// Authenticated routes with CSRF
	r.Group(func(r chi.Router) {
		r.Use(csrfMw)
		r.Use(middleware.RequireAuth(sm))
		// Tracker actions (htmx endpoints)
		r.Post("/jobs/{id}/interested", trackerH.Interested)
		r.Post("/jobs/{id}/apply", trackerH.Apply)
		r.Post("/jobs/{id}/status", trackerH.SetStatus)
		r.Post("/jobs/{id}/notes", trackerH.SetNotes)
		// Pipeline
		r.Get("/pipeline", pipelineH.List)
		// Companies
		r.Get("/companies", companyH.List)
		r.Post("/companies/{id}/hide", companyH.Hide)
		r.Post("/companies/{id}/show", companyH.Show)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
