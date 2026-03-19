package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

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
// before storage. Returns an error if the email is already taken.
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
// ports.ErrInvalidCredentials for unknown email or wrong password.
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// Avoid leaking whether the email exists.
		return "", ports.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ports.ErrInvalidCredentials
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

// ChangePassword verifies the current password and sets a new one.
// Returns ports.ErrInvalidCredentials if the current password is wrong.
func (s *authService) ChangePassword(ctx context.Context, userID domain.UserID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ports.ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("change password: hash: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return fmt.Errorf("change password: update: %w", err)
	}
	return nil
}

// DeleteAccount permanently removes the user and all associated data.
// Requires the current password for confirmation.
// Returns ports.ErrInvalidCredentials if the password is wrong.
func (s *authService) DeleteAccount(ctx context.Context, userID domain.UserID, password string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete account: get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ports.ErrInvalidCredentials
	}

	if err := s.users.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// generateToken creates a 32-byte cryptographically random hex string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
