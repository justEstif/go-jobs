package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/justestif/go-jobs/internal/adapters/crypto"
	"github.com/justestif/go-jobs/internal/adapters/enrichment"
	httphandlers "github.com/justestif/go-jobs/internal/adapters/http"
	"github.com/justestif/go-jobs/internal/adapters/http/api"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/adapters/httpclient"
	"github.com/justestif/go-jobs/internal/adapters/llm"
	"github.com/justestif/go-jobs/internal/adapters/postgres"
	"github.com/justestif/go-jobs/internal/adapters/scrapers"
	"github.com/justestif/go-jobs/internal/cli"
	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
	"github.com/justestif/go-jobs/internal/core/services"
	"github.com/justestif/go-jobs/web"
)

func main() {
	// ----------------------------------------------------------------
	// Remote mode detection
	//
	// Check --base-url flag and BASE_URL env before booting any adapters.
	// Commands that require direct DB access (serve, scrape) always
	// run in local mode regardless of base URL.
	// ----------------------------------------------------------------
	baseURL := resolveBaseURL(os.Args[1:])
	cmd := firstCommand(os.Args[1:])
	localOnlyCmd := cmd == "serve" || cmd == "scrape" || cmd == "backfill-tags"
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

	// Enrichment adapter (tiered: ATS → rules; LLM tier removed — see Job Coach)
	enricher := enrichment.NewTieredEnricher()

	// Encryption for user API keys at rest.
	encryptor, err := crypto.NewKeyEncryptor()
	if err != nil {
		log.Printf("Warning: %v — LLM API key encryption disabled (Ollama still works)", err)
	}
	encryptFn := func(plaintext string) (string, error) {
		if encryptor == nil {
			return plaintext, nil // no encryption configured
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
	searchService := services.NewJobSearchService(jobRepo)
	applicationService := services.NewApplicationService(userJobRepo, jobRepo)
	authService := services.NewAuthService(userRepo, userRepo)
	userService := services.NewUserService(userRepo, encryptFn)
	companyService := services.NewCompanyService(companyRepo, userCompanyRepo)
	coachCacheRepo := postgres.NewCoachCacheRepo(postgres.DB)
	coachService := services.NewJobCoachService(userRepo, jobRepo, companyRepo, coachCacheRepo, llm.NewClient, decryptFn)
	contactRepo := postgres.NewContactRepo(postgres.DB)
	contactService := services.NewContactService(contactRepo, contactRepo, companyRepo)
	serve := func(ctx context.Context) error {
		return runHTTPServer(
			ctx,
			scrapeService,
			authService,
			searchService,
			applicationService,
			userService,
			companyService,
			coachService,
			contactService,
		)
	}

	// No CLI args → start the web server directly.
	if len(os.Args) == 1 {
		if err := serve(context.Background()); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		return
	}

	backfillTags := services.BackfillTagsFn(jobRepo, enricher)
	cliServices := cli.Services{
		Scrape:       scrapeService,
		Search:       searchService,
		Application:  applicationService,
		Session:      userRepo,
		Auth:         authService,
		User:         userService,
		Coach:        coachService,
		Contacts:     contactService,
		Serve:        serve,
		BackfillTags: backfillTags,
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
		// Scrape, Serve are not available in remote mode.
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
	authService ports.AuthService,
	searchService ports.JobSearchService,
	applicationService ports.ApplicationService,
	userService ports.UserService,
	companyService ports.CompanyService,
	coachService ports.JobCoachService,
	contactService ports.ContactService,
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
	staticFS, err := fs.Sub(web.Static, "static")
	if err != nil {
		return fmt.Errorf("failed to create static sub-filesystem: %w", err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	authH := httphandlers.NewAuthHandler(authService, sm)
	jobsH := httphandlers.NewJobSearchHandler(searchService, applicationService, userService, companyService)
	trackerH := httphandlers.NewTrackerHandler(applicationService, searchService)
	pipelineH := httphandlers.NewPipelineHandler(applicationService)
	settingsH := httphandlers.NewSettingsHandler(authService, applicationService, companyService, sm)
	coachH := httphandlers.NewCoachHandler(coachService, userService)
	apiH := api.New(authService, searchService, applicationService)
	contactsH := httphandlers.NewContactsHandler(contactService)
	contactsAPI := api.NewContactsAPI(contactService, searchService)

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
		r.Post("/jobs/{id}/analyze", coachH.Analyze)
		r.Get("/pipeline", pipelineH.List)
		r.Get("/contacts", contactsH.Show)
		r.Post("/contacts/import", contactsH.Import)
		r.Post("/contacts/delete", contactsH.Delete)
		r.Post("/companies/{id}/hide", settingsH.HideCompany)
		r.Post("/companies/{id}/show", settingsH.ShowCompany)
		r.Get("/settings", settingsH.Show)
		r.Post("/settings/password", settingsH.ChangePassword)
		r.Post("/settings/reset-tracker", settingsH.ResetTracker)
		r.Post("/settings/delete-account", settingsH.DeleteAccount)
		r.Get("/coach", coachH.Show)
		r.Post("/coach/resume", coachH.SaveResume)
		r.Post("/coach/llm", coachH.SaveLLM)
		r.Post("/coach/case-study", coachH.CaseStudyGenerate)
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
			r.Post("/contacts/import", contactsAPI.ImportCSV)
			r.Get("/contacts/stats", contactsAPI.Stats)
			r.Delete("/contacts", contactsAPI.DeleteAll)
			r.Get("/jobs/{id}/referrals", contactsAPI.Referrals)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	server := &http.Server{Addr: ":" + port, Handler: r}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Printf("Server starting on http://localhost:%s", port)

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
