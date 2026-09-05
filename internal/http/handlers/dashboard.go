package handlers

import "net/http"

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	csrfToken, ok := h.Auth.GetCSRFToken(r)
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	h.render(w, "dashboard.html", map[string]any{
		"Mangas":    h.LibrarySvc.All(),
		"CSRFToken": csrfToken,
	})
}
