package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/shsiddhant/mone-pore/internal/db"
	"github.com/shsiddhant/mone-pore/internal/route"
	"github.com/shsiddhant/mone-pore/internal/session"
	"github.com/shsiddhant/mone-pore/ui/templates"
)

// UnlockJournalForm renders the unlock journal form.
//
// Expected method: GET
func (app *Application) UnlockJournalForm(w http.ResponseWriter, r *http.Request) {

	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	key := session.JournalUnlockedKey(journalID)

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
	key := session.JournalUnlockedKey(journalID)
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

	layout := "2006-01-02"

	queryParams := r.URL.Query()
	fromValue := queryParams.Get("from")
	toValue := queryParams.Get("to")

	var fromDate, toDate time.Time

	if fromValue != "" {
		fromDate, err = time.Parse(layout, fromValue)
		if err != nil {
			http.Error(
				w,
				"Invalid date format. Please use YYYY-MM-DD.",
				http.StatusBadRequest,
			)
			return
		}
	}

	if toValue != "" {
		toDate, err = time.Parse(layout, toValue)
		if err != nil {
			http.Error(
				w,
				"Invalid date format. Please use YYYY-MM-DD.",
				http.StatusBadRequest,
			)
			return
		}
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

	memories, err := app.DB.ListMemoryDetail(
		r.Context(),
		journalID,
		fromDate,
		toDate,
		db.DESC,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		internalServerError(w)
		return
	}

	if err := templates.JournalIndex(journal, memories).
		Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// LockJournal locks a journal for the session and redirects to home page.
//
// Expected method: POST
func (app *Application) LockJournal(w http.ResponseWriter, r *http.Request) {

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

	key := session.JournalUnlockedKey(journalID)

	// Remove key if present
	if app.SessionManager.GetBool(r.Context(), key) {
		app.SessionManager.Remove(r.Context(), key)
	}

	// Redirect to the home.
	// Use HX-Redirect so HTMX performs a full-page navigation.
	// A normal HTTP redirect is followed by HTMX as part of the request,
	// causing the redirected page to be swapped into hx-target instead.
	w.Header().Set("HX-Redirect", route.HomeURL())
	w.WriteHeader(http.StatusNoContent)

}

// NewJournalPage renders the page for creating a new journal.
//
// Expected method: GET
func (app *Application) NewJournalPage(w http.ResponseWriter, r *http.Request) {
	if err := templates.NewJournal("").Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// NewJournal attempts to create a new journal in the database.
// If successful, it redirects to home.
// If there's an error message, it renders the NewJournalPage again,
// with the same error message.
//
// Expected method: POST
func (app *Application) NewJournal(w http.ResponseWriter, r *http.Request) {

	// Ensure the request is a POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	journalName := r.FormValue("journalname")
	password := r.FormValue("password")

	errorMessage := ""

	if strings.TrimSpace(journalName) == "" {
		errorMessage = "Journal name cannot be blank"
	}

	if strings.TrimSpace(password) == "" {
		errorMessage = "Password cannot be blank"
	}

	passwordHash, err := hashPassword(password)

	if err != nil {
		internalServerError(w)
		return
	}

	_, err = app.DB.CreateJournal(r.Context(), journalName, passwordHash)

	err = db.CheckError(err)

	if errors.Is(err, db.ErrUniqueConstraint) {
		errorMessage = fmt.Sprintf(
			"A journal with name '%s' already exists.",
			journalName,
		)
	}

	if errorMessage != "" {
		if err := templates.NewJournal(errorMessage).
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		return
	}

	http.Redirect(w, r, route.HomeURL(), http.StatusSeeOther)

}

// UpdateJournalName updates a journal's name.
// If successful, it redirects to the journal's settings page.
// In case of a user facing error, it shows the error on the page.
//
// Expected method: POST
func (app *Application) UpdateJournalName(w http.ResponseWriter, r *http.Request) {
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

	newName := r.FormValue("new_name")

	if strings.TrimSpace(newName) == "" {
		if err := templates.JournalSettings(journal, "Journal name cannot be blank").
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		return
	}

	err = app.DB.UpdateJournalName(r.Context(), journal.ID, newName)

	err = db.CheckError(err)

	if errors.Is(err, db.ErrUniqueConstraint) {
		errorMessage := fmt.Sprintf(
			"Another journal with name '%s' already exists.",
			newName,
		)
		if err := templates.JournalSettings(journal, errorMessage).
			Render(r.Context(), w); err != nil {
			internalServerError(w)
		}
		return
	}
	if err != nil {
		internalServerError(w)
		return
	}

	http.Redirect(w, r, route.JournalSettingsURL(journal.ID), http.StatusSeeOther)

}

// UpdateJournalPassword updates a journal's password.
// If successful, it redirects to the journal's settings page.
// In case of a user facing error, it shows the error on the page.
//
// Expected method: POST
func (app *Application) UpdateJournalPassword(w http.ResponseWriter, r *http.Request) {
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

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmNewPassword := r.FormValue("confirm_new_password")

	if strings.TrimSpace(newPassword) == "" {
		if err := templates.JournalSettings(journal, "Password cannot be blank").
			Render(r.Context(), w); err != nil {
			internalServerError(w)
		}
		return
	}

	if newPassword != confirmNewPassword {
		if err := templates.JournalSettings(journal, "Passwords don't match").
			Render(r.Context(), w); err != nil {
			internalServerError(w)
		}
		return
	}

	if !checkPasswordHash(currentPassword, journal.Password) {
		if err := templates.JournalSettings(journal, "Current password is wrong").
			Render(r.Context(), w); err != nil {
			internalServerError(w)
		}
		return
	}

	newPasswordHash, err := hashPassword(newPassword)
	if err != nil {
		internalServerError(w)
		return
	}

	err = app.DB.UpdateJournalPassword(r.Context(), journal.ID, newPasswordHash)
	err = db.CheckError(err)

	if err != nil {
		internalServerError(w)
		return
	}

	http.Redirect(w, r, route.JournalSettingsURL(journal.ID), http.StatusSeeOther)

}

// JournalSettings renders the journal settings page.
//
// Expected method: GET
func (app *Application) JournalSettings(w http.ResponseWriter, r *http.Request) {

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

	if err := templates.JournalSettings(journal, "").
		Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// checkPasswordHash compares password and password hash.
func checkPasswordHash(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}

// hashPassword uses bcrypt to hash a password.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
