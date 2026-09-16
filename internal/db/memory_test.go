package db

import (
	"context"
	"database/sql"
	"errors"
	"maps"
	"testing"
	"time"
)

func TestCreateTag(t *testing.T) {
	database := testDB(t)

	tagTitle := "Go"

	tag, err := database.CreateTag(
		context.Background(),
		tagTitle,
	)

	if err != nil {
		t.Fatal(err)
	}

	if tag.Title != tagTitle {
		t.Errorf("got tag title %q, want %q", tag.Title, tagTitle)
	}

}

func TestGetTag(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	expected, err := database.CreateTag(ctx, "Go")
	if err != nil {
		t.Fatal(err)
	}

	got, err := database.GetTag(ctx, expected.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Errorf("got %+v, want %+v", got, expected)
	}
}

func TestGetTagNotFound(t *testing.T) {
	database := testDB(t)

	_, err := database.GetTag(context.Background(), 123)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}
}

func TestGetTagByTitleNotFound(t *testing.T) {
	database := testDB(t)

	_, err := database.GetTagByTitle(context.Background(), "Go")

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}
}

func TestCreateTagDuplicate(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	_, err := database.CreateTag(ctx, "Go")
	if err != nil {
		t.Fatal(err)
	}

	_, err = database.CreateTag(ctx, "Go")
	if err == nil {
		t.Fatal("expected duplicate tag creation to fail")
	}
}

func TestGetTagByTitle(t *testing.T) {
	database := testDB(t)

	tagTitle := "Go"
	tagExpected, err := database.CreateTag(
		context.Background(),
		tagTitle,
	)

	if err != nil {
		t.Fatal(err)
	}

	tag, err := database.GetTagByTitle(
		context.Background(),
		tagTitle,
	)

	if err != nil {
		t.Fatal(err)
	}

	if tagExpected != tag {
		t.Errorf("got %+v, want %+v", tagExpected, tag)
	}

}

