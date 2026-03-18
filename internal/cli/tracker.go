package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/justestif/go-jobs/internal/core/domain"

	jsonv2 "github.com/go-json-experiment/json"
)

// resolveUser reads the stored CLI token and resolves it to a domain.User via
// the SessionRepository. Returns a clear error if not logged in.
func resolveUser(ctx context.Context, services Services) (domain.User, error) {
	token, err := requireToken()
	if err != nil {
		return domain.User{}, err
	}
	user, err := services.Session.GetUserByToken(ctx, token)
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid or expired session — run 'go-jobs login': %w", err)
	}
	return user, nil
}

// parseJobID parses a UUID job ID from a string argument.
func parseJobID(raw string) (domain.JobID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return domain.JobID{}, fmt.Errorf("invalid job id %q: %w", raw, err)
	}
	return id, nil
}

// -----------------------------------------------------------------------
// go-jobs interested [<job-id>]
// -----------------------------------------------------------------------

// newInterestedCmd returns the `go-jobs interested` command.
//
// With no arguments: lists all jobs in StatusInterested as JSON.
// With a job ID argument: marks the job as interested.
func newInterestedCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interested [job-id]",
		Short: "Mark a job as interested, or list interested jobs",
		Long: `With no arguments, lists all jobs you have marked as interested (JSON).
With a job ID, marks that job as interested in your pipeline.`,
		Example: `  go-jobs interested
  go-jobs interested 018e1234-abcd-7000-8000-000000000001`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				// List mode.
				jobs, err := services.Application.ListByStatus(ctx, user.ID, domain.StatusInterested)
				if err != nil {
					return fmt.Errorf("list interested: %w", err)
				}
				return printJSON(cmd, jobs)
			}

			// Mark mode.
			jobID, err := parseJobID(args[0])
			if err != nil {
				return err
			}
			if err := services.Application.SetStatus(ctx, user.ID, jobID, domain.StatusInterested); err != nil {
				return fmt.Errorf("mark interested: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Marked as interested.")
			return nil
		},
	}
	return cmd
}

// -----------------------------------------------------------------------
// go-jobs apply <job-id>
// -----------------------------------------------------------------------

// newApplyCmd returns the `go-jobs apply` command.
func newApplyCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "apply <job-id>",
		Short:   "Mark a job as applied (auto-captures date; sets interested if not already)",
		Example: `  go-jobs apply 018e1234-abcd-7000-8000-000000000001`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}
			jobID, err := parseJobID(args[0])
			if err != nil {
				return err
			}
			if err := services.Application.SetStatus(ctx, user.ID, jobID, domain.StatusApplied); err != nil {
				return fmt.Errorf("mark applied: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Marked as applied.")
			return nil
		},
	}
	return cmd
}

// -----------------------------------------------------------------------
// go-jobs status <job-id> <status>
// -----------------------------------------------------------------------

var validStatuses = []string{
	string(domain.StatusInterested),
	string(domain.StatusApplied),
	string(domain.StatusInterviewing),
	string(domain.StatusOffer),
	string(domain.StatusRejected),
	string(domain.StatusWithdrawn),
}

// newStatusCmd returns the `go-jobs status` command.
func newStatusCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <job-id> <status>",
		Short: "Update the pipeline status of a job",
		Long: `Set the status of a job in your pipeline.

Valid statuses: ` + strings.Join(validStatuses, ", "),
		Example: `  go-jobs status 018e1234-abcd-7000-8000-000000000001 interviewing
  go-jobs status 018e1234-abcd-7000-8000-000000000001 rejected`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}
			jobID, err := parseJobID(args[0])
			if err != nil {
				return err
			}
			status := domain.ApplicationStatus(args[1])
			if !isValidStatus(status) {
				return fmt.Errorf("invalid status %q — valid values: %s", args[1], strings.Join(validStatuses, ", "))
			}
			if err := services.Application.SetStatus(ctx, user.ID, jobID, status); err != nil {
				return fmt.Errorf("set status: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status set to %q.\n", status)
			return nil
		},
	}
	return cmd
}

func isValidStatus(s domain.ApplicationStatus) bool {
	for _, v := range validStatuses {
		if string(s) == v {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------
// go-jobs notes <job-id> "<text>"
// -----------------------------------------------------------------------

// newNotesCmd returns the `go-jobs notes` command.
func newNotesCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notes <job-id> <text>",
		Short:   "Set notes on a tracked job",
		Example: `  go-jobs notes 018e1234-abcd-7000-8000-000000000001 "Strong match for backend role"`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}
			jobID, err := parseJobID(args[0])
			if err != nil {
				return err
			}
			if err := services.Application.SetNotes(ctx, user.ID, jobID, args[1]); err != nil {
				return fmt.Errorf("set notes: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Notes saved.")
			return nil
		},
	}
	return cmd
}

// -----------------------------------------------------------------------
// go-jobs applied
// -----------------------------------------------------------------------

// newAppliedCmd returns the `go-jobs applied` command.
func newAppliedCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "applied",
		Short: "List jobs you have applied to (JSON)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}
			jobs, err := services.Application.ListByStatus(ctx, user.ID, domain.StatusApplied)
			if err != nil {
				return fmt.Errorf("list applied: %w", err)
			}
			return printJSON(cmd, jobs)
		},
	}
	return cmd
}

// -----------------------------------------------------------------------
// go-jobs pipeline
// -----------------------------------------------------------------------

// pipelineOutput is the JSON shape for the pipeline command.
type pipelineOutput struct {
	Status string       `json:"status"`
	Jobs   []domain.Job `json:"jobs"`
}

// newPipelineCmd returns the `go-jobs pipeline` command.
func newPipelineCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "List all tracked jobs grouped by status (JSON)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			user, err := resolveUser(ctx, services)
			if err != nil {
				return err
			}
			grouped, err := services.Application.ListPipeline(ctx, user.ID)
			if err != nil {
				return fmt.Errorf("list pipeline: %w", err)
			}

			// Emit groups in a stable pipeline order.
			order := []domain.ApplicationStatus{
				domain.StatusInterested,
				domain.StatusApplied,
				domain.StatusInterviewing,
				domain.StatusOffer,
				domain.StatusRejected,
				domain.StatusWithdrawn,
			}
			out := make([]pipelineOutput, 0, len(order))
			for _, status := range order {
				jobs := grouped[status]
				if len(jobs) == 0 {
					continue
				}
				out = append(out, pipelineOutput{
					Status: string(status),
					Jobs:   jobs,
				})
			}

			b, err := jsonv2.Marshal(out)
			if err != nil {
				return fmt.Errorf("marshal pipeline: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
	return cmd
}
