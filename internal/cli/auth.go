package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newRegisterCmd creates the `go-jobs register` command.
// Accepts --email and --password flags, or prompts interactively when omitted.
func newRegisterCmd(services Services) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				fmt.Print("Email: ")
				if _, err := fmt.Scanln(&email); err != nil {
					return fmt.Errorf("read email: %w", err)
				}
			}
			if password == "" {
				fmt.Print("Password: ")
				if _, err := fmt.Scanln(&password); err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			ctx := context.Background()
			user, err := services.Auth.Register(ctx, email, password)
			if err != nil {
				return fmt.Errorf("register: %w", err)
			}

			fmt.Printf("Account created: %s\n", user.Email)
			fmt.Println("Run `go-jobs login` to sign in.")
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Account email address")
	cmd.Flags().StringVar(&password, "password", "", "Account password")
	return cmd
}

// newLoginCmd creates the `go-jobs login` command.
// On success the opaque session token is written to $XDG_CONFIG_HOME/go-jobs/token.
func newLoginCmd(services Services) *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in and save session token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				fmt.Print("Email: ")
				if _, err := fmt.Scanln(&email); err != nil {
					return fmt.Errorf("read email: %w", err)
				}
			}
			if password == "" {
				fmt.Print("Password: ")
				if _, err := fmt.Scanln(&password); err != nil {
					return fmt.Errorf("read password: %w", err)
				}
			}

			ctx := context.Background()
			token, err := services.Auth.Login(ctx, email, password)
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}

			if err := writeToken(token); err != nil {
				return fmt.Errorf("save token: %w", err)
			}

			fmt.Printf("Logged in as %s\n", email)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Account email address")
	cmd.Flags().StringVar(&password, "password", "", "Account password")
	return cmd
}

// newLogoutCmd creates the `go-jobs logout` command.
// Deletes the server-side session token and removes the local token file.
func newLogoutCmd(services Services) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign out and delete session token",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := readToken()
			if err != nil {
				return err
			}
			if token == "" {
				fmt.Println("Not currently logged in.")
				return nil
			}

			ctx := context.Background()
			// Best-effort: delete server-side token, then always remove local file.
			if err := services.Auth.Logout(ctx, token); err != nil {
				fmt.Printf("Warning: could not delete server-side token: %v\n", err)
			}

			if err := deleteToken(); err != nil {
				return fmt.Errorf("delete local token: %w", err)
			}

			fmt.Println("Logged out.")
			return nil
		},
	}
}
