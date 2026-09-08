// Package route provides URL builders for the application
package route

import "fmt"

// JournalURL returns the URL for a journal from its ID.
func JournalURL(journalID int64) string {
	return fmt.Sprintf("/journals/%d", journalID)
}
