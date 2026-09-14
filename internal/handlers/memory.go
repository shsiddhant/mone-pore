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
	// Normalize text area input
	body := strings.ReplaceAll(r.FormValue("body"), "\r\n", "\n")
	tagsString := r.FormValue("tags")

	tags := strings.Split(tagsString, ",")
	tags = normalizeTags(tags)

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

// EditMemoryPage renders the page for editing a memory in a journal.
//
// Expected method: GET
func (app *Application) EditMemoryPage(w http.ResponseWriter, r *http.Request) {
	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	memoryID, err := strconv.ParseInt(r.PathValue("memory_id"), 10, 64)
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

	currentMemory, err := app.DB.GetMemoryDetail(r.Context(), memoryID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	if err := templates.EditMemory(journal, currentMemory, "").
		Render(r.Context(), w); err != nil {
		internalServerError(w)
		return
	}
}

// EditMemory attempts to edit a memory.
// If successful, it redirects to the same memory's index page.
// If there's an error message, it renders the EditMemory Page again,
// with the same error message.
//
// Expected method: POST
func (app *Application) EditMemory(w http.ResponseWriter, r *http.Request) {

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
	memoryID, err := strconv.ParseInt(r.PathValue("memory_id"), 10, 64)
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
	// Normalize text area input
	body := strings.ReplaceAll(r.FormValue("body"), "\r\n", "\n")
	tagsString := r.FormValue("tags")

	tags := strings.Split(tagsString, ",")
	tags = normalizeTags(tags)

	if title == "" {
		errorMessage = "Title cannot be blank"
	}

	err = app.DB.UpdateMemory(
		r.Context(),
		journalID,
		memoryID,
		memoryDate,
		title,
		body,
		tags,
	)

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

	currentMemory, err := app.DB.GetMemoryDetail(r.Context(), memoryID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	if errorMessage != "" {
		if err := templates.EditMemory(journal, currentMemory, errorMessage).
			Render(r.Context(), w); err != nil {
			internalServerError(w)
			return
		}
		return
	}
	http.Redirect(w, r, route.MemoryURL(journalID, memoryID), http.StatusSeeOther)
}

// normalizeTags normalizes a slice of tag strings by removing blanks and duplicates.
func normalizeTags(tags []string) []string {
	// Create a set for seen tags
	seen := make(map[string]struct{})

	// Slice for normalized tags
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		//If tag is empty, skip
		if tag == "" {
			continue
		}
		// If tag is seen, skip (to prevent duplication)
		if _, ok := seen[tag]; ok {
			continue
		}

		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}

	return normalized
}
