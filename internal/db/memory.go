package db

import (
	"context"
	"database/sql"
	"time"
)

// Memory represents a memory, i.e. a journal entry, stored in the database.
type Memory struct {
	ID         int64
	JournalID  int64
	MemoryDate time.Time
	Title      string
	Body       string
	Created    time.Time
	Modified   time.Time
}

// Tag represents a tag stored in the database.
type Tag struct {
	ID    int64
	Title string
}

// CreateTag creates a new tag with the given title.
// It returns the newly created tag or an error if the operation fails.
func (db *DB) CreateTag(
	ctx context.Context,
	title string,
) (Tag, error) {
	var tag Tag

	queryString := `
	INSERT INTO tag (title)
	VALUES (?)
	RETURNING id, title
	`

	err := db.QueryRowContext(ctx, queryString, title).Scan(
		&tag.ID,
		&tag.Title,
	)
	return tag, err
}

// GetTag retrieves a tag by its ID.
// It returns an error if the tag cannot be retrieved.
func (db *DB) GetTag(
	ctx context.Context,
	tagID int64,
) (Tag, error) {
	var tag Tag

	queryString := `
	SELECT * FROM tag
	WHERE id = ?
	`

	err := db.QueryRowContext(ctx, queryString, tagID).Scan(
		&tag.ID,
		&tag.Title,
	)
	return tag, err
}

// GetTagByTitle retrieves a tag by its title.
// It returns an error if the tag cannot be retrieved.
func (db *DB) GetTagByTitle(
	ctx context.Context,
	title string,
) (Tag, error) {
	var tag Tag

	queryString := `
	SELECT * FROM tag
	WHERE title = ?
	`

	err := db.QueryRowContext(ctx, queryString, title).Scan(
		&tag.ID,
		&tag.Title,
	)
	return tag, err
}

// CreateMemory creates a new memory in the database.
// It returns the newly created memory or an error if the operation fails.
func (db *DB) CreateMemory(
	ctx context.Context,
	journalID int64,
	memoryDate time.Time,
	title string,
	body string,
	tags []string,
) (Memory, error) {

	// Create a helper function for preparing failure results.
	fail := func(err error) (Memory, error) {
		return Memory{}, err
	}

	// Get a Tx for making transaction requests.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	var memory Memory

	queryString := `
	INSERT INTO memory (
		journal_id,
		memorydate,
		title,
		body
	)
	VALUES (?, ?, ?, ?)
	RETURNING id, journal_id, memorydate, title, body, created, modified
	`

	err = tx.QueryRowContext(ctx, queryString, journalID, memoryDate, title, body).Scan(
		&memory.ID,
		&memory.JournalID,
		&memory.MemoryDate,
		&memory.Title,
		&memory.Body,
		&memory.Created,
		&memory.Modified,
	)
	if err != nil {
		return fail(err)
	}

	// Add tags to the database
	for _, title := range normalizeTags(tags) {
		// getOrCreateTag
		tag, err := getOrCreateTag(ctx, tx, title)

		if err != nil {
			return fail(err)
		}
		// Update the junction table 'memory_tag'
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO memory_tag (journal_id, memory_id, tag_id)
			VALUES (?, ?, ?)
			`,
			memory.JournalID,
			memory.ID,
			tag.ID,
		)

		if err != nil {
			return fail(err)
		}
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return memory, nil
}

// getOrCreateTag creates a new tag in the database. If a tag with the same
// title already exists in the database, the it returns that tag.
//
// It expects a transaction, so that the operation may be used inside
// a transaction.
func getOrCreateTag(
	ctx context.Context,
	tx *sql.Tx,
	title string,
) (Tag, error) {

	// Create a helper function for preparing failure results.
	fail := func(err error) (Tag, error) {
		return Tag{}, err
	}

	// Try to insert the tag. On conflict do nothing.
	_, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO tag (title)
        VALUES (?)
        ON CONFLICT(title) DO NOTHING
		`,
		title,
	)

	if err != nil {
		return fail(err)
	}

	var tag Tag

	err = tx.QueryRowContext(
		ctx,
		`
        SELECT id, title
        FROM tag
        WHERE title = ?
		`,
		title,
	).Scan(
		&tag.ID,
		&tag.Title,
	)

	if err != nil {
		return fail(err)
	}

	return tag, nil
}

// normalizeTags normalizes a slice of tag strings by removing blanks and duplicates.
func normalizeTags(tags []string) []string {
	// Create a set for seen tags
	seen := make(map[string]struct{})

	// Slice for normalized tags
	normalized := make([]string, 0, len(tags))

	for _, tag := range tags {
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
