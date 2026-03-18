package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// ErrInvalidCredentials is returned by Login when the email/password pair does
// not match a known user.
var ErrInvalidCredentials = errors.New("invalid email or password")

// authService implements ports.AuthService.
type authService struct {
	users    ports.UserRepository
	sessions ports.SessionRepository
}

// NewAuthService constructs an AuthService.
// users and sessions may be the same concrete adapter (UserRepo satisfies both).
func NewAuthService(users ports.UserRepository, sessions ports.SessionRepository) ports.AuthService {
	return &authService{users: users, sessions: sessions}
}

// Register creates a new user account. The password is hashed with bcrypt
// before storage. Returns ErrInvalidCredentials if email is already taken
// (wrapped as a friendly message).
func (s *authService) Register(ctx context.Context, email, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		return domain.User{}, fmt.Errorf("register %q: %w", email, err)
	}

	return user, nil
}

// Login verifies the email/password and, on success, generates and persists an
// opaque random token that the caller can store (cookie or file). Returns
// ErrInvalidCredentials for unknown email or wrong password.
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Avoid leaking whether the email exists.
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	if err := s.sessions.SaveToken(ctx, user.ID, token); err != nil {
		return "", fmt.Errorf("save session token: %w", err)
	}

	return token, nil
}

// Logout deletes the opaque token from the session store.
func (s *authService) Logout(ctx context.Context, token string) error {
	if err := s.sessions.DeleteToken(ctx, token); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

// Authenticate resolves an opaque token to the associated User.
// Used by HTTP middleware (reads session cookie value) and CLI
// (reads token file) on every request that requires identity.
func (s *authService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	user, err := s.sessions.GetUserByToken(ctx, token)
	if err != nil {
		return domain.User{}, fmt.Errorf("authenticate: %w", err)
	}
	return user, nil
}

// generateToken creates a 32-byte cryptographically random hex string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
