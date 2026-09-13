package handlers

import (
	"net/http"
	"strings"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/route"
	"github.com/shsiddhant/mone-pore/ui/templates"
)

// SetupAdminPage renders first time admin settings page.
//
// Expected method: GET
func (app *Application) SetupAdminPage(w http.ResponseWriter, r *http.Request) {
	if err := templates.SetupAdmin("").Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// SetupAdmin page attemps to set admin settings.
// If successful, it redirects to home.
// If there's an error message, it renders the SetupAdminPage again,
// with the same error message.
//
// Expected method: POST
func (app *Application) SetupAdmin(w http.ResponseWriter, r *http.Request) {

	// Ensure the request is a POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")

	errorMessage := ""

	if strings.TrimSpace(password) == "" {
		errorMessage = "Password cannot be blank"
	}

	passwordHash, err := hashPassword(password)

	if err != nil {
		internalServerError(w)
		return
	}

	_, err = app.DB.CreateSetting(
		r.Context(),
		db.SettingAdminPasswordHash,
		passwordHash,
	)

	if err != nil {
		internalServerError(w)
		return
	}

	if errorMessage != "" {
		if err := templates.SetupAdmin(errorMessage).
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		return
	}
	http.Redirect(w, r, route.HomeURL(), http.StatusSeeOther)
}
