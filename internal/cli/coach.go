package cli

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/justestif/go-jobs/internal/core/services"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd(services Services) *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:   "analyze <job-id>",
		Short: "Analyze a job against your resume (requires LLM provider)",
		Long: `Run Job Coach analysis: ATS keyword gaps, fit assessment, and an
optimized resume tailored to the specified job. Requires a resume and
LLM provider configured (via web settings or direct DB).

Results are cached per job — use --refresh to force a new analysis.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.Coach == nil {
				return fmt.Errorf("analyze command requires local mode (direct DB access)")
			}

			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			userID, err := resolveUserID(cmd.Context(), services)
			if err != nil {
				return err
			}

			fmt.Println("Analyzing...")
			result, err := services.Coach.AnalyzeJob(cmd.Context(), userID, jobID, refresh)
			if err != nil {
				return err
			}

			fmt.Print(result)
			return nil
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "Bypass cache and force a new analysis")

	return cmd
}

func newSystemPromptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "system-prompt",
		Short: "Print the Job Coach system prompt",
		Long:  `Print the raw system prompt used by Job Coach. No arguments, no DB, no LLM provider needed.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(services.AnalyzeSystemPrompt())
			return nil
		},
	}
}

func newPromptCmd(services Services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt <job-id>",
		Short: "Output the raw Job Coach prompt (no LLM call)",
		Long: `Print the system prompt and user prompt that Job Coach would send to an
LLM for the specified job. Pipe this to your own LLM CLI or copy it
into ChatGPT, Claude, etc.

Only requires a resume — no LLM provider needed.

Examples:
  jobs prompt abc123                       # print to terminal
  jobs prompt abc123 | pbcopy              # copy to clipboard
  jobs prompt abc123 | llm -m gpt-4o       # pipe to llm CLI
  jobs prompt abc123 > prompt.md           # save to file`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if services.Coach == nil {
				return fmt.Errorf("prompt command requires local mode (direct DB access)")
			}

			jobID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid job ID: %w", err)
			}

			userID, err := resolveUserID(cmd.Context(), services)
			if err != nil {
				return err
			}

			systemPrompt, userPrompt, err := services.Coach.BuildAnalyzePrompt(cmd.Context(), userID, jobID)
			if err != nil {
				return err
			}

			fmt.Println("--- SYSTEM PROMPT ---")
			fmt.Println(systemPrompt)
			fmt.Println()
			fmt.Println("--- USER PROMPT ---")
			fmt.Println(userPrompt)

			return nil
		},
	}

	return cmd
}
