package middleware

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
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

// Register [16]byte so gorilla/sessions can serialize uuid.UUID values
// (uuid.UUID is [16]byte) via encoding/gob. Without this, storing a UUID in
// a session silently fails at runtime when the cookie is decoded on the next
// request.
func init() {
	gob.Register([16]byte{})
}

const (
	SessionName   = "app_session"
	SessionUserID = "user_id"
)

// SessionManager wraps a gorilla/sessions store.
type SessionManager struct {
	store sessions.Store
}

// NewSessionManager creates a SessionManager backed by a signed cookie store.
// secret should come from SESSION_SECRET env var.
// Set Secure: true in production (HTTPS only).
func NewSessionManager(secret []byte) *SessionManager {
	store := sessions.NewCookieStore(secret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		// SameSite: Lax (not Strict) so the session cookie is sent on OAuth
		// redirect callbacks from the provider back to your app. Strict would
		// drop the cookie on those cross-site redirects, breaking the flow.
		// The CSRF middleware uses SameSite: Strict separately for its own cookie.
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // set true in production
	}
	return &SessionManager{store: store}
}

// SetUserSession stores the user ID (as [16]byte) in the session.
func (sm *SessionManager) SetUserSession(w http.ResponseWriter, r *http.Request, userID [16]byte) error {
	session, err := sm.store.Get(r, SessionName)
	if err != nil {
		return err
	}
	session.Values[SessionUserID] = userID
	return session.Save(r, w)
}

// GetUserIDFromSession retrieves the user ID from the session.
// Returns http.ErrNoCookie if no valid session exists.
func (sm *SessionManager) GetUserIDFromSession(r *http.Request) ([16]byte, error) {
	session, err := sm.store.Get(r, SessionName)
	if err != nil {
		return [16]byte{}, err
	}
	id, ok := session.Values[SessionUserID].([16]byte)
	if !ok {
		return [16]byte{}, http.ErrNoCookie
	}
	return id, nil
}

// ClearSession invalidates the session cookie.
func (sm *SessionManager) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := sm.store.Get(r, SessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
