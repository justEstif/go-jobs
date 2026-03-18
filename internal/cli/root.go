package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/justestif/go-jobs/internal/core/ports"
)

// Services holds the driving-port implementations injected by the composition root.
type Services struct {
	Scrape      ports.ScrapeService
	Enrich      ports.EnrichService
	Search      ports.JobSearchService
	Application ports.ApplicationService
	Session     ports.SessionRepository
	Auth        ports.AuthService
	Serve       func(ctx context.Context) error
}

// NewRootCmd builds the cobra root command with all sub-commands attached.
// services is injected by cmd/jobs/main.go — the only place that knows concrete types.
//
// The --base-url flag is registered here for help text and shell completion.
// The flag value is pre-resolved by the composition root before cobra runs
// (see cmd/jobs/main.go) so that the correct adapters can be wired without
// requiring the database when targeting a remote server.
func NewRootCmd(services Services) *cobra.Command {
	root := &cobra.Command{
		Use:   "go-jobs",
		Short: "Self-hosted job aggregator",
		Long:  "go-jobs — CLI for jobs.estifanos.cc. Search job postings and track applications.",
	}

	// --base-url is a persistent flag so it appears in every sub-command's
	// --help output. The actual value is consumed by the composition root
	// before Execute() is called; cobra stores it here as documentation only.
	root.PersistentFlags().String("base-url", "", "go-jobs server URL (overrides BASE_URL env; default: http://127.0.0.1:3000)")

	root.AddCommand(newScrapeCmd(services))
	root.AddCommand(newEnrichCmd(services))
	root.AddCommand(newSearchCmd(services))
	root.AddCommand(newInterestedCmd(services))
	root.AddCommand(newApplyCmd(services))
	root.AddCommand(newStatusCmd(services))
	root.AddCommand(newNotesCmd(services))
	root.AddCommand(newAppliedCmd(services))
	root.AddCommand(newPipelineCmd(services))
	root.AddCommand(newRegisterCmd(services))
	root.AddCommand(newLoginCmd(services))
	root.AddCommand(newLogoutCmd(services))
	root.AddCommand(newServeCmd(services))

	return root
}
