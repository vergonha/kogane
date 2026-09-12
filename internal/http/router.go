package apphttp

import (
	"net/http"
	"os"
	"slices"
	"strings"

	"kogane/internal/auth"
	"kogane/internal/http/handlers"
)

var socialPreviewBots = []string{
	"facebookexternalhit",
	"twitterbot",
	"discordbot",
	"whatsapp",
	"telegrambot",
	"slackbot",
	"linkedinbot",
	"skypeuripreview",
}

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
	mux.HandleFunc("POST /logout", authService.RequireAuth(h.Logout))
	mux.HandleFunc("GET /dashboard", authService.RequireAuth(h.Dashboard))
	mux.HandleFunc("GET /read", authService.RequireAuth(h.Reader))
	mux.HandleFunc("GET /pdf", authService.RequireAuth(h.PDF))
	mux.HandleFunc("GET /cover", authService.RequireAuth(h.Cover))

	mux.HandleFunc("GET /api/progress", authService.RequireAuth(h.ProgressGetAll))
	mux.HandleFunc("GET /api/progress/{mangadex_id}", authService.RequireAuth(h.ProgressGet))
	mux.HandleFunc("POST /api/progress", authService.RequireAuth(h.ProgressUpsert))
	mux.HandleFunc("POST /api/progress/{mangadex_id}/complete", authService.RequireAuth(h.ProgressComplete))
	mux.HandleFunc("DELETE /api/progress/{mangadex_id}", authService.RequireAuth(h.ProgressDelete))

	//link previews are fetched by bots that carry no session, so they get the
	// public og/twitter meta page instead of a redirect to the login form.
	mangaDetails := authService.RequireAuth(h.MangaDetails)
	mux.HandleFunc("GET /manga", func(w http.ResponseWriter, r *http.Request) {
		if isSocialPreviewBot(r.UserAgent()) {
			h.MangaPreview(w, r)
			return
		}

		mangaDetails(w, r)
	})

	mux.Handle(
		"GET /ap0/",
		http.StripPrefix(
			"/ap0/",
			http.FileServer(noDirListingFS{http.Dir("./static")}),
		),
	)

	return mux
}

func isSocialPreviewBot(userAgent string) bool {
	ua := strings.ToLower(userAgent)

	return slices.ContainsFunc(socialPreviewBots, func(bot string) bool {
		return strings.Contains(ua, bot)
	})
}
