package handlers

import (
	"net/http"

	"kogane/internal/database"
	"kogane/internal/library"
)

func (h *Handler) Reader(w http.ResponseWriter, r *http.Request, session database.Session) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !library.ValidComponent(title) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	manga, ok := h.Library.ByTitle(title)
	if !ok {
		http.NotFound(w, r)
		return
	}

	if !manga.HasVolume(vol) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	h.render(w, "reader.html", map[string]string{
		"Title":      title,
		"Vol":        vol,
		"VolLabel":   library.VolumeLabel(vol),
		"CSRFToken":  session.CSRFToken,
		"MangaDexID": manga.MangaDexID,
	})
}
