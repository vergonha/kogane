package handlers

import (
	"errors"
	"log"
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/database"
)

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	adminExists, err := h.Auth.AdminExists()
	if err != nil {
		log.Printf("check admin: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !adminExists {
		h.render(w, "register.html", map[string]string{})
		return
	}

	h.render(w, "login.html", map[string]string{
		"IP":               r.Header.Get("CF-Connecting-IP"),
		"UserAgent":        r.UserAgent(),
		"TurnstileSiteKey": h.Config.TurnstileSiteKey,
	})
}

func (h *Handler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	adminExists, err := h.Auth.AdminExists()
	if err != nil {
		log.Printf("check admin: %v", err)
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

	userID, err := h.Auth.Authenticate(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Drop any older sessions so a login cannot leave a stale one usable.
	if err := h.Auth.DeleteByUserID(userID); err != nil {
		log.Printf("clear sessions for user %d: %v", userID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !h.startSession(w, userID) {
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *Handler) registerInitialAdmin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	created, err := h.Auth.CreateInitialAdmin(username, password, r.FormValue("password_confirm"))
	switch {
	case errors.Is(err, auth.ErrInvalidRegistration),
		errors.Is(err, auth.ErrPasswordsDoNotMatch):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	case err != nil:
		log.Printf("create initial admin: %v", err)
		http.Error(w, "Could not create admin", http.StatusConflict)
		return
	}

	if !created {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := h.Auth.Authenticate(username, password)
	if err != nil {
		log.Printf("authenticate new admin: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !h.startSession(w, userID) {
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request, session database.Session) {
	if !auth.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Invalid CSRF", http.StatusForbidden)
		return
	}

	h.Auth.DeleteSession(r)
	h.Auth.ClearSessionCookie(w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// startSession reports whether the session was created; it writes the error
// response itself when it was not.
func (h *Handler) startSession(w http.ResponseWriter, userID int64) bool {
	session, err := h.Auth.NewSession(userID)
	if err != nil {
		log.Printf("create session for user %d: %v", userID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return false
	}

	h.Auth.SetSessionCookie(w, session.ID)
	return true
}