func TestCreateMemory(t *testing.T) {
	database := testDB(t)

	journal := seedJournal(t, database)

	istOffset := int((5*time.Hour + 30*time.Minute) / time.Second)
	loc := time.FixedZone("IST", istOffset)

	memoryDate := time.Date(2023, 12, 17, 0, 0, 0, 0, loc)

	title := "The Best Day"
	body := "It was the happiest day of my life."

	memory, err := database.CreateMemory(
		context.Background(),
		journal.ID,
		memoryDate,
		title,
		body,
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	if memory.ID == 0 {
		t.Error("expected memory ID to be set")
	}

	if memory.JournalID != journal.ID {
		t.Errorf("got journal ID %d, want %d", memory.JournalID, journal.ID)
	}

	if !memory.MemoryDate.Equal(memoryDate) {
		t.Errorf("got memory date %v, want %v", memory.MemoryDate, memoryDate)
	}

	if memory.Title != title {
		t.Errorf("got title %q, want %q", memory.Title, title)
	}

	if memory.Body != body {
		t.Errorf("got body %q, want %q", memory.Body, body)
	}

	if memory.Created.IsZero() {
		t.Error("expected Created to be set")
	}

	if memory.Modified.IsZero() {
		t.Error("expected Modified to be set")
	}

	var count int

	err = database.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM memory_tag WHERE memory_id = ?`,
		memory.ID,
	).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("got %d memory tags, want 0", count)
	}

}

func TestCreateMemoryWithExistingTag(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)
	tag, err := database.CreateTag(ctx, "Go")
	if err != nil {
		t.Fatal(err)
	}

	memory, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go"},
	)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the relationship uses the existing tag.
	var tagID int64

	err = database.QueryRowContext(
		ctx,
		`
		SELECT tag_id
		FROM memory_tag
		WHERE memory_id = ?
		`,
		memory.ID,
	).Scan(&tagID)

	if err != nil {
		t.Fatal(err)
	}

	if tagID != tag.ID {
		t.Errorf("got tag ID %d, want %d", tagID, tag.ID)
	}
}

func TestCreateMemoryWithNewTag(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	memory, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go"},
	)
	if err != nil {
		t.Fatal(err)
	}

	tag, err := database.GetTagByTitle(ctx, "Go")

	if err != nil {
		t.Fatal(err)
	}

	if tag.ID == 0 {
		t.Error("expected tag ID to be set")
	}

	var tagID int64

	err = database.QueryRowContext(
		ctx,
		`
		SELECT tag_id
		FROM memory_tag
		WHERE memory_id = ?
		`,
		memory.ID,
	).Scan(&tagID)

	if err != nil {
		t.Fatal(err)
	}

	if tagID != tag.ID {
		t.Errorf("got tag ID %d, want %d", tagID, tag.ID)
	}
}

func TestCreateMemoryWithMultipleTags(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	memory, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go", "SQLite", "HTMX"},
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, title := range []string{"Go", "SQLite", "HTMX"} {
		tag, err := database.GetTagByTitle(ctx, title)
		if err != nil {
			t.Fatal(err)
		}

		var count int

		err = database.QueryRowContext(
			ctx,
			`
			SELECT COUNT(*)
			FROM memory_tag
			WHERE memory_id = ? AND tag_id = ?
			`,
			memory.ID,
			tag.ID,
		).Scan(&count)

		if err != nil {
			t.Fatal(err)
		}

		if count != 1 {
			t.Errorf(
				"got %d relationship(s) for tag %q, want 1",
				count,
				title,
			)
		}
	}
}

func TestCreateMemoryWithDuplicateTags(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	memory, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go", "Go", "SQLite", "Go"},
	)
	if err != nil {
		t.Fatal(err)
	}

	var count int

	err = database.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM memory_tag
		WHERE memory_id = ?
		`,
		memory.ID,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("got %d memory tags, want 2", count)
	}

	err = database.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM tag
		WHERE title = 'Go'
		`,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("got %d Go tags, want 1", count)
	}
}

func TestCreateMemoryRollback(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	_, err := database.ExecContext(ctx, `
		CREATE TRIGGER fail_memory_tag_insert
		BEFORE INSERT ON memory_tag
		BEGIN
			SELECT RAISE(ABORT, 'intentional test failure');
		END
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go"},
	)
	if err == nil {
		t.Fatal("expected CreateMemory to fail")
	}

	var count int

	err = database.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM memory`,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("got %d memories, want 0", count)
	}

	err = database.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM tag WHERE title = 'Go'`,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("got %d Go tags, want 0", count)
	}
}

func TestGetMemory(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	expected, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go"},
	)

	if err != nil {
		t.Fatal(err)
	}

	got, err := database.GetMemory(ctx, expected.JournalID, expected.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got != expected {
		t.Errorf("got %+v, want %+v", got, expected)
	}
}

