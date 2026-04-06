package httphandlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/justestif/go-jobs/components"
	"github.com/justestif/go-jobs/internal/adapters/http/middleware"
	"github.com/justestif/go-jobs/internal/core/ports"
)

// ContactsHandler manages LinkedIn contact import and display.
type ContactsHandler struct {
	contacts ports.ContactService
}

// NewContactsHandler constructs a ContactsHandler.
func NewContactsHandler(contacts ports.ContactService) *ContactsHandler {
	return &ContactsHandler{contacts: contacts}
}

// Show renders the contacts management page (GET /contacts).
func (h *ContactsHandler) Show(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	csrfToken := csrf.Token(r)

	total, linked, companies, err := h.contacts.Stats(r.Context(), userID)
	if err != nil {
		log.Printf("contacts stats: %v", err)
	}

	flash := r.URL.Query().Get("flash")
	components.ContactsPage(total, linked, companies, flash, true, csrfToken).Render(r.Context(), w)
}

// Import handles CSV file upload (POST /contacts/import).
func (h *ContactsHandler) Import(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := h.contacts.ImportCSV(r.Context(), userID, file)
	if err != nil {
		log.Printf("contact import error: %v", err)
		http.Redirect(w, r, "/contacts?flash=Import+failed:+"+err.Error(), http.StatusSeeOther)
		return
	}

	flash := fmt.Sprintf("Imported %d contacts. %d companies linked, %d new companies registered, %d unmatched.",
		result.ContactsImported, result.CompaniesLinked, result.CompaniesRegistered, result.CompaniesUnmatched)
	http.Redirect(w, r, "/contacts?flash="+flash, http.StatusSeeOther)
}

// Delete removes all contacts for the user (POST /contacts/delete).
func (h *ContactsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	if err := h.contacts.DeleteContacts(r.Context(), userID); err != nil {
		log.Printf("contact delete error: %v", err)
		http.Redirect(w, r, "/contacts?flash=Delete+failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/contacts?flash=All+contacts+deleted.", http.StatusSeeOther)
}
