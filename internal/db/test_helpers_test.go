package db

import (
	"context"
	"testing"
)

func testDB(t *testing.T) *DB {
	t.Helper()

	cfg := SQLiteConfig{
		DBPath:      ":memory:",
		ForeignKeys: true,
	}

	database, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		database.Close()
	})

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

func TestForeignKeys(t *testing.T) {
	database := testDB(t)

	ctx := context.Background()

	var foreignKeys int

	err := database.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("foreign_keys = %d", foreignKeys)
}
