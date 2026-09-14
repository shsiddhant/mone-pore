package db

import (
	"context"
	"testing"
	"time"
)

const (
	seedISTOffset = int((5*time.Hour + 30*time.Minute) / time.Second)
	seedTitle     = "The Best Day"
	seedBody      = "It was the happiest day of my life."
)

var seedLoc = time.FixedZone("IST", seedISTOffset)
var seedMemoryDate = time.Date(2023, 12, 17, 0, 0, 0, 0, seedLoc)
var seedTags = []string{"bestday", "date"}

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

func seedMemory(t *testing.T, db *DB) Memory {
	t.Helper()

	journal := seedJournal(t, db)

	memory, err := db.CreateMemory(
		context.Background(),
		journal.ID,
		seedMemoryDate,
		seedTitle,
		seedBody,
		seedTags,
	)
	if err != nil {
		t.Fatal(err)
	}
	return memory
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
