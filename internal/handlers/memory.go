package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shsiddhant/mone-pore/internal/route"
	"github.com/shsiddhant/mone-pore/ui/templates"
)

// NewMemoryPage renders the page for creating a new memory in a journal.
//
// Expected method: GET
func (app *Application) NewMemoryPage(w http.ResponseWriter, r *http.Request) {

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

	if err := templates.NewMemory(journal, "").Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// NewMemory attempts to create a new memory in the database.
// If successful, it redirects to the parent journal index page.
// If there's an error message, it renders the NewMemoryPage again,
// with the same error message.
//
// Expected method: POST
func (app *Application) NewMemory(w http.ResponseWriter, r *http.Request) {

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

	errorMessage := ""

	// Go timestamp parsing reference: 2006-01-02T15:04:05Z07:00
	layout := "2006-01-02"

	memoryDate, err := time.Parse(layout, r.FormValue("memorydate"))

	if err != nil {
		internalServerError(w)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	body := r.FormValue("body")
	tagsString := r.FormValue("tags")

	tags := strings.Split(tagsString, ",")

	if title == "" {
		errorMessage = "Title cannot be blank"
	}

	_, err = app.DB.CreateMemory(r.Context(), journalID, memoryDate, title, body, tags)

	if err != nil {
		internalServerError(w)
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

	if errorMessage != "" {
		if err := templates.NewMemory(journal, errorMessage).
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		return
	}

	http.Redirect(w, r, route.JournalURL(journalID), http.StatusSeeOther)
}

// MemoryIndex renders the index page of a memory.
//
// Expected method: GET
func (app *Application) MemoryIndex(w http.ResponseWriter, r *http.Request) {

	memoryID, err := strconv.ParseInt(r.PathValue("memory_id"), 10, 64)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	memoryDetail, err := app.DB.GetMemoryDetail(r.Context(), memoryID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	if err := templates.MemoryIndex(memoryDetail).Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}

}
