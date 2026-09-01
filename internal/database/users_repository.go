package database

import "database/sql"

type User struct {
	ID       int64
	Username string
	Hash     string
	IsAdmin  bool
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username string, hash string, isAdmin bool) error {
	_, err := r.db.Exec(`
		INSERT INTO users (username, hash, is_admin)
		VALUES (?, ?, ?)
	`, username, hash, boolToInt(isAdmin))

	return err
}

func (r *UserRepository) CreateInitialAdmin(username string, hash string) (bool, error) {
	result, err := r.db.Exec(`
		INSERT INTO users (username, hash, is_admin)
		SELECT ?, ?, 1
		WHERE NOT EXISTS (
			SELECT 1
			FROM users
			WHERE is_admin = 1
		)
	`, username, hash)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}

func (r *UserRepository) AdminExists() (bool, error) {
	var count int

	err := r.db.QueryRow(`
		SELECT COUNT(1)
		FROM users
		WHERE is_admin = 1
	`).Scan(&count)

	return count > 0, err
}

func (r *UserRepository) GetByUsername(username string) (User, error) {
	var user User
	var isAdmin int

	err := r.db.QueryRow(`
		SELECT id, username, hash, is_admin
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Hash,
		&isAdmin,
	)
	if err != nil {
		return User{}, err
	}

	user.IsAdmin = isAdmin == 1
	return user, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
