package database

import (
	"database/sql"
	"time"
)

type ReadingProgress struct {
	ID         int64
	UserID     int64
	MangadexID string
	Volume     string
	Page       int
	Completed  bool
	UpdatedAt  int64
}

type ReadingProgressRepository struct {
	db *sql.DB
}

func NewReadingProgressRepository(db *sql.DB) *ReadingProgressRepository {
	return &ReadingProgressRepository{db: db}
}

func (r *ReadingProgressRepository) Upsert(p ReadingProgress) error {
	p.UpdatedAt = time.Now().Unix()

	_, err := r.db.Exec(`
		INSERT INTO reading_progress
			(user_id, mangadex_id, volume, page, completed, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, mangadex_id) DO UPDATE SET
			volume     = excluded.volume,
			page   	   = excluded.page,
			completed  = excluded.completed,
			updated_at = excluded.updated_at
	`, p.UserID, p.MangadexID, p.Volume, p.Page, p.Completed, p.UpdatedAt)

	return err
}

func (r *ReadingProgressRepository) GetByUserAndManga(userID int64, mangadexID string) (ReadingProgress, error) {
	var p ReadingProgress

	err := r.db.QueryRow(`
		SELECT id, user_id, mangadex_id, volume, page, completed, updated_at
		FROM reading_progress
		WHERE user_id = ? AND mangadex_id = ?
	`, userID, mangadexID).Scan(
		&p.ID, &p.UserID, &p.MangadexID, &p.Volume, &p.Page, &p.Completed, &p.UpdatedAt,
	)

	return p, err
}

func (r *ReadingProgressRepository) GetAllByUser(userID int64) ([]ReadingProgress, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, mangadex_id, volume, page, completed, updated_at
		FROM reading_progress
		WHERE user_id = ?
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReadingProgress
	for rows.Next() {
		var p ReadingProgress
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.MangadexID, &p.Volume, &p.Page, &p.Completed, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, p)
	}

	return results, rows.Err()
}

func (r *ReadingProgressRepository) MarkCompleted(userID int64, mangadexID string) error {
	_, err := r.db.Exec(`
		UPDATE reading_progress
		SET completed = 1, updated_at = ?
		WHERE user_id = ? AND mangadex_id = ?
	`, time.Now().Unix(), userID, mangadexID)

	return err
}

func (r *ReadingProgressRepository) DeleteByUserAndManga(userID int64, mangadexID string) error {
	_, err := r.db.Exec(`
		DELETE FROM reading_progress WHERE user_id = ? AND mangadex_id = ?
	`, userID, mangadexID)

	return err
}
