package apphttp

import (
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/http/handlers"
	httpmw "kogane/internal/http/middleware"
)

func NewRouter(h *handlers.Handler, authService *auth.Service) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", h.LoginPage)
	mux.HandleFunc("POST /", h.LoginSubmit)
	mux.HandleFunc(
		"POST /logout",
		httpmw.RequireAuth(authService, h.Logout),
	)
	mux.HandleFunc(
		"GET /dashboard",
		httpmw.RequireAuth(authService, h.Dashboard),
	)
	mux.HandleFunc(
		"GET /read",
		httpmw.RequireAuth(authService, h.Reader),
	)
	mux.HandleFunc(
		"GET /pdf",
		httpmw.RequireAuth(authService, h.PDF),
	)
	mux.HandleFunc(
		"GET /cover",
		httpmw.RequireAuth(authService, h.Cover),
	)

	return mux
}
