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

// MemoryURL returns the URL for a memory `/journal/{id}/memory/{memory_id}`
func MemoryURL(journalID int64, memoryID int64) string {
	return fmt.Sprintf("/journal/%d/memory/%d", journalID, memoryID)
}

func JournalJSONExportURL(journalID int64) string {
	return fmt.Sprintf("/journal/%d/export_json", journalID)
}
