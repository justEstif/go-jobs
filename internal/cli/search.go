package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/justestif/go-jobs/internal/core/domain"

	jsonv2 "github.com/go-json-experiment/json"
)

// newSearchCmd returns the `go-jobs search` command.
func newSearchCmd(services Services) *cobra.Command {
	var (
		query          string
		roles          []string
		seniorities    []string
		remotePolicies []string
		countries      []string
		techStack      []string
		companyIDs     []string
		limit          int
		offset         int
		format         string
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search job postings with optional filters",
		Long: `Search active job postings. Returns JSON by default (agent-friendly).
Use --format table for a human-readable view.

Filter semantics:
  --role, --seniority, --remote, --country  OR within the field
  --tech                                    AND — job must mention ALL terms
  --company                                 OR — match any company ID`,
		Example: `  go-jobs search
  go-jobs search --query "backend engineer"
  go-jobs search --role engineering --seniority senior --seniority mid
  go-jobs search --remote remote --country US
  go-jobs search --tech Go --tech PostgreSQL
  go-jobs search --limit 20 --format table`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Convert string flags to domain types.
			roleTypes := make([]domain.RoleType, len(roles))
			for i, r := range roles {
				roleTypes[i] = domain.RoleType(r)
			}
			sens := make([]domain.Seniority, len(seniorities))
			for i, s := range seniorities {
				sens[i] = domain.Seniority(s)
			}
			remote := make([]domain.WorkplaceType, len(remotePolicies))
			for i, r := range remotePolicies {
				remote[i] = domain.WorkplaceType(r)
			}
			compIDs := make([]domain.CompanyID, 0, len(companyIDs))
			for _, raw := range companyIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("invalid company id %q: %w", raw, err)
				}
				compIDs = append(compIDs, id)
			}

			filters := domain.SearchFilters{
				Query:        query,
				RoleTypes:    roleTypes,
				Seniorities:  sens,
				RemotePolicy: remote,
				Countries:    countries,
				TechStack:    techStack,
				CompanyIDs:   compIDs,
				Limit:        limit,
				Offset:       offset,
			}

			jobs, err := services.Search.Search(ctx, filters, nil)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			switch strings.ToLower(format) {
			case "table":
				return printTable(cmd, jobs)
			default:
				return printJSON(cmd, jobs)
			}
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Free-text search on title or company name")
	cmd.Flags().StringArrayVar(&roles, "role", nil, "Filter by role type (engineering, design, product, data, ...)")
	cmd.Flags().StringArrayVar(&seniorities, "seniority", nil, "Filter by seniority (intern, junior, mid, senior, staff, lead)")
	cmd.Flags().StringArrayVar(&remotePolicies, "remote", nil, "Filter by remote policy (remote, hybrid, onsite)")
	cmd.Flags().StringArrayVar(&countries, "country", nil, "Filter by country ISO code (US, DE, BR, ...)")
	cmd.Flags().StringArrayVar(&techStack, "tech", nil, "Filter by tech (AND — job must mention all specified terms)")
	cmd.Flags().StringArrayVar(&companyIDs, "company", nil, "Filter by company UUID (repeatable)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Result offset for pagination")
	cmd.Flags().StringVar(&format, "format", "json", "Output format: json (default) or table")

	return cmd
}

// printJSON marshals jobs to JSON and writes them to stdout.
func printJSON(cmd *cobra.Command, jobs []domain.Job) error {
	b, err := jsonv2.Marshal(jobs)
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

// printTable writes a human-readable tabular view of jobs.
func printTable(cmd *cobra.Command, jobs []domain.Job) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TITLE\tCOMPANY\tLOCATION\tROLE\tSENIORITY\tREMOTE\tCOUNTRY\tTECH")
	for _, j := range jobs {
		role, seniority, remote, country, tech := "", "", "", "", ""
		if j.Tags != nil {
			role = string(j.Tags.RoleType)
			seniority = string(j.Tags.Seniority)
			remote = string(j.Tags.RemotePolicy)
			country = j.Tags.Country
			if len(j.Tags.TechStack) > 0 {
				tech = strings.Join(j.Tags.TechStack, ",")
				if len(tech) > 40 {
					tech = tech[:37] + "..."
				}
			}
		}
		title := j.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			title, j.CompanyName, j.Location, role, seniority, remote, country, tech)
	}
	return w.Flush()
}
