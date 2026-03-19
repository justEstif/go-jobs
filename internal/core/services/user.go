package services

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// KeyEncryptor encrypts API keys before storage. Injected by the composition root.
type KeyEncryptor func(plaintext string) (string, error)

// userService implements ports.UserService.
type userService struct {
	users      ports.UserRepository
	encryptKey KeyEncryptor
}

// NewUserService constructs a UserService backed by the given UserRepository.
// encryptKey encrypts API keys before they are persisted. Pass a no-op function
// if encryption is not configured (e.g. in tests).
func NewUserService(users ports.UserRepository, encryptKey KeyEncryptor) ports.UserService {
	return &userService{users: users, encryptKey: encryptKey}
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
func (s *userService) SetResume(ctx context.Context, userID domain.UserID, resume string) error {
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
