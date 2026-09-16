package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestCreateJournal(t *testing.T) {
	expectedID := int64(1)
	journalName := "Reconstruction"
	passwordHash := "password-hash"
	database := testDB(t)

	journal, err := database.CreateJournal(
		context.Background(),
		journalName,
		passwordHash,
	)

	if err != nil {
		t.Fatal(err)
	}

	if journal.ID != expectedID {
		t.Errorf("got ID = %d, want %d", journal.ID, expectedID)
	}

	if journal.JournalName != journalName {
		t.Errorf("got journalname = %q, want %q", journal.JournalName, journalName)
	}

	if journal.Password != passwordHash {
		t.Errorf("got password = %q, want %q", journal.Password, passwordHash)
	}

	if journal.Created.IsZero() {
		t.Error("Created is zero")
	}

	if journal.Modified.IsZero() {
		t.Error("Modified is zero")
	}

}

func TestGetJournal(t *testing.T) {
	database := testDB(t)

	journalExpected := seedJournal(t, database)

	journal, err := database.GetJournal(context.Background(), journalExpected.ID)

	if err != nil {
		t.Fatal(err)
	}

	if journalExpected != journal {
		t.Errorf("got journal %+v, want %+v", journal, journalExpected)
	}
}

func TestGetJournalByName(t *testing.T) {
	database := testDB(t)

	journalExpected := seedJournal(t, database)
	journalName := journalExpected.JournalName
	passwordHash := journalExpected.Password

	journal, err := database.GetJournalByName(context.Background(), journalName)

	if err != nil {
		t.Fatal(err)
	}

	if journal.JournalName != journalName {
		t.Errorf("got journalname = %q, want %q", journal.JournalName, journalName)
	}

	if journal.Password != passwordHash {
		t.Errorf("got password = %q, want %q", journal.Password, passwordHash)
	}
}

func TestUpdateJournalName(t *testing.T) {

	ctx := context.Background()
	database := testDB(t)

	_ = seedJournal(t, database)

	const (
		journalID int64  = 1
		newName   string = "New Name"
	)

	err := database.UpdateJournalName(ctx, journalID, newName)
	if err != nil {
		t.Fatalf("UpdateJournalName() error = %v", err)
	}

	journal, err := database.GetJournal(ctx, journalID)
	if err != nil {
		t.Fatalf("GetJournal() error = %v", err)
	}

	if journal.JournalName != newName {
		t.Errorf(
			"JournalName = %q, want %q",
			journal.JournalName,
			newName,
		)
	}

}

func TestUpdateJournalPassword(t *testing.T) {

	ctx := context.Background()
	database := testDB(t)

	_ = seedJournal(t, database)

	const (
		journalID   int64  = 1
		newPassword string = "new-password"
	)

	err := database.UpdateJournalPassword(ctx, journalID, newPassword)
	if err != nil {
		t.Fatalf("UpdateJournalName() error = %v", err)
	}

	journal, err := database.GetJournal(ctx, journalID)
	if err != nil {
		t.Fatalf("GetJournal() error = %v", err)
	}

	if journal.Password != newPassword {
		t.Errorf(
			"JournalName = %q, want %q",
			journal.Password,
			newPassword,
		)
	}

}

func TestDeleteJournal(t *testing.T) {

	ctx := context.Background()
	database := testDB(t)

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

	if err := database.DeleteJournal(ctx, journal.ID); err != nil {
		t.Fatal(err)
	}

	// Journal is gone
	_, err = database.GetJournal(ctx, journal.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}

	// Memory is gone
	_, err = database.GetMemory(ctx, journal.ID, memory.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("got error %v, want sql.ErrNoRows", err)
	}

	// Memory-tag relationship is gone
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

	// Tag survives
	_, err = database.GetTag(ctx, tag.ID)
	if err != nil {
		t.Errorf("tag was deleted with journal: %v", err)
	}
}
