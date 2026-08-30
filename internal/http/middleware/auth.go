package middleware

import (
	"net/http"

	"kogane/internal/auth"
)

func RequireAuth(
	authService *auth.Service,
	next http.HandlerFunc,
) http.HandlerFunc {
	return authService.RequireAuth(next)
}
