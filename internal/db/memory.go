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

// MemoryDetail represents a denormalized memory.
type MemoryDetail struct {
	Memory Memory
	Tags   []Tag
}

type Ordering string

const (
	ASC  Ordering = "ASC"
	DESC Ordering = "DESC"
	NONE Ordering = "NONE"
)

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

// GetTagsByMemoryID retrieves all tags for a memory.
// It returns an error if tags cannot be retrieved.
func (db *DB) GetTagsByMemoryID(
	ctx context.Context,
	memoryID int64,
) ([]Tag, error) {

	queryString := `
	SELECT t.id, t.title
	FROM tag t
	JOIN memory_tag mt ON t.id = mt.tag_id
	WHERE mt.memory_id = ?
	`

	rows, err := db.QueryContext(ctx, queryString, memoryID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []Tag{}

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Title); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tags, err

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
	for _, title := range tags {
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
			ON CONFLICT DO NOTHING
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

// GetMemory retrieves a memory by its ID and journalID.
// It returns an error if the memory cannot be retrieved.
func (db *DB) GetMemory(
	ctx context.Context,
	journalID int64,
	memoryID int64,
) (Memory, error) {
	var memory Memory

	queryString := `
	SELECT * FROM memory
	WHERE
		id = ? AND
		journal_id = ?
	`

	err := db.QueryRowContext(ctx, queryString, memoryID, journalID).Scan(
		&memory.ID,
		&memory.JournalID,
		&memory.MemoryDate,
		&memory.Title,
		&memory.Body,
		&memory.Created,
		&memory.Modified,
	)
	return memory, err
}

// GetMemoryDetail returns MemoryDetail from memory ID and journalID.
// It returns an error if the memory cannot be retrieved.
func (db *DB) GetMemoryDetail(
	ctx context.Context,
	journalID int64,
	memoryID int64,
) (MemoryDetail, error) {

	var memoryDetail MemoryDetail

	memory, err := db.GetMemory(ctx, journalID, memoryID)

	if err != nil {
		return memoryDetail, err
	}

	tags, err := db.GetTagsByMemoryID(ctx, memory.ID)

	if err != nil {
		return memoryDetail, err
	}

	memoryDetail.Memory = memory
	memoryDetail.Tags = tags

	return memoryDetail, nil
}

// ListMemoryDetail retrieves all memories in the database along with their
// respective tags.
func (db *DB) ListMemoryDetail(
	ctx context.Context,
	journalID int64,
	fromDate time.Time,
	toDate time.Time,
	dateOrdering Ordering,
) ([]MemoryDetail, error) {

	var memories []MemoryDetail

	memoryQuery := `
	SELECT id, journal_id, title, body, memorydate, created, modified
	FROM memory
	WHERE journal_id = ?
	`

	queryParams := []any{journalID}

	if !fromDate.IsZero() {
		memoryQuery += " AND memorydate >= ?"
		queryParams = append(queryParams, fromDate)
	}

	if !toDate.IsZero() {
		memoryQuery += " AND memorydate <= ?"
		queryParams = append(queryParams, toDate)
	}

	switch dateOrdering {
	case "ASC":
		memoryQuery += " ORDER BY memorydate ASC, created ASC"
	case "DESC":
		memoryQuery += " ORDER BY memorydate DESC, created DESC"
	}

	memRows, err := db.QueryContext(ctx, memoryQuery, queryParams...)

	if err != nil {
		return nil, err
	}
	defer memRows.Close()

	memMap := make(map[int64]*MemoryDetail)

	var orderedIDs []int64

	for memRows.Next() {
		var md MemoryDetail
		err := memRows.Scan(
			&md.Memory.ID,
			&md.Memory.JournalID,
			&md.Memory.Title,
			&md.Memory.Body,
			&md.Memory.MemoryDate,
			&md.Memory.Created,
			&md.Memory.Modified,
		)
		if err != nil {
			return nil, err
		}

		md.Tags = []Tag{}

		memories = append(memories, md)
		orderedIDs = append(orderedIDs, md.Memory.ID)

		memMap[md.Memory.ID] = &memories[len(memories)-1]
	}
	if err = memRows.Err(); err != nil {
		return nil, err
	}

	if len(memories) == 0 {
		return memories, nil
	}

	tagQuery := `
	SELECT mt.memory_id, t.id, t.title
	FROM tag t
	JOIN memory_tag mt ON t.id = mt.tag_id
	JOIN memory m ON m.id = mt.memory_id
	WHERE m.journal_id = ?
	`
	tagRows, err := db.QueryContext(ctx, tagQuery, journalID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var memoryID int64
		var tag Tag
		if err := tagRows.Scan(&memoryID, &tag.ID, &tag.Title); err != nil {
			return nil, err
		}
		if targetDetail, exists := memMap[memoryID]; exists {
			targetDetail.Tags = append(targetDetail.Tags, tag)
		}
	}

	if err = tagRows.Err(); err != nil {
		return nil, err
	}
	return memories, nil
}

// DeleteMemory deletes a memory from the database.
// It returns an error if the operation fails.
func (db *DB) DeleteMemory(
	ctx context.Context,
	memoryID int64,
) error {

	queryString := `
	DELETE FROM memory
	WHERE id = ?
	`

	_, err := db.ExecContext(ctx, queryString, memoryID)

	return err
}

// ListAllMemoryDetail retrieves all memories in the database along with their
// respective tags.
func (db *DB) ListAllMemoryDetail(
	ctx context.Context,
	dateOrdering Ordering,
) ([]MemoryDetail, error) {
	var memories []MemoryDetail

	memoryQuery := `
	SELECT id, journal_id, title, body, memorydate, created, modified
	FROM memory
	`

	switch dateOrdering {
	case "ASC":
		memoryQuery += " ORDER BY memorydate ASC, created ASC"
	case "DESC":
		memoryQuery += " ORDER BY memorydate DESC, created DESC"
	}

	memRows, err := db.QueryContext(ctx, memoryQuery)
	if err != nil {
		return nil, err
	}
	defer memRows.Close()

	memMap := make(map[int64]*MemoryDetail)

	for memRows.Next() {
		var md MemoryDetail

		if err := memRows.Scan(
			&md.Memory.ID,
			&md.Memory.JournalID,
			&md.Memory.Title,
			&md.Memory.Body,
			&md.Memory.MemoryDate,
			&md.Memory.Created,
			&md.Memory.Modified,
		); err != nil {
			return nil, err
		}

		md.Tags = []Tag{}

		memories = append(memories, md)
		memMap[md.Memory.ID] = &memories[len(memories)-1]
	}

	if err := memRows.Err(); err != nil {
		return nil, err
	}

	if len(memories) == 0 {
		return memories, nil
	}

	tagQuery := `
	SELECT mt.memory_id, t.id, t.title
	FROM tag t
	JOIN memory_tag mt ON t.id = mt.tag_id
	JOIN memory m ON m.id = mt.memory_id
	`

	tagRows, err := db.QueryContext(ctx, tagQuery)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var memoryID int64
		var tag Tag

		if err := tagRows.Scan(
			&memoryID,
			&tag.ID,
			&tag.Title,
		); err != nil {
			return nil, err
		}

		if targetDetail, exists := memMap[memoryID]; exists {
			targetDetail.Tags = append(targetDetail.Tags, tag)
		}
	}

	if err := tagRows.Err(); err != nil {
		return nil, err
	}

	return memories, nil
}

// UpdateMemory updates a memory in the database.
// It returns an error if the update fails.
func (db *DB) UpdateMemory(
	ctx context.Context,
	journalID,
	memoryID int64,
	newMemoryDate time.Time,
	newTitle string,
	newBody string,
	newTags []string,
) error {
	// Get a Tx for making transaction requests.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	queryString := `
	UPDATE memory
	SET
		memorydate = ?,
		title = ?,
		body = ?,
		modified = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err = tx.ExecContext(
		ctx,
		queryString,
		newMemoryDate,
		newTitle,
		newBody,
		memoryID,
	)
	if err != nil {
		return err
	}

	// Delete tags from junction table 'memory_tag'
	_, err = tx.ExecContext(
		ctx,
		`
		DELETE FROM memory_tag
		WHERE memory_id = ?
		`,
		memoryID)
	if err != nil {
		return err
	}

	for _, title := range newTags {
		tag, err := getOrCreateTag(ctx, tx, title)
		if err != nil {
			return err
		}

		// Update junction table 'memory_tag'
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO memory_tag (journal_id, memory_id, tag_id)
			VALUES (?, ?, ?)
			ON CONFLICT DO NOTHING
			`,
			journalID,
			memoryID,
			tag.ID,
		)
		if err != nil {
			return err
		}
	}
	// Commit the transaction.
	return tx.Commit()
}

// ------------------
// Private Helpers //
// -----------------

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
