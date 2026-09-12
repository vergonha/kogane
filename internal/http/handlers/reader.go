package handlers

import (
	"net/http"
	"strings"

	"kogane/internal/database"
	"kogane/internal/library"
)

func (h *Handler) Reader(w http.ResponseWriter, r *http.Request, session database.Session) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !library.ValidComponent(title) ||
		!library.ValidComponent(vol) ||
		!strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	manga, ok := h.Library.ByTitle(title)
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.render(w, "reader.html", map[string]string{
		"Title":      title,
		"Vol":        vol,
		"CSRFToken":  session.CSRFToken,
		"MangaDexID": manga.MangaDexID,
	})
}
