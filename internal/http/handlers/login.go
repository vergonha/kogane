package handlers

import (
	"errors"
	"log"
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/config"
	"kogane/internal/database"
	"kogane/internal/library"
	"kogane/internal/storage"
	apptemplates "kogane/internal/templates"
	"kogane/internal/turnstile"
)

type Handler struct {
	Config     config.Config
	Auth       *auth.Service
	Renderer   *apptemplates.Renderer
	Turnstile  *turnstile.Client
	Storage    *storage.Client
	LibrarySvc *library.Service
	Repository *database.Repository
}

func New(
	cfg config.Config,
	authService *auth.Service,
	renderer *apptemplates.Renderer,
	turnstileClient *turnstile.Client,
	storageClient *storage.Client,
	librarySvc *library.Service,
	repository *database.Repository,
) *Handler {
	return &Handler{
		Config:     cfg,
		Auth:       authService,
		Renderer:   renderer,
		Turnstile:  turnstileClient,
		Storage:    storageClient,
		LibrarySvc: librarySvc,
		Repository: repository,
	}
}

func (h *Handler) render(
	w http.ResponseWriter,
	name string,
	data any,
) {
	if err := h.Renderer.Render(w, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	adminExists, err := h.Auth.AdminExists()
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !adminExists {
		h.render(w, "register.html", map[string]string{})
		return
	}

	data := map[string]string{
		"IP":               r.Header.Get("CF-Connecting-IP"),
		"UserAgent":        r.UserAgent(),
		"TurnstileSiteKey": h.Config.TurnstileSiteKey,
	}

	h.render(w, "login.html", data)
}

func (h *Handler) LoginSubmit(
	w http.ResponseWriter,
	r *http.Request,
) {
	adminExists, err := h.Auth.AdminExists()
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !adminExists {
		h.registerInitialAdmin(w, r)
		return
	}

	if !h.Turnstile.Verify(r) {
		http.Error(w, "Invalid CAPTCHA", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	userID, err := h.Auth.Authenticate(username, password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.Auth.DeleteByUserID(userID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	sessionID, _, err := h.Auth.NewSession(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.Auth.SetSessionCookie(w, sessionID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) registerInitialAdmin(
	w http.ResponseWriter,
	r *http.Request,
) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirm")

	created, err := h.Auth.CreateInitialAdmin(
		username,
		password,
		confirm,
	)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidRegistration):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, auth.ErrPasswordsDoNotMatch):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(
				w,
				"Could not create admin",
				http.StatusConflict,
			)
		}
		return
	}

	if !created {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := h.Auth.Authenticate(username, password)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	sessionID, _, err := h.Auth.NewSession(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.Auth.SetSessionCookie(w, sessionID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.RequireCSRFFormValue(r) {
		http.Error(w, "Invalid CSRF", http.StatusForbidden)
		return
	}

	h.Auth.DeleteSession(r)
	h.Auth.ClearSessionCookie(w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
