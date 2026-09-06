package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func testDB(t *testing.T) *DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	database := &DB{db}

	if err := RunMigrations(database, nil); err != nil {
		t.Fatal(err)
	}

	return database
}

func testSeededDB(t *testing.T) *DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	database := &DB{db}

	if err := RunMigrations(database, nil); err != nil {
		t.Fatal(err)
	}

	journalName := "Reconstruction"
	passwordHash := "password-hash"

	_, err = database.CreateJournal(
		context.Background(),
		journalName,
		passwordHash,
	)

	if err != nil {
		t.Fatal(err)
	}

	return database
}

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

func TestGetJournalByName(t *testing.T) {
	journalName := "Reconstruction"
	passwordHash := "password-hash"
	database := testSeededDB(t)

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
	database := testSeededDB(t)

	const (
		journalID int64  = 1
		newName   string = "New Name"
	)

	err := database.UpdateJournalName(ctx, journalID, newName)
	if err != nil {
		t.Fatalf("UpdateJournalName() error = %v", err)
	}

	journal, err := database.GetJournal(ctx, fmt.Sprint(journalID))
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
	database := testSeededDB(t)

	const (
		journalID   int64  = 1
		newPassword string = "new-password"
	)

	err := database.UpdateJournalPassword(ctx, journalID, newPassword)
	if err != nil {
		t.Fatalf("UpdateJournalName() error = %v", err)
	}

	journal, err := database.GetJournal(ctx, fmt.Sprint(journalID))
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
