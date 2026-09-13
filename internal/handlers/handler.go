package handlers

import (
	"net/http"

	"github.com/alexedwards/scs/v2"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/session"
	"github.com/shsiddhant/mone-pore/ui/templates"
)

type Application struct {
	DB             *db.DB
	SessionManager *scs.SessionManager
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {

	journals, err := app.DB.ListJournals(r.Context())

	unlockedMap := make(map[int64]bool)

	for _, journal := range journals {
		unlockedMap[journal.ID] = session.IsJournalUnlocked(
			r.Context(),
			app.SessionManager,
			journal.ID,
		)
	}

	if err != nil {
		internalServerError(w)
		return
	}

	if err := templates.Home(journals, unlockedMap).Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}
