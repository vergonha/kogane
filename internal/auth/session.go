package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"kogane/internal/database"
)

const SessionCookieName = "session_id"

// Handler is an http.HandlerFunc for a route behind RequireAuth, which has
// already loaded and validated the session.
type Handler func(w http.ResponseWriter, r *http.Request, session database.Session)

func (s *Service) RequireAuth(next Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := s.session(r)
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next(w, r, session)
	}
}

func ValidCSRF(session database.Session, token string) bool {
	if token == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(session.CSRFToken),
		[]byte(token),
	) == 1
}

func (s *Service) NewSession(userID int64) (database.Session, error) {
	sessionID, err := randomToken()
	if err != nil {
		return database.Session{}, err
	}

	csrfToken, err := randomToken()
	if err != nil {
		return database.Session{}, err
	}

	session := database.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.sessionDuration).Unix(),
		CSRFToken: csrfToken,
	}

	if err := s.repository.Session.Create(session); err != nil {
		return database.Session{}, err
	}

	return session, nil
}

func (s *Service) DeleteByUserID(userID int64) error {
	return s.repository.Session.DeleteByUserID(userID)
}

func (s *Service) DeleteSession(r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return
	}

	if err := s.repository.Session.DeleteById(cookie.Value); err != nil {
		log.Printf("delete session: %v", err)
	}
}

func (s *Service) session(r *http.Request) (database.Session, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return database.Session{}, false
	}

	session, err := s.repository.Session.GetById(cookie.Value)
	if err != nil {
		return database.Session{}, false
	}

	if time.Now().Unix() > session.ExpiresAt {
		if err := s.repository.Session.DeleteById(session.ID); err != nil {
			log.Printf("delete expired session: %v", err)
		}
		return database.Session{}, false
	}

	return session, true
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
