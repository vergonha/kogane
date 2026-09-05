package apphttp

import (
	"net/http"
	"os"

	"kogane/internal/auth"
	"kogane/internal/http/handlers"
	httpmw "kogane/internal/http/middleware"
)

// noDirListingFS hides directory listings from http.FileServer: any request
// resolving to a directory is treated as not found instead of being rendered.
type noDirListingFS struct {
	http.FileSystem
}

func (fs noDirListingFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	if stat, err := f.Stat(); err == nil && stat.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}

	return f, nil
}

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
		"GET /manga",
		httpmw.RequireAuth(authService, h.MangaDetails),
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

	mux.Handle(
		"GET /ap0/",
		http.StripPrefix(
			"/ap0/",
			http.FileServer(noDirListingFS{http.Dir("./static")}),
		),
	)

	mux.HandleFunc(
		"GET /api/progress",
		httpmw.RequireAuth(authService, h.ProgressGetAll),
	)
	mux.HandleFunc(
		"GET /api/progress/{mangadex_id}",
		httpmw.RequireAuth(authService, h.ProgressGet),
	)
	mux.HandleFunc(
		"POST /api/progress",
		httpmw.RequireAuth(authService, h.ProgressUpsert),
	)
	mux.HandleFunc(
		"POST /api/progress/{mangadex_id}/complete",
		httpmw.RequireAuth(authService, h.ProgressComplete),
	)
	mux.HandleFunc(
		"DELETE /api/progress/{mangadex_id}",
		httpmw.RequireAuth(authService, h.ProgressDelete),
	)

	return mux
}
