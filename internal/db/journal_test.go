package db

import (
	"context"
	"testing"
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
