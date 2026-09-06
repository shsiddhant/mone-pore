package db

import (
	"embed"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// RunMigrations applies pending database migrations.
func RunMigrations(db *DB, logger *log.Logger) error {
	goose.SetBaseFS(embedMigrations)
	if logger != nil {
		goose.SetLogger(logger)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("Failed to set goose dialect: %v", err)
	}

	if logger != nil {
		logger.Println("Running database migrations...")
	}
	if err := goose.Up(db.DB, "migrations"); err != nil {
		return fmt.Errorf("Migration failed: %v", err)
	}
	return nil

}
