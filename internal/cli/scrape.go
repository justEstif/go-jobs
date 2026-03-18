package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// newScrapeCmd returns the `go-jobs scrape` command.
func newScrapeCmd(services Services) *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "scrape",
		Short: "Scrape job postings from ATS platforms",
		Long: `Scrape discovers companies from the Simplify README sources,
then fetches open job postings from Greenhouse, Lever, and Ashby.

By default all three platforms are scraped. Use --source to restrict to one.`,
		Example: `  go-jobs scrape
  go-jobs scrape --source greenhouse
  go-jobs scrape --source lever
  go-jobs scrape --source ashby`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if source != "" {
				fmt.Printf("Scraping source: %s\n", source)
				// TODO M1+: filter to a single scraper by source
				// For now, Run() scrapes all — single-source filtering is a convenience
				// feature that requires the service to accept a source filter.
				log.Printf("--source flag noted but full pipeline will run for M1")
			}

			fmt.Println("Starting scrape pipeline...")
			if err := services.Scrape.Run(ctx); err != nil {
				return fmt.Errorf("scrape failed: %w", err)
			}

			run, err := services.Scrape.LatestRun(ctx)
			if err != nil {
				log.Printf("warning: could not retrieve run stats: %v", err)
				return nil
			}

			fmt.Printf("Scrape complete:\n")
			fmt.Printf("  Jobs added:   %d\n", run.JobsAdded)
			fmt.Printf("  Jobs updated: %d\n", run.JobsUpdated)
			fmt.Printf("  Jobs removed: %d\n", run.JobsRemoved)
			if run.Error != "" {
				fmt.Printf("  Error: %s\n", run.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Scrape a single ATS source (greenhouse|lever|ashby)")
	return cmd
}
