package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"kogane/internal/database"

	"golang.org/x/crypto/bcrypt"
)

const SessionCookieName = "session_id"

var ErrInvalidCredentials = errors.New("credenciais inválidas")

type Service struct {
	db              *sql.DB
	dummyHash       []byte
	sessionDuration time.Duration
	development     bool
}

func NewService(
	db *sql.DB,
	development bool,
	bcryptCost int,
	sessionDuration time.Duration,
) (*Service, error) {
	dummyHash, err := bcrypt.GenerateFromPassword(
		[]byte("dummy-kogane"),
		bcryptCost,
	)
	if err != nil {
		return nil, err
	}

	return &Service{
		db:              db,
		dummyHash:       dummyHash,
		sessionDuration: sessionDuration,
		development:     development,
	}, nil
}

func (s *Service) Authenticate(
	username, password string,
) (int64, error) {
	user, err := database.GetUserByUsername(s.db, username)
	if err != nil {
		_, _ = bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.MinCost,
		)
		_ = bcrypt.CompareHashAndPassword(
			s.dummyHash,
			[]byte(password),
		)
		return 0, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Hash),
		[]byte(password),
	); err != nil {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}

func (s *Service) SetSessionCookie(
	w http.ResponseWriter,
	sessionID string,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   !s.development,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(s.sessionDuration.Seconds()),
		Expires:  time.Now().Add(s.sessionDuration),
	})
}

func (s *Service) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   SessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
