package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// resolveUserID reads the stored CLI token and resolves it to a user ID.
func resolveUserID(ctx context.Context, services Services) (uuid.UUID, error) {
	user, err := resolveUser(ctx, services)
	if err != nil {
		return uuid.Nil, err
	}
	return user.ID, nil
}

func newResumeCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Manage your resume for Job Coach",
	}

	cmd.AddCommand(newResumeSetCmd(services))
	cmd.AddCommand(newResumeShowCmd(services))
	cmd.AddCommand(newResumeClearCmd(services))

	return cmd
}

func newResumeSetCmd(services Services) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set your resume from file or stdin",
		Long: `Set your resume from a file or stdin for Job Coach analysis.

Examples:
  jobs resume set --file resume.md
  jobs resume set < resume.md
  cat resume.md | jobs resume set`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.User == nil {
				return fmt.Errorf("resume command requires local mode (direct DB access)")
			}

			userID, err := resolveUserID(cmd.Context(), services)
			if err != nil {
				return err
			}

			var data []byte
			if file != "" {
				data, err = os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
			} else {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
			}

			resume := string(data)
			if resume == "" {
				return fmt.Errorf("resume is empty")
			}
			if len(resume) > domain.MaxResumeLength {
				return fmt.Errorf("resume too long: %d characters exceeds maximum of %d", len(resume), domain.MaxResumeLength)
			}

			if err := services.User.SetResume(cmd.Context(), userID, resume); err != nil {
				return err
			}

			fmt.Printf("Resume saved (%d chars)\n", len(resume))
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to resume file")

	return cmd
}

func newResumeShowCmd(services Services) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display your current resume",
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.User == nil {
				return fmt.Errorf("resume command requires local mode (direct DB access)")
			}

			userID, err := resolveUserID(cmd.Context(), services)
			if err != nil {
				return err
			}

			user, err := services.User.GetByID(cmd.Context(), userID)
			if err != nil {
				return err
			}

			if user.Resume == "" {
				fmt.Println("No resume configured. Use: jobs resume set --file resume.md")
				return nil
			}

			fmt.Print(user.Resume)
			return nil
		},
	}
}

func newResumeClearCmd(services Services) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove your stored resume",
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.User == nil {
				return fmt.Errorf("resume command requires local mode (direct DB access)")
			}

			userID, err := resolveUserID(cmd.Context(), services)
			if err != nil {
				return err
			}

			if err := services.User.SetResume(cmd.Context(), userID, ""); err != nil {
				return err
			}

			fmt.Println("Resume cleared.")
			return nil
		},
	}
}
