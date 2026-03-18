package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/justestif/go-jobs/internal/core/ports"
)

const bearerTokenContextKey contextKey = "bearer_token"

// BearerAuth validates the Authorization: Bearer <token> header on each
// request. On success it stores both the resolved user ID and the raw token in
// the request context, then calls the next handler. On failure it responds
// 401 Unauthorized.
//
// Retrieve the raw token in a handler with:
//
//	token, ok := TokenFromContext(r.Context())
func BearerAuth(auth ports.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := extractBearer(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := auth.Authenticate(r.Context(), token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := ContextWithUserID(r.Context(), user.ID)
			ctx = context.WithValue(ctx, bearerTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TokenFromContext retrieves the raw Bearer token stored by BearerAuth.
func TokenFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(bearerTokenContextKey)
	token, ok := v.(string)
	return token, ok && token != ""
}

// extractBearer parses the raw token from the Authorization header.
func extractBearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	token := strings.TrimPrefix(h, "Bearer ")
	if token == "" {
		return "", false
	}
	return token, true
}
