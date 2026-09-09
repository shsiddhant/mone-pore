package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/route"
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

	if err != nil {
		internalServerError(w)
		return
	}

	if err := templates.Home(journals).Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// UnlockJournalForm renders the unlock journal form.
//
// Expected method: GET
func (app *Application) UnlockJournalForm(w http.ResponseWriter, r *http.Request) {

	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	key := fmt.Sprintf("journal_unlocked:%d", journalID)

	if app.SessionManager.GetBool(r.Context(), key) {
		w.Header().Set("HX-Redirect", route.JournalURL(journalID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	journal, err := app.DB.GetJournal(r.Context(), journalID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	journalSummary := db.JournalSummary{
		ID:          journal.ID,
		JournalName: journal.JournalName,
	}

	if err := templates.UnlockJournal(journalSummary, "").
		Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}

}

// UnlockJournal verifies the password hash and redirects to the journal page.
// If password hashes don't match, it renders the unlock journal form again,
// with an error message.
//
// Expected method: POST
func (app *Application) UnlockJournal(w http.ResponseWriter, r *http.Request) {

	// Ensure the request is a POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	journal, err := app.DB.GetJournal(r.Context(), journalID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	password := r.FormValue("password")

	if !checkPasswordHash(password, journal.Password) {
		journalSummary := db.JournalSummary{
			ID:          journal.ID,
			JournalName: journal.JournalName,
		}
		if err := templates.UnlockJournal(journalSummary, "Incorrect password").
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		// return to stop executing
		return
	}

	// Set session key
	key := fmt.Sprintf("journal_unlocked:%d", journalID)
	app.SessionManager.Put(r.Context(), key, true)

	// Redirect to the journal's index page.
	// Use HX-Redirect so HTMX performs a full-page navigation.
	// A normal HTTP redirect is followed by HTMX as part of the request,
	// causing the redirected page to be swapped into hx-target instead.
	w.Header().Set("HX-Redirect", route.JournalURL(journalID))
	w.WriteHeader(http.StatusNoContent)

}

// JournalIndex renders the index page of a journal.
//
// Expected method: GET
func (app *Application) JournalIndex(w http.ResponseWriter, r *http.Request) {

	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	journal, err := app.DB.GetJournal(r.Context(), journalID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	if err := templates.JournalIndex(journal).Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// checkPasswordHash compares password and password hash.
func checkPasswordHash(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
