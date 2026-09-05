package handlers

import (
	"net/http"
	"strings"

	"kogane/internal/library"
)

func (h *Handler) Reader(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")
	mangadex_id := r.URL.Query().Get("mangadex_id")

	if !library.ValidComponent(title) ||
		!library.ValidComponent(vol) ||
		!strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	csrfToken, _ := h.Auth.GetCSRFToken(r)

	h.render(w, "reader.html", map[string]string{
		"Title":      title,
		"Vol":        vol,
		"CSRFToken":  csrfToken,
		"MangaDexID": mangadex_id,
	})
}
