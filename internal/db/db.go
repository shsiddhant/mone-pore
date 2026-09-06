package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	// "os"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// SQLiteConfig contains configuration options for an SQLite database connection.
type SQLiteConfig struct {
	// DBPath is the path to the SQLite database file.
	DBPath string

	// BusyTimeout specifies how long SQLite should wait for a locked database
	// before returning SQLITE_BUSY, in milliseconds.
	BusyTimeout int

	// JournalMode specifies the SQLite journal mode, such as "WAL".
	JournalMode string

	// ForeignKeys enables or disables SQLite foreign key constraint enforcement.
	ForeignKeys bool

	// Cache specifies the SQLite cache mode.
	Cache string

	// Synchronous specifies the SQLite synchronous setting.
	Synchronous string
}

// ToDSN converts the SQLite configuration into a database/sql connection string.
func (c SQLiteConfig) ToDSN() string {
	val := url.Values{}

	if c.BusyTimeout > 0 {
		val.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", c.BusyTimeout))
	}
	if c.JournalMode != "" {
		val.Add("_pragma", fmt.Sprintf("journal_mode(%s)", c.JournalMode))
	}
	fkVal := "0"
	if c.ForeignKeys {
		fkVal = "1"
	}
	val.Add("_pragma", fmt.Sprintf("foreign_keys(%s)", fkVal))

	if c.Cache != "" {
		val.Add("_pragma", fmt.Sprintf("cache(%s)", c.Cache))
	}

	if c.Synchronous != "" {
		val.Add("_pragma", fmt.Sprintf("synchronous(%s)", c.Synchronous))
	}

	u := url.URL{
		Scheme:   "file",
		Opaque:   c.DBPath,
		RawQuery: val.Encode(),
	}
	return u.String()
}

// Open initializes the SQLite database from the given configuration.
func Open(cfg SQLiteConfig) (*DB, error) {

	// Open database connection
	sqlDB, err := sql.Open("sqlite", cfg.ToDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{sqlDB}

	return db, nil
}
