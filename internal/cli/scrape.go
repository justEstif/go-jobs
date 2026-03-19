package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// newScrapeCmd returns the `jobs scrape` command.
//
// With --loop the command runs continuously, re-scraping on --interval. This is
// the entry point used by the Dokku worker process defined in the Procfile.
func newScrapeCmd(services Services) *cobra.Command {
	var (
		source   string
		loop     bool
		interval time.Duration
	)

	// defaultInterval reads SCRAPE_INTERVAL from the environment so the Procfile
	// doesn't need shell variable expansion. Falls back to 6h if unset or invalid.
	defaultInterval := 6 * time.Hour
	if v := os.Getenv("SCRAPE_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			defaultInterval = d
		}
	}

	cmd := &cobra.Command{
		Use:   "scrape",
		Short: "Scrape job postings from ATS platforms",
		Long: `Scrape discovers companies from the Simplify README sources,
then fetches open job postings from Greenhouse, Lever, and Ashby.

By default all three platforms are scraped. Use --source to restrict to one.

Pass --loop to run continuously as a daemon (used by the Dokku worker process).`,
		Example: `  jobs scrape
  jobs scrape --source greenhouse
  jobs scrape --loop --interval 6h`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if source != "" {
				log.Printf("--source flag noted but full pipeline will run (single-source filter is post-M1)")
			}

			runOnce := func() {
				fmt.Println("Starting scrape pipeline...")
				if err := services.Scrape.Run(ctx); err != nil {
					log.Printf("scrape failed: %v", err)
					return
				}

				run, err := services.Scrape.LatestRun(ctx)
				if err != nil {
					log.Printf("warning: could not retrieve run stats: %v", err)
				} else {
					fmt.Printf("Scrape complete: added=%d updated=%d removed=%d\n",
						run.JobsAdded, run.JobsUpdated, run.JobsRemoved)
					if run.Error != "" {
						fmt.Printf("  Error: %s\n", run.Error)
					}
				}
			}

			runOnce()

			if !loop {
				return nil
			}

			log.Printf("scrape: loop mode — re-scraping every %s (SIGTERM to stop)", interval)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					log.Println("scrape: shutting down")
					return nil
				case <-ticker.C:
					runOnce()
				}
			}
		},
	}

	cmd.Flags().StringVar(&source, "source", "", "Scrape a single ATS source (greenhouse|lever|ashby)")
	cmd.Flags().BoolVar(&loop, "loop", false, "Run continuously, re-scraping on --interval")
	cmd.Flags().DurationVar(&interval, "interval", defaultInterval, "Interval between scrape runs (used with --loop); also reads SCRAPE_INTERVAL env")
	return cmd
}
