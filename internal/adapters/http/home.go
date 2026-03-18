package httphandlers

import (
	"net/http"

	"github.com/justestif/go-jobs/components"
)

func Home(w http.ResponseWriter, r *http.Request) {
	components.Home().Render(r.Context(), w)
}
