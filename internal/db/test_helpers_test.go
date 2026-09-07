package db

import (
	"context"
	"database/sql"
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

func seedJournal(t *testing.T, db *DB) Journal {
	t.Helper()

	journalName := "Reconstruction"
	passwordHash := "password-hash"

	journal, err := db.CreateJournal(
		context.Background(),
		journalName,
		passwordHash,
	)

	if err != nil {
		t.Fatal(err)
	}

	return journal
}
