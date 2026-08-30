package database

import "database/sql"

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt int64
	CSRFToken string
}

func CreateSession(db *sql.DB, session Session) error {
	_, err := db.Exec(`
		INSERT INTO sessions
			(id, user_id, expires_at, csrf_token)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.ExpiresAt, session.CSRFToken)

	return err
}

func GetSessionByID(db *sql.DB, id string) (Session, error) {
	var session Session

	err := db.QueryRow(`
		SELECT id, user_id, expires_at, csrf_token
		FROM sessions
		WHERE id = ?
	`, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CSRFToken,
	)

	return session, err
}

func DeleteSessionByID(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func DeleteSessionsByUserID(db *sql.DB, userID int64) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func CleanupExpiredSessions(db *sql.DB, now int64) error {
	_, err := db.Exec(`
		DELETE FROM sessions
		WHERE expires_at <= ?
	`, now)

	return err
}
