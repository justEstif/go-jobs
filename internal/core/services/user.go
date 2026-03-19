package services

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// KeyEncryptor encrypts API keys before storage. Injected by the composition root.
type KeyEncryptor func(plaintext string) (string, error)

// userService implements ports.UserService.
type userService struct {
	users      ports.UserRepository
	userJobs   ports.UserJobRepository
	encryptKey KeyEncryptor
}

// NewUserService constructs a UserService backed by the given repositories.
// encryptKey encrypts API keys before they are persisted. Pass a no-op function
// if encryption is not configured (e.g. in tests).
func NewUserService(users ports.UserRepository, userJobs ports.UserJobRepository, encryptKey KeyEncryptor) ports.UserService {
	return &userService{users: users, userJobs: userJobs, encryptKey: encryptKey}
}

// SetLLMConfig stores the user's LLM provider configuration.
// The API key is encrypted before storage. For Ollama, apiKey may be empty.
func (s *userService) SetLLMConfig(ctx context.Context, userID domain.UserID, provider domain.LLMProvider, apiKey, model, baseURL string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("set llm config: get user: %w", err)
	}

	// Encrypt the API key before storage. Empty keys stay empty.
	encrypted := ""
	if apiKey != "" {
		encrypted, err = s.encryptKey(apiKey)
		if err != nil {
			return fmt.Errorf("set llm config: encrypt key: %w", err)
		}
	}

	user.LLMAPIKey = encrypted
	user.LLMProvider = provider
	user.LLMModel = model
	user.LLMBaseURL = baseURL

	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("set llm config: update user: %w", err)
	}
	return nil
}

// SetResume stores the user's resume for Job Coach analysis.
// Returns an error if the resume exceeds domain.MaxResumeLength characters.
func (s *userService) SetResume(ctx context.Context, userID domain.UserID, resume string) error {
	if len(resume) > domain.MaxResumeLength {
		return fmt.Errorf("resume too long: %d characters exceeds maximum of %d", len(resume), domain.MaxResumeLength)
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("set resume: get user: %w", err)
	}
	user.Resume = resume
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("set resume: update user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by their UUID primary key.
func (s *userService) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id %s: %w", id, err)
	}
	return user, nil
}

// TouchLastVisited updates the user's LastVisitedAt timestamp to now.
func (s *userService) TouchLastVisited(ctx context.Context, userID domain.UserID) error {
	if err := s.users.TouchLastVisited(ctx, userID); err != nil {
		return fmt.Errorf("touch last visited for user %s: %w", userID, err)
	}
	return nil
}

// ChangePassword verifies the current password and sets a new one.
// Returns ErrInvalidCredentials if the current password is wrong.
func (s *userService) ChangePassword(ctx context.Context, userID domain.UserID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
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

// ResetTracker deletes all pipeline state (user_jobs) for the user.
func (s *userService) ResetTracker(ctx context.Context, userID domain.UserID) error {
	if err := s.userJobs.DeleteAll(ctx, userID); err != nil {
		return fmt.Errorf("reset tracker for user %s: %w", userID, err)
	}
	return nil
}

// DeleteAccount permanently removes the user and all associated data.
// Requires the current password for confirmation.
func (s *userService) DeleteAccount(ctx context.Context, userID domain.UserID, password string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("delete account: get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	if err := s.users.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}
