package middleware

import (
	"net/http"
)

// OptionalAuth loads the session user ID into the request context if a valid
// session exists, but does not block unauthenticated requests. Use this on
// pages that render differently for logged-in vs anonymous visitors (e.g. the
// landing page navbar showing "Sign in" vs an avatar).
//
// Retrieve the session user ID in a handler with:
//
//	id, ok := UserIDFromContext(r.Context())
func OptionalAuth(sm *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := sm.GetUserIDFromSession(r)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := ContextWithUserID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth rejects unauthenticated requests with a redirect to /login.
// Place this on any route group that requires a logged-in user.
func RequireAuth(sm *SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := sm.GetUserIDFromSession(r); err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
