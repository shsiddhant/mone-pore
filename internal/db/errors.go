package db

import (
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var ErrIntegrity = errors.New("database integrity violation")

// CheckError returns ErrIntegrity if error is sqlite3 constraint violation.
// Otherwise returns the original error.
func CheckError(err error) error {
	if err == nil {
		return nil
	}

	var sqliteErr *sqlite.Error

	if errors.Is(err, sqliteErr) {
		if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT {
			return ErrIntegrity
		}
	}
	return err
}
