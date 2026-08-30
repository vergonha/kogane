package database

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64
	Username string
	Hash     string
}

func EnsureAdminUser(
	db *sql.DB,
	password string,
	bcryptCost int,
) error {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcryptCost,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO users (username, hash)
		VALUES ('admin', ?)
		ON CONFLICT(username) DO UPDATE SET hash = excluded.hash
	`, string(hash))

	return err
}

func GetUserByUsername(db *sql.DB, username string) (User, error) {
	var user User

	err := db.QueryRow(`
		SELECT id, username, hash
		FROM users
		WHERE username = ?
	`, username).Scan(&user.ID, &user.Username, &user.Hash)

	return user, err
}
