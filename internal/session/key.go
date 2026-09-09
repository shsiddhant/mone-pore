package session

import "fmt"

// JournalUnlockKey returns a key name from journal ID.
func JournalUnlockedKey(journalID int64) string {
	return fmt.Sprintf("journal_unlocked:%d", journalID)
}
