package auth

import (
	"crypto/rand"
	"encoding/hex"
	"kogane/internal/database"
	"net/http"
	"time"
)

func (s *Service) NewSession(userID int64) (string, string, error) {
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		return "", "", err
	}

	csrfBytes := make([]byte, 32)
	if _, err := rand.Read(csrfBytes); err != nil {
		return "", "", err
	}

	sessionID := hex.EncodeToString(sessionBytes)
	csrfToken := hex.EncodeToString(csrfBytes)

	session := database.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(s.sessionDuration).Unix(),
		CSRFToken: csrfToken,
	}

	if err := s.repository.Session.Create(session); err != nil {
		return "", "", err
	}

	return sessionID, csrfToken, nil
}

func (s *Service) DeleteByUserID(userID int64) error {
	return s.repository.Session.DeleteByUserID(userID)
}

func (s *Service) GetSessionUserID(r *http.Request) (int64, bool) {
	sessionID, ok := s.sessionIDFromRequest(r)
	if !ok {
		return 0, false
	}

	session, err := s.repository.Session.GetById(sessionID)
	if err != nil {
		return 0, false
	}

	if time.Now().Unix() > session.ExpiresAt {
		_ = s.repository.Session.DeleteById(sessionID)
		return 0, false
	}

	return session.UserID, true
}

func (s *Service) DeleteSession(r *http.Request) {
	sessionID, ok := s.sessionIDFromRequest(r)
	if !ok {
		return
	}

	_ = s.repository.Session.DeleteById(sessionID)
}

func (s *Service) sessionIDFromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}
