package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newImportCmd(services Services) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import LinkedIn Connections.csv",
		Long:  "Parse a LinkedIn data export CSV and import contacts. Companies are auto-matched to the job database.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.Contacts == nil {
				return fmt.Errorf("import is only available in local mode (requires database access)")
			}

			if filePath == "" {
				return fmt.Errorf("--file is required")
			}

			f, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open file: %w", err)
			}
			defer f.Close()

			// Resolve user from stored session token.
			token, err := ReadStoredToken()
			if err != nil || token == "" {
				return fmt.Errorf("not logged in — run 'jobs login' first")
			}

			user, err := services.Auth.Authenticate(cmd.Context(), token)
			if err != nil {
				return fmt.Errorf("session expired — run 'jobs login' again: %w", err)
			}

			result, err := services.Contacts.ImportCSV(cmd.Context(), user.ID, f)
			if err != nil {
				return fmt.Errorf("import: %w", err)
			}

			fmt.Printf("Imported %d contacts\n", result.ContactsImported)
			fmt.Printf("  %d companies linked (matched existing)\n", result.CompaniesLinked)
			fmt.Printf("  %d companies registered (new ATS boards found)\n", result.CompaniesRegistered)
			fmt.Printf("  %d companies unmatched\n", result.CompaniesUnmatched)

			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to LinkedIn Connections.csv")

	return cmd
}
