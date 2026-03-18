package cli

import (
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
}

// NewRootCmd builds the cobra root command with all sub-commands attached.
// services is injected by cmd/jobs/main.go — the only place that knows concrete types.
func NewRootCmd(services Services) *cobra.Command {
	root := &cobra.Command{
		Use:   "go-jobs",
		Short: "Self-hosted job aggregator",
		Long:  "go-jobs scrapes startup ATS platforms and tracks job applications.",
	}

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

	return root
}
