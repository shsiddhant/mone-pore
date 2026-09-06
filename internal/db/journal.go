package db

import (
	"context"
	"time"
)

// Journal represents a journal stored in the database
type Journal struct {
	ID          int64
	JournalName string
	Password    string
	Created     time.Time
	Modified    time.Time
}

// CreateJournal creates a new journal with the given name and password.
// It returns the newly created journal or an error if the operation fails.
//
// Note: password is in general expected to be the hashed password instead.
func (db *DB) CreateJournal(
	ctx context.Context,
	journalname string,
	password string,
) (Journal, error) {
	var journal Journal

	queryString := `
	INSERT INTO journal (journalname, password)
	VALUES (?, ?)
	RETURNING id, journalname, password, created, modified
	`

	err := db.QueryRowContext(ctx, queryString, journalname, password).Scan(
		&journal.ID,
		&journal.JournalName,
		&journal.Password,
		&journal.Created,
		&journal.Modified,
	)
	return journal, err
}

// GetJournal retrieves a journal by its ID.
// It returns an error if the journal cannot be retrieved.
func (db *DB) GetJournal(
	ctx context.Context,
	journalID string,
) (Journal, error) {
	var journal Journal

	queryString := `
	SELECT * FROM journal
	WHERE id = ?`

	err := db.QueryRowContext(ctx, queryString, journalID).Scan(
		&journal.ID,
		&journal.JournalName,
		&journal.Password,
		&journal.Created,
		&journal.Modified,
	)
	return journal, err
}

// GetJournal retrieves a journal by its name.
// It returns an error if the journal cannot be retrieved.
func (db *DB) GetJournalByName(
	ctx context.Context,
	journalname string,
) (Journal, error) {
	var journal Journal

	queryString := `
	SELECT * FROM journal
	WHERE journalname = ?`

	err := db.QueryRowContext(ctx, queryString, journalname).Scan(
		&journal.ID,
		&journal.JournalName,
		&journal.Password,
		&journal.Created,
		&journal.Modified,
	)
	return journal, err
}

// UpdateJournalName updates the name of an existing journal.
// It returns an error if the update fails.
func (db *DB) UpdateJournalName(
	ctx context.Context,
	journalID int64,
	newName string,
) error {

	queryString := `
	UPDATE journal
	SET
		journalname = ?,
		modified = CURRENT_TIMESTAMP
	WHERE id = ?`

	_, err := db.ExecContext(ctx, queryString, newName, journalID)

	return err
}

// UpdateJournalPassword updates the password of an existing journal.
// It returns an error if the update fails.
//
// Note: newPassword is in general expected to be the hashed password instead.
func (db *DB) UpdateJournalPassword(
	ctx context.Context,
	journalID int64,
	newPassword string,
) error {

	queryString := `
	UPDATE journal
	SET
		password = ?,
		modified = CURRENT_TIMESTAMP
	WHERE id = ?`

	_, err := db.ExecContext(ctx, queryString, newPassword, journalID)

	return err
}
