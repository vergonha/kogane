package database

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func Init(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id       INTEGER PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			hash     TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			csrf_token TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	return ensureUsersSchema(db)
}
