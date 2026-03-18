package httphandlers

import (
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
)

func ContactForm(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := middleware.UserIDFromContext(r.Context())
	csrfToken := csrf.Token(r)
	components.Contact(loggedIn, csrfToken).Render(r.Context(), w)
}
