package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"kogane/internal/database"
	"kogane/internal/library"
)

func (h *Handler) PDF(w http.ResponseWriter, r *http.Request, _ database.Session) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !library.ValidComponent(title) ||
		!library.ValidComponent(vol) ||
		!strings.HasSuffix(strings.ToLower(vol), ".pdf") {
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
