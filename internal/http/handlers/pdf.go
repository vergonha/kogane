package handlers

import (
	"log"
	"net/http"
	"time"

	"kogane/internal/database"
	"kogane/internal/library"
)

func (h *Handler) PDF(w http.ResponseWriter, r *http.Request, _ database.Session) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !library.ValidComponent(title) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	manga, ok := h.Library.ByTitle(title)
	if !ok || !manga.HasVolume(vol) {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	url, err := h.Storage.PresignGetObject(r.Context(), title+"/"+vol, 15*time.Minute)
	if err != nil {
		log.Printf("presign %s/%s: %v", title, vol, err)
		http.Error(w, "Error generating download link", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
