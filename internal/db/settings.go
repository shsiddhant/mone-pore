package db

import (
	"context"
	"time"
)

// Setting represents a setting in the database.
type Setting struct {
	Key      string
	Value    string
	Modified time.Time
}

// Admin password hash setting key.
const (
	SettingAdminPasswordHash string = "admin_password_hash"
)

// CreateSetting creates a new setting with given key and value.
// Returns newly created setting or an error if the operation failes.
func (db *DB) CreateSetting(
	ctx context.Context,
	key string,
	value string,
) (Setting, error) {

	var setting Setting

	queryString := `
	INSERT INTO setting (key, value)
	VALUES (?, ?)
	RETURNING key, value, modified
	`

	err := db.QueryRowContext(ctx, queryString, key, value).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Modified,
	)
	return setting, err
}

// GetSetting retrieves a setting by its key.
// Returns an error if setting cannot be retrieved.
func (db *DB) GetSetting(
	ctx context.Context,
	key string,
) (Setting, error) {

	var setting Setting

	queryString := `
	SELECT key, value, modified
	FROM setting
	WHERE key = ?
	`

	err := db.QueryRowContext(ctx, queryString, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Modified,
	)
	return setting, err
}

// UpdateSetting updates the value of an existing setting.
// It returns an error if the update fails.
func (db *DB) UpdateSetting(
	ctx context.Context,
	key string,
	newValue string,
) error {

	queryString := `
	UPDATE setting
	SET
		value = ?,
		modified = CURRENT_TIMESTAMP
	WHERE key = ?`

	_, err := db.ExecContext(ctx, queryString, newValue, key)

	return err
}
