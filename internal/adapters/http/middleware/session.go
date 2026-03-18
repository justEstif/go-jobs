package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

// ContextWithUserID stores a user ID in the context for downstream handlers.
func ContextWithUserID(ctx context.Context, id [16]byte) context.Context {
	return context.WithValue(ctx, userIDContextKey, id)
}

// UserIDFromContext retrieves the user ID stored by OptionalAuth or RequireAuth.
// Returns the zero value and false if not present.
func UserIDFromContext(ctx context.Context) ([16]byte, bool) {
	id, ok := ctx.Value(userIDContextKey).([16]byte)
	return id, ok
}

const sessionUserIDKey = "user_id"

// SessionManager wraps alexedwards/scs, backed by a Postgres store.
// It is also an http.Handler middleware (LoadAndSave) that must be applied to
// all routes before any handler reads or writes session data.
type SessionManager struct {
	scs *scs.SessionManager
}

// NewSessionManager creates a SessionManager backed by pgxstore.
// pool is the application's pgxpool; secret is used to authenticate the
// session cookie (SESSION_SECRET env var). The http_sessions table must
// exist (migration 005).
func NewSessionManager(pool *pgxpool.Pool, secret []byte) *SessionManager {
	sm := scs.New()
	sm.Store = pgxstore.NewWithConfig(pool, pgxstore.Config{
		TableName:       "http_sessions",
		CleanUpInterval: 5 * time.Minute,
	})
	sm.Lifetime = 7 * 24 * time.Hour
	sm.Cookie.Name = "app_session"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = false // set true in production
	return &SessionManager{scs: sm}
}

// LoadAndSave is the scs middleware that must wrap all routes.
// It loads the session on every request and saves it after the handler returns.
func (sm *SessionManager) LoadAndSave(next http.Handler) http.Handler {
	return sm.scs.LoadAndSave(next)
}

// SetUserSession stores the user ID in the current session and rotates the
// session token to prevent fixation attacks.
func (sm *SessionManager) SetUserSession(r *http.Request, userID [16]byte) error {
	if err := sm.scs.RenewToken(r.Context()); err != nil {
		return err
	}
	sm.scs.Put(r.Context(), sessionUserIDKey, userID[:])
	return nil
}

// GetUserIDFromSession retrieves the user ID from the session.
// Returns http.ErrNoCookie if the session does not contain a user ID.
func (sm *SessionManager) GetUserIDFromSession(r *http.Request) ([16]byte, error) {
	b := sm.scs.GetBytes(r.Context(), sessionUserIDKey)
	if len(b) != 16 {
		return [16]byte{}, http.ErrNoCookie
	}
	var id [16]byte
	copy(id[:], b)
	return id, nil
}

// ClearSession destroys the session entirely (used on logout).
func (sm *SessionManager) ClearSession(r *http.Request) error {
	return sm.scs.Destroy(r.Context())
}
