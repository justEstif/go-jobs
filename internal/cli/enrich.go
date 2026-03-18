package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newEnrichCmd returns the `go-jobs enrich` command.
func newEnrichCmd(services Services) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "enrich",
		Short: "Enrich un-tagged job postings with structured metadata",
		Long: `Enrich fetches un-tagged jobs from the database and runs them through
the tiered enrichment pipeline:

  Tier 1 — ATS metadata extraction (always runs)
  Tier 2 — keyword/regex matching on title and description (always runs)
  Tier 3 — LLM structured output (requires a configured API key; skipped otherwise)

Results are persisted to the job_tags table. Safe to run repeatedly — already
enriched jobs are not re-processed.`,
		Example: `  go-jobs enrich
  go-jobs enrich --limit 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			fmt.Printf("Starting enrichment pipeline (limit=%d)...\n", limit)
			enriched, failed, err := services.Enrich.Run(ctx, limit)
			if err != nil {
				return fmt.Errorf("enrich failed: %w", err)
			}

			fmt.Printf("Enrichment complete:\n")
			fmt.Printf("  Enriched: %d\n", enriched)
			fmt.Printf("  Failed:   %d\n", failed)
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 1000, "Maximum number of jobs to enrich per run")
	return cmd
}
