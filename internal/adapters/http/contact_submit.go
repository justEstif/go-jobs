package httphandlers

import (
	"log"
	"net/http"

	"github.com/justestif/go-jobs/components"
)

func ContactSubmit(w http.ResponseWriter, r *http.Request) {
	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")

	// Log the submission (in production, save to DB or send email)
	log.Printf("Contact form submission - Name: %s, Email: %s, Message: %s", name, email, message)

	// Return HTMX response
	components.ContactSuccess().Render(r.Context(), w)
}
