package middleware

import (
	"net/http"

	"github.com/gorilla/csrf"
)

// SetupCSRF creates CSRF protection middleware
//
// Usage in main.go:
//
//	csrfMw := middleware.SetupCSRF([]byte(os.Getenv("CSRF_KEY")), false)
//	r.Use(csrfMw)
//
// In templates, access token with:
//
//	csrf.Token(r)
//
// IMPORTANT:
// - CSRF_KEY must be 32 bytes long
// - Set secure=true in production (HTTPS only)
// - Token automatically validated on POST/PUT/DELETE requests
// - Token field name is "gorilla.csrf.Token"
//
// SameSite note: CSRF uses Strict; session cookies (see internal/auth/session.go)
// use Lax. This is intentional. The CSRF cookie only needs to be sent on
// same-site requests, so Strict is correct. The session cookie must be Lax so
// it is sent when the OAuth provider redirects back to your app (a cross-site
// top-level navigation) — Strict would drop the session cookie on that redirect,
// breaking the logged-in state after OAuth login.
func SetupCSRF(key []byte, secure bool) func(http.Handler) http.Handler {
	return csrf.Protect(
		key,
		csrf.Secure(secure), // Only send over HTTPS in production
		csrf.Path("/"),      // Cookie path
		csrf.SameSite(csrf.SameSiteStrictMode),
	)
}
