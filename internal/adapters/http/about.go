package httphandlers

import (
	"net/http"

	"github.com/justestif/go-jobs/components"
)

func About(w http.ResponseWriter, r *http.Request) {
	components.About().Render(r.Context(), w)
}
