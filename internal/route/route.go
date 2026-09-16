// Package route provides URL builders for the application
package route

import "fmt"

// HomeURL returns the URL for home page.
func HomeURL() string {
	return "/"
}

// SetupAdminURL returns the URL /admin/setup
func SetupAdminURL() string {
	return "/admin/setup"
}

// JournalURL returns the URL for a journal from its ID.
func JournalURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d", journalID)
}

// UnlockJournalURL returns the URL `/journal/{id}/unlock`
func UnlockJournalURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d/unlock", journalID)
}

// LockJournalURL returns the URL `/journal/{id}/lock`
func LockJournalURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d/lock", journalID)
}

// NewJournalURL returns the URL /journal/new
func NewJournalURL() string {
	return "/journal/new"
}

// NewMemoryURL returns the URL /memory/new
func NewMemoryURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d/memory/new", journalID)
}

// MemoryURL returns the URL for a memory:
//
//	`/journal/{id}/memory/{memory_id}`
func MemoryURL(journalID int64, memoryID int64) string {
	return fmt.Sprintf("/journal/%d/memory/%d", journalID, memoryID)
}

// EditMemoryURL returns the URL for editing a memory:
//
//	`/journal/{id}/memory/{memory_id}/edit`
func EditMemoryURL(journalID int64, memoryID int64) string {
	return fmt.Sprintf("/journal/%d/memory/%d/edit", journalID, memoryID)
}

// JournalJSONExportURL returns the URL for exporting a journal to JSON:
//
//	`journal/{id}/export_json`
func JournalJSONExportURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d/export_json", journalID)
}

// DeleteMemoryURL returns the URL for deleting a memory:
//
//	`journal/{id}/memory/{memory_id}/delete`
func DeleteMemoryURL(journalID int64, memoryID int64) string {
	return fmt.Sprintf("/journal/%d/memory/%d/delete", journalID, memoryID)
}
