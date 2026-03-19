package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newBackfillTagsCmd returns a hidden one-time command to enrich the existing
// backlog of un-tagged jobs. Not shown in --help output.
func newBackfillTagsCmd(services Services) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:    "backfill-tags",
		Hidden: true,
		Short:  "One-time backfill: enrich all un-tagged jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.BackfillTags == nil {
				return fmt.Errorf("backfill-tags requires direct DB access (run without --base-url, or with --base-url local)")
			}
			ctx := context.Background()
			enriched, failed, err := services.BackfillTags(ctx, limit)
			if err != nil {
				return fmt.Errorf("backfill-tags: %w", err)
			}
			fmt.Printf("Backfill complete: enriched=%d failed=%d\n", enriched, failed)
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100000, "Maximum number of jobs to enrich")
	return cmd
}
