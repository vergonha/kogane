package auth

import (
	"crypto/subtle"
	"net/http"
	"time"
)

func (s *Service) GetCSRFToken(r *http.Request) (string, bool) {
	sessionID, ok := s.sessionIDFromRequest(r)
	if !ok {
		return "", false
	}

	session, err := s.repository.Session.GetById(sessionID)
	if err != nil || session.CSRFToken == "" {
		return "", false
	}

	if time.Now().Unix() > session.ExpiresAt {
		_ = s.repository.Session.DeleteById(sessionID)
		return "", false
	}

	return session.CSRFToken, true
}

func (s *Service) RequireCSRF(r *http.Request) bool {
	sessionID, ok := s.sessionIDFromRequest(r)
	if !ok {
		return false
	}

	token := r.FormValue("csrf_token")
	if token == "" {
		return false
	}

	session, err := s.repository.Session.GetById(sessionID)
	if err != nil || session.CSRFToken == "" {
		return false
	}

	if time.Now().Unix() > session.ExpiresAt {
		_ = s.repository.Session.DeleteById(sessionID)
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(session.CSRFToken),
		[]byte(token),
	) == 1
}
