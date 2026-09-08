package handlers

import (
	"net/http"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/ui/templates"
)

type Application struct {
	DB *db.DB
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {

	journals, err := app.DB.ListJournals(r.Context())

	if err != nil {
		internalServerError(w)
	}

	if err := templates.Home(journals).Render(r.Context(), w); err != nil {
		internalServerError(w)
	}
}
