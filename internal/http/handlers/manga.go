package handlers

import "net/http"

func (h *Handler) MangaDetails(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	manga, ok := h.LibrarySvc.ByTitle(title)
	if !ok {
		http.NotFound(w, r)
		return
	}

	csrfToken, ok := h.Auth.GetCSRFToken(r)
	if !ok {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	h.render(w, "manga.html", map[string]any{
		"Manga":     manga,
		"CSRFToken": csrfToken,
	})
}
