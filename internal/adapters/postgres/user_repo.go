package postgres

import (
	"context"
	"fmt"

	"github.com/justestif/go-jobs/internal/adapters/postgres/queries"
	"github.com/justestif/go-jobs/internal/core/domain"
)

// UserRepo implements ports.UserRepository and ports.SessionRepository
// against PostgreSQL. The composition root may wire both ports to this
// single struct.
type UserRepo struct {
	q *queries.Queries
}

// NewUserRepo constructs a UserRepo backed by the given Queries handle.
func NewUserRepo(q *queries.Queries) *UserRepo {
	return &UserRepo{q: q}
}

// Create inserts a new user with the given email and bcrypt password hash.
func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, queries.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user %q: %w", email, err)
	}
	return domainUserFromDB(row), nil
}

// GetByEmail retrieves a user by email address.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email %q: %w", email, err)
	}
	return domainUserFromDB(row), nil
}

// GetByID retrieves a user by their UUID primary key.
func (r *UserRepo) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, uuidToPg(id))
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id %s: %w", id, err)
	}
	return domainUserFromDB(row), nil
}

// Update writes mutable user fields back to the DB.
func (r *UserRepo) Update(ctx context.Context, user domain.User) error {
	err := r.q.UpdateUser(ctx, queries.UpdateUserParams{
		ID:            uuidToPg(user.ID),
		LlmApiKey:     user.LLMAPIKey,
		LlmProvider:   string(user.LLMProvider),
		LlmModel:      user.LLMModel,
		LlmBaseUrl:    user.LLMBaseURL,
		Resume:        user.Resume,
		LastVisitedAt: timePtrToPg(user.LastVisitedAt),
	})
	if err != nil {
		return fmt.Errorf("update user %s: %w", user.ID, err)
	}
	return nil
}

// UpdatePassword sets a new bcrypt password hash for the user.
func (r *UserRepo) UpdatePassword(ctx context.Context, userID domain.UserID, passwordHash string) error {
	err := r.q.UpdatePassword(ctx, queries.UpdatePasswordParams{
		ID:           uuidToPg(userID),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("update password for user %s: %w", userID, err)
	}
	return nil
}

// Delete permanently removes a user from the database.
func (r *UserRepo) Delete(ctx context.Context, userID domain.UserID) error {
	err := r.q.DeleteUser(ctx, uuidToPg(userID))
	if err != nil {
		return fmt.Errorf("delete user %s: %w", userID, err)
	}
	return nil
}

// TouchLastVisited updates last_visited_at to NOW() for the given user.
func (r *UserRepo) TouchLastVisited(ctx context.Context, userID domain.UserID) error {
	err := r.q.TouchUserLastVisited(ctx, uuidToPg(userID))
	if err != nil {
		return fmt.Errorf("touch last visited for user %s: %w", userID, err)
	}
	return nil
}

// SaveToken associates token with userID in the sessions table.
func (r *UserRepo) SaveToken(ctx context.Context, userID domain.UserID, token string) error {
	err := r.q.SaveSession(ctx, queries.SaveSessionParams{
		Token:  token,
		UserID: uuidToPg(userID),
	})
	if err != nil {
		return fmt.Errorf("save session token for user %s: %w", userID, err)
	}
	return nil
}

// DeleteToken removes a session token from the sessions table.
func (r *UserRepo) DeleteToken(ctx context.Context, token string) error {
	err := r.q.DeleteSession(ctx, token)
	if err != nil {
		return fmt.Errorf("delete session token: %w", err)
	}
	return nil
}

// GetUserByToken resolves a session token to the associated User.
func (r *UserRepo) GetUserByToken(ctx context.Context, token string) (domain.User, error) {
	row, err := r.q.GetUserByToken(ctx, token)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by token: %w", err)
	}
	return domainUserFromDB(row), nil
}
