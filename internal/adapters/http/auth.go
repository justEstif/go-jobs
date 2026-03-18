package httphandlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
	"github.com/justestif/go-jobs/internal/core/services"
)

// AuthHandler handles registration, login, and logout for the web layer.
type AuthHandler struct {
	auth ports.AuthService
	sm   *middleware.SessionManager
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(auth ports.AuthService, sm *middleware.SessionManager) *AuthHandler {
	return &AuthHandler{auth: auth, sm: sm}
}

// ShowRegister renders the registration form.
func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	csrfToken := csrf.Token(r)
	components.RegisterPage("", csrfToken).Render(r.Context(), w)
}

// Register handles POST /register. On success it logs the user in and
// redirects to the home page.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.auth.Register(r.Context(), email, password)
	if err != nil {
		csrfTokenErr := csrf.Token(r)
		components.RegisterPage(err.Error(), csrfTokenErr).Render(r.Context(), w)
		return
	}

	if err := h.sm.SetUserSession(r, user.ID); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ShowLogin renders the login form.
func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	csrfToken := csrf.Token(r)
	components.LoginPage("", csrfToken).Render(r.Context(), w)
}

// Login handles POST /login. Verifies credentials via AuthService.Login
// (which also persists an opaque CLI token), resolves the user via
// Authenticate, stores the user ID in the scs session, then redirects.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	token, err := h.auth.Login(r.Context(), email, password)
	if err != nil {
		msg := "Invalid email or password"
		if !errors.Is(err, services.ErrInvalidCredentials) {
			msg = "An error occurred. Please try again."
		}
		csrfTokenErr := csrf.Token(r)
		components.LoginPage(msg, csrfTokenErr).Render(r.Context(), w)
		return
	}

	// Resolve user ID from the token we just created so we can store it in
	// the scs cookie session. The opaque token itself is not used by the web
	// layer — it exists for CLI auth — but Authenticate is the only port
	// method that maps token → User.
	user, err := h.auth.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	if err := h.sm.SetUserSession(r, user.ID); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles POST /logout. Destroys the scs session and redirects to /login.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sm.ClearSession(r); err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
