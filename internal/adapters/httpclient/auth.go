package httpclient

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// AuthClient implements ports.AuthService and ports.SessionRepository against
// the /api/v1/auth/ endpoints.
type AuthClient struct {
	c *Client
}

// NewAuthClient constructs an AuthClient.
func NewAuthClient(c *Client) *AuthClient {
	return &AuthClient{c: c}
}

// Register calls POST /api/v1/auth/register. Does not require a token.
func (a *AuthClient) Register(ctx context.Context, email, password string) (domain.User, error) {
	body := map[string]string{"email": email, "password": password}
	var resp struct {
		Email string `json:"email"`
	}
	if err := a.c.post(ctx, "auth/register", false, body, &resp); err != nil {
		return domain.User{}, fmt.Errorf("register: %w", err)
	}
	return domain.User{Email: resp.Email}, nil
}

// Login calls POST /api/v1/auth/login. Returns the opaque Bearer token.
// Does not require an existing token.
func (a *AuthClient) Login(ctx context.Context, email, password string) (string, error) {
	body := map[string]string{"email": email, "password": password}
	var resp struct {
		Token string `json:"token"`
	}
	if err := a.c.post(ctx, "auth/login", false, body, &resp); err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	return resp.Token, nil
}

// Logout calls POST /api/v1/auth/logout with the client's Bearer token.
// The token argument is accepted for interface compatibility but the client's
// stored token is what is sent in the Authorization header.
func (a *AuthClient) Logout(ctx context.Context, _ string) error {
	if err := a.c.post(ctx, "auth/logout", true, nil, nil); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

// Authenticate calls GET /api/v1/auth/me. Used by the CLI to resolve the
// stored token into a domain.User before making authenticated service calls.
func (a *AuthClient) Authenticate(ctx context.Context, _ string) (domain.User, error) {
	return a.Me(ctx)
}

// GetUserByToken implements ports.SessionRepository. Calls GET /api/v1/auth/me
// to resolve the stored token into a domain.User.
// The token argument is accepted for interface compatibility; the client's
// stored token is used in the Authorization header.
func (a *AuthClient) GetUserByToken(ctx context.Context, _ string) (domain.User, error) {
	return a.Me(ctx)
}

// SaveToken and DeleteToken are server-side concerns; they are no-ops in the
// HTTP client adapter — the server manages token persistence.

// SaveToken implements ports.SessionRepository. No-op in remote mode.
func (a *AuthClient) SaveToken(_ context.Context, _ domain.UserID, _ string) error {
	return nil
}

// DeleteToken implements ports.SessionRepository. No-op in remote mode.
func (a *AuthClient) DeleteToken(_ context.Context, _ string) error {
	return nil
}

// ChangePassword is not supported in remote CLI mode — use the web settings page.
func (a *AuthClient) ChangePassword(_ context.Context, _ domain.UserID, _, _ string) error {
	return fmt.Errorf("ChangePassword: not supported in remote mode — use the web settings page")
}

// DeleteAccount is not supported in remote CLI mode — use the web settings page.
func (a *AuthClient) DeleteAccount(_ context.Context, _ domain.UserID, _ string) error {
	return fmt.Errorf("DeleteAccount: not supported in remote mode — use the web settings page")
}

// Me calls GET /api/v1/auth/me and returns the authenticated user.
func (a *AuthClient) Me(ctx context.Context) (domain.User, error) {
	var resp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := a.c.get(ctx, "auth/me", nil, true, &resp); err != nil {
		return domain.User{}, fmt.Errorf("auth/me: %w", err)
	}
	id, err := uuid.Parse(resp.ID)
	if err != nil {
		return domain.User{}, fmt.Errorf("parse user id: %w", err)
	}
	return domain.User{ID: id, Email: resp.Email}, nil
}
