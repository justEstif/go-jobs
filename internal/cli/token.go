package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// tokenFilePath returns the path to the stored CLI session token.
// Uses $XDG_CONFIG_HOME/go-jobs/token, falling back to $HOME/.config/go-jobs/token.
func tokenFilePath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "go-jobs", "token"), nil
}

// readToken reads the stored CLI session token from disk.
// Returns an empty string (not an error) if no token file exists.
func readToken() (string, error) {
	path, err := tokenFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// writeToken persists a session token to disk, creating parent directories
// as needed.
func writeToken(token string) error {
	path, err := tokenFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(token), 0o600); err != nil {
		return fmt.Errorf("write token: %w", err)
	}
	return nil
}

// deleteToken removes the stored session token from disk.
// No-op if the file does not exist.
func deleteToken() error {
	path, err := tokenFilePath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete token: %w", err)
	}
	return nil
}

// requireToken reads the stored token and returns an error guiding the user
// to log in if no token is found.
func requireToken() (string, error) {
	token, err := readToken()
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", fmt.Errorf("not logged in — run 'go-jobs login' first")
	}
	return token, nil
}
