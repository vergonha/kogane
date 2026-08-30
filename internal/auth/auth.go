package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"kogane/internal/database"

	"golang.org/x/crypto/bcrypt"
)

const SessionCookieName = "session_id"

var (
	ErrInvalidCredentials  = errors.New("credenciais inválidas")
	ErrInvalidRegistration = errors.New("dados de cadastro inválidos")
	ErrPasswordsDoNotMatch = errors.New("senhas não conferem")
)

type Service struct {
	db              *sql.DB
	dummyHash       []byte
	sessionDuration time.Duration
	development     bool
	bcryptCost      int
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
		bcryptCost:      bcryptCost,
	}, nil
}

func (s *Service) Authenticate(
	username, password string,
) (int64, error) {
	user, err := database.GetUserByUsername(
		s.db,
		normalizeUsername(username),
	)
	if err != nil {
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

func (s *Service) AdminExists() (bool, error) {
	return database.AdminExists(s.db)
}

func (s *Service) CreateUser(
	username, password string,
	isAdmin bool,
) error {
	username = normalizeUsername(username)

	if err := validateNewUserInput(username, password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.bcryptCost,
	)
	if err != nil {
		return err
	}

	return database.CreateUser(
		s.db,
		username,
		string(hash),
		isAdmin,
	)
}

func (s *Service) CreateInitialAdmin(
	username, password, confirm string,
) (bool, error) {
	if password != confirm {
		return false, ErrPasswordsDoNotMatch
	}

	username = normalizeUsername(username)

	if err := validateNewUserInput(username, password); err != nil {
		return false, err
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.bcryptCost,
	)
	if err != nil {
		return false, err
	}

	return database.CreateInitialAdmin(
		s.db,
		username,
		string(hash),
	)
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

func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func validateNewUserInput(
	username, password string,
) error {
	if strings.TrimSpace(username) == "" {
		return ErrInvalidRegistration
	}

	if strings.TrimSpace(password) == "" {
		return ErrInvalidRegistration
	}

	return nil
}
