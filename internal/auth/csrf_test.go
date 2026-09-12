package auth

import (
	"testing"

	"kogane/internal/database"
)

func TestValidCSRF(t *testing.T) {
	session := database.Session{CSRFToken: "abc123"}

	tests := []struct {
		name    string
		session database.Session
		token   string
		want    bool
	}{
		{"match", session, "abc123", true},
		{"mismatch", session, "abc124", false},
		{"empty token", session, "", false},
		{"empty session token", database.Session{}, "", false},
		{"empty session token, token sent", database.Session{}, "abc123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidCSRF(tt.session, tt.token); got != tt.want {
				t.Errorf("ValidCSRF(%q, %q) = %v, want %v",
					tt.session.CSRFToken, tt.token, got, tt.want)
			}
		})
	}
}
