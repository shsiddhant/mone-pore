package db

import (
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var ErrUniqueConstraint = errors.New("database unique constraint violation")

// CheckError returns ErrIntegrity if error is sqlite3 unique constraint violation.
// Otherwise returns the original error.
func CheckError(err error) error {
	if err == nil {
		return nil
	}

	if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok {
		if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {

			return ErrUniqueConstraint
		}
	}
	return err
}
