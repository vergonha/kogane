package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"kogane/internal/database"
)

func (h *Handler) MangaPreview(w http.ResponseWriter, r *http.Request) {
	manga, ok := h.Library.ByTitle(r.URL.Query().Get("title"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	h.render(w, "manga_preview.html", map[string]any{
		"Manga":   manga,
		"PageURL": h.Config.PublicURL + r.URL.RequestURI(),
	})
}

func (h *Handler) MangaDetails(w http.ResponseWriter, r *http.Request, session database.Session) {
	manga, ok := h.Library.ByTitle(r.URL.Query().Get("title"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	progress, err := h.Repository.ReadingProgress.GetByUserAndManga(
		session.UserID,
		manga.MangaDexID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("read progress for %s: %v", manga.MangaDexID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.render(w, "manga.html", map[string]any{
		"Manga":       manga,
		"CSRFToken":   session.CSRFToken,
		"Progress":    progress,
		"HasProgress": err == nil,
	})
}
