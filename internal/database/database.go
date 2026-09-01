package database

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type Repository struct {
	User    *UserRepository
	Session *SessionRepository
}

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := Init(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		User:    NewUserRepository(db),
		Session: NewSessionRepository(db),
	}
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

	return err
}

func StartSessionCleanup(ctx context.Context, repo *Repository) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.Session.CleanupExpiredSessions(time.Now().Unix()); err != nil {
				log.Printf("session cleanup: %v", err)
			}
		}
	}
}