func TestDeleteMemory(t *testing.T) {
	database := testDB(t)
	ctx := context.Background()

	journal := seedJournal(t, database)

	memory, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Test memory",
		"Test body",
		[]string{"Go"},
	)

	if err != nil {
		t.Fatal(err)
	}

	tag, err := database.GetTagByTitle(ctx, "Go")
	if err != nil {
		t.Fatal(err)
	}

	if err := database.DeleteMemory(ctx, memory.ID); err != nil {
		t.Fatal(err)
	}

	_, err = database.GetMemory(ctx, memory.JournalID, memory.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}

	var count int

	err = database.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM memory_tag
		WHERE memory_id = ?
		`,
		memory.ID,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("got %d memory_tag rows, want 0", count)
	}

	_, err = database.GetTag(ctx, tag.ID)
	if err != nil {
		t.Errorf("tag was deleted with memory: %v", err)
	}
}

func TestGetMemoryNotFound(t *testing.T) {
	database := testDB(t)

	_, err := database.GetMemory(context.Background(), 1, 123)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}
}

func TestDeleteMemoryDoesNotDeleteOtherMemories(t *testing.T) {
	database := testDB(t)

	ctx := context.Background()

	journal := seedJournal(t, database)

	first, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"First memory",
		"First body",
		[]string{"Go"},
	)
	if err != nil {
		t.Fatal(err)
	}

	second, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Second memory",
		"Second body",
		[]string{"Go"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := database.DeleteMemory(ctx, first.ID); err != nil {
		t.Fatal(err)
	}

	got, err := database.GetMemory(ctx, second.JournalID, second.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != second.ID {
		t.Errorf("got memory ID %d, want %d", got.ID, second.ID)
	}
}

func TestGetTagsByMemoryID(t *testing.T) {
	database := testDB(t)

	ctx := context.Background()

	journal := seedJournal(t, database)

	first, err := database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"First memory",
		"First body",
		[]string{"first", "common"},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = database.CreateMemory(
		ctx,
		journal.ID,
		time.Now(),
		"Second memory",
		"Second body",
		[]string{"second", "common"},
	)
	if err != nil {
		t.Fatal(err)
	}

	got, err := database.GetTagsByMemoryID(ctx, first.ID)

	if err != nil {
		t.Fatal(err)
	}

	gotSet := make(map[string]struct{})
	expectedSet := map[string]struct{}{
		"first":  {},
		"common": {},
	}

	for _, tag := range got {
		gotSet[tag.Title] = struct{}{}
	}

	if !maps.Equal(gotSet, expectedSet) {
		t.Errorf("got %s, want %s", gotSet, expectedSet)
	}
}

func TestUpdateMemory(t *testing.T) {

	ctx := context.Background()
	database := testDB(t)

	oldMemory := seedMemory(t, database)

	newTitle := "The Best Day of My Life"
	newBody := oldMemory.Body + "\n" + "We went to watch Wonka."
	newTags := []string{"bestday", "moviedate", "firstdate"}

	err := database.UpdateMemory(
		ctx,
		oldMemory.JournalID,
		oldMemory.ID,
		oldMemory.MemoryDate,
		newTitle,
		newBody,
		newTags,
	)
	if err != nil {
		t.Fatalf("UpdateMemory() error = %v", err)
	}

	newMemory, err := database.GetMemory(ctx, oldMemory.JournalID, oldMemory.ID)

	if err != nil {
		t.Fatal(err)
	}

	if newMemory.ID != oldMemory.ID {
		t.Errorf("got memory ID %d, want %d", newMemory.ID, oldMemory.ID)
	}
	if newMemory.JournalID != oldMemory.JournalID {
		t.Errorf("got memory ID %d, want %d", newMemory.JournalID, oldMemory.JournalID)
	}
	if !newMemory.MemoryDate.Equal(oldMemory.MemoryDate) {
		t.Errorf(
			"got memory date %v, want %v",
			newMemory.MemoryDate,
			oldMemory.MemoryDate,
		)
	}

	if newMemory.Title != newTitle {
		t.Errorf("got title %q, want %q", newMemory.Title, newTitle)
	}

	if newMemory.Body != newBody {
		t.Errorf("got body %q, want %q", newMemory.Body, newBody)
	}

	for _, title := range newTags {
		tag, err := database.GetTagByTitle(ctx, title)
		if err != nil {
			t.Fatal(err)
		}

		var count int

		err = database.QueryRowContext(
			ctx,
			`
			SELECT COUNT(*)
			FROM memory_tag
			WHERE memory_id = ? AND tag_id = ?
			`,
			newMemory.ID,
			tag.ID,
		).Scan(&count)

		if err != nil {
			t.Fatal(err)
		}

		if count != 1 {
			t.Errorf(
				"got %d relationships for tag %q, want 1",
				count,
				title,
			)
		}
	}

	var totalCount int
	err = database.QueryRowContext(
		ctx,
		`
    SELECT COUNT(*)
    FROM memory_tag
    WHERE memory_id = ?
    `,
		newMemory.ID,
	).Scan(&totalCount)

	if err != nil {
		t.Fatal(err)
	}
	if totalCount != len(newTags) {
		t.Errorf(
			"got %d total relationships, want %d",
			totalCount,
			len(newTags),
		)
	}

}
