package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newServeCmd returns the `go-jobs serve` command.
func newServeCmd(services Services) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.Serve == nil {
				return fmt.Errorf("serve command not configured")
			}
			return services.Serve(context.Background())
		},
	}
}
