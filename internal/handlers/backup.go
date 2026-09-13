package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shsiddhant/mone-pore/internal/db"
)

// JournalBackup represents the top-level exported journal data.
type JournalBackup struct {
	ID          int64          `json:"id"`
	JournalName string         `json:"journal_name"`
	Created     time.Time      `json:"created_at"`
	Modified    time.Time      `json:"modified_at"`
	Memories    []MemoryBackup `json:"memories"`
}

// MemoryBackup represents the denormalized exported memory data.
type MemoryBackup struct {
	ID         int64     `json:"id"`
	MemoryDate time.Time `json:"memory_date"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Tags       []string  `json:"tags"`
	Created    time.Time `json:"created_at"`
	Modified   time.Time `json:"modified_at"`
}

// ExportJournalToJSON serves a canonical JSON export download for a journal.
//
// Expected method : GET
func (app *Application) ExportJournalToJSON(w http.ResponseWriter, r *http.Request) {

	journalID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	dbJournal, err := app.DB.GetJournal(r.Context(), journalID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	dbMemories, err := app.DB.ListMemoryDetail(r.Context(), journalID, db.DESC)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		internalServerError(w)
		return
	}

	journalBackup := JournalBackup{
		ID:          dbJournal.ID,
		JournalName: dbJournal.JournalName,
		Created:     dbJournal.Created,
		Modified:    dbJournal.Modified,
		Memories:    []MemoryBackup{},
	}

	for _, memory := range dbMemories {
		tags := make([]string, 0, len(memory.Tags))
		for _, dbTag := range memory.Tags {
			tags = append(tags, dbTag.Title)
		}
		memoryBackup := MemoryBackup{
			ID:         memory.Memory.ID,
			MemoryDate: memory.Memory.MemoryDate,
			Title:      memory.Memory.Title,
			Body:       memory.Memory.Body,
			Tags:       tags,
			Created:    memory.Memory.Created,
			Modified:   memory.Memory.Modified,
		}

		journalBackup.Memories = append(journalBackup.Memories, memoryBackup)
	}

	fileName := fmt.Sprintf("%d-%s.json", dbJournal.ID, dbJournal.JournalName)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(
		"Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, fileName),
	)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(journalBackup); err != nil {
		http.Error(w, "Failed to generate export", http.StatusInternalServerError)
		return
	}
}
