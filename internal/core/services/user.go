package services

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/core/domain"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// userService implements ports.UserService.
type userService struct {
	users ports.UserRepository
}

// NewUserService constructs a UserService backed by the given UserRepository.
func NewUserService(users ports.UserRepository) ports.UserService {
	return &userService{users: users}
}

// SetLLMKey stores an encrypted LLM API key and provider for the given user.
func (s *userService) SetLLMKey(ctx context.Context, userID domain.UserID, provider domain.LLMProvider, apiKey string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("set llm key: get user: %w", err)
	}
	user.LLMAPIKey = apiKey
	user.LLMProvider = provider
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("set llm key: update user: %w", err)
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
