package session

import (
	"context"
	"fmt"

	"github.com/alexedwards/scs/v2"
)

// JournalUnlockKey returns a key name from journal ID.
func JournalUnlockedKey(journalID int64) string {
	return fmt.Sprintf("journal_unlocked:%d", journalID)
}

// IsJournalUnlocked checks whether or not a journal is unlocked in the current session.
func IsJournalUnlocked(
	ctx context.Context,
	sessionManager *scs.SessionManager,
	journalID int64,
) bool {
	key := JournalUnlockedKey(journalID)

	return sessionManager.GetBool(ctx, key)
}
