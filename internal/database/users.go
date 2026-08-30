package database

import "database/sql"

type User struct {
	ID       int64
	Username string
	Hash     string
	IsAdmin  bool
}

func GetUserByUsername(db *sql.DB, username string) (User, error) {
	var user User
	var isAdmin int

	err := db.QueryRow(`
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

func CreateUser(
	db *sql.DB,
	username, hash string,
	isAdmin bool,
) error {
	_, err := db.Exec(`
		INSERT INTO users (username, hash, is_admin)
		VALUES (?, ?, ?)
	`, username, hash, boolToInt(isAdmin))

	return err
}

func CreateInitialAdmin(
	db *sql.DB,
	username, hash string,
) (bool, error) {
	result, err := db.Exec(`
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

func AdminExists(db *sql.DB) (bool, error) {
	var count int

	err := db.QueryRow(`
		SELECT COUNT(1)
		FROM users
		WHERE is_admin = 1
	`).Scan(&count)

	return count > 0, err
}

func ensureUsersSchema(db *sql.DB) error {
	hasIsAdmin, err := userTableHasColumn(db, "is_admin")
	if err != nil {
		return err
	}

	if !hasIsAdmin {
		if _, err := db.Exec(`
			ALTER TABLE users
			ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0
		`); err != nil {
			return err
		}
	}

	_, err = db.Exec(`
		UPDATE users
		SET is_admin = 1
		WHERE username = 'admin'
	`)

	return err
}

func userTableHasColumn(db *sql.DB, name string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(users)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var columnName string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int

		if err := rows.Scan(
			&cid,
			&columnName,
			&columnType,
			&notNull,
			&defaultValue,
			&pk,
		); err != nil {
			return false, err
		}

		if columnName == name {
			return true, nil
		}
	}

	return false, rows.Err()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
