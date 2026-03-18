package api

import (
	"encoding/json"
	"net/http"

	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	Email string `json:"email"`
}

// Register handles POST /api/v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{Email: user.Email})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /api/v1/auth/login. Returns an opaque Bearer token.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

type meResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Me handles GET /api/v1/auth/me.
// Requires Bearer auth. Returns the authenticated user's profile.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token, ok := middleware.TokenFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}

	user, err := h.auth.Authenticate(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}

// Logout handles POST /api/v1/auth/logout.
// Requires Authorization: Bearer <token>. Invalidates the token server-side.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token, ok := middleware.TokenFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}

	if err := h.auth.Logout(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
