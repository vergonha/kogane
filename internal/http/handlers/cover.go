package handlers

import (
	"log"
	"net/http"
	"time"

	"kogane/internal/library"
)

func (h *Handler) Cover(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	if !library.ValidComponent(title) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	url, err := h.Storage.PresignGetObject(
		r.Context(),
		title+"/cover.jpg",
		1*time.Hour,
	)
	if err != nil {
		log.Printf("error generating image link: %v", err)
		http.Error(
			w,
			"Error generating image link",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
