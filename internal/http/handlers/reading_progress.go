package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/database"
)

// csrfOK rejects state-changing API calls that do not echo the session token.
func csrfOK(w http.ResponseWriter, r *http.Request, session database.Session) bool {
	if auth.ValidCSRF(session, r.Header.Get("X-CSRF-Token")) {
		return true
	}

	http.Error(w, "Invalid CSRF", http.StatusForbidden)
	return false
}

func (h *Handler) ProgressGetAll(w http.ResponseWriter, r *http.Request, session database.Session) {
	progress, err := h.Repository.ReadingProgress.GetAllByUser(session.UserID)
	if err != nil {
		log.Printf("list progress for user %d: %v", session.UserID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, progress)
}

func (h *Handler) ProgressGet(w http.ResponseWriter, r *http.Request, session database.Session) {
	progress, err := h.Repository.ReadingProgress.GetByUserAndManga(
		session.UserID,
		r.PathValue("mangadex_id"),
	)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	writeJSON(w, progress)
}

func (h *Handler) ProgressUpsert(w http.ResponseWriter, r *http.Request, session database.Session) {
	if !csrfOK(w, r, session) {
		return
	}

	var body struct {
		MangadexID string `json:"mangadex_id"`
		Volume     string `json:"volume"`
		Page       int    `json:"page"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.MangadexID == "" || body.Volume == "" || body.Page < 1 {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	err := h.Repository.ReadingProgress.Upsert(database.ReadingProgress{
		UserID:     session.UserID,
		MangadexID: body.MangadexID,
		Volume:     body.Volume,
		Page:       body.Page,
	})
	if err != nil {
		log.Printf("upsert progress for %s: %v", body.MangadexID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProgressComplete(w http.ResponseWriter, r *http.Request, session database.Session) {
	if !csrfOK(w, r, session) {
		return
	}

	mangadexID := r.PathValue("mangadex_id")

	if err := h.Repository.ReadingProgress.MarkCompleted(session.UserID, mangadexID); err != nil {
		log.Printf("complete progress for %s: %v", mangadexID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProgressDelete(w http.ResponseWriter, r *http.Request, session database.Session) {
	if !csrfOK(w, r, session) {
		return
	}

	mangadexID := r.PathValue("mangadex_id")

	if err := h.Repository.ReadingProgress.DeleteByUserAndManga(session.UserID, mangadexID); err != nil {
		log.Printf("delete progress for %s: %v", mangadexID, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
