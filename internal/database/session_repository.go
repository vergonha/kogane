package database

import "database/sql"

type Session struct {
	ID        string
	UserID    int64
	ExpiresAt int64
	CSRFToken string
}

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session Session) error {
	_, err := r.db.Exec(`
		INSERT INTO sessions
			(id, user_id, expires_at, csrf_token)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.ExpiresAt, session.CSRFToken)

	return err
}

func (r *SessionRepository) GetById(id string) (Session, error) {
	var session Session

	err := r.db.QueryRow(`
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

func (r *SessionRepository) DeleteById(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *SessionRepository) DeleteByUserID(userID int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

func (r *SessionRepository) CleanupExpiredSessions(now int64) error {
	_, err := r.db.Exec(`
		DELETE FROM sessions
		WHERE expires_at <= ?
	`, now)

	return err
}
