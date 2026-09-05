package handlers

import (
	"encoding/json"
	"net/http"

	"kogane/internal/database"
)

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) ProgressGetAll(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.Auth.GetSessionUserID(r)
	progress, err := h.Repository.ReadingProgress.GetAllByUser(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, progress)
}

func (h *Handler) ProgressGet(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.Auth.GetSessionUserID(r)
	mangadexID := r.PathValue("mangadex_id")

	p, err := h.Repository.ReadingProgress.GetByUserAndManga(userID, mangadexID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ProgressUpsert(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.RequireCSRFHeader(r) {
		http.Error(w, "Invalid CSRF", http.StatusForbidden)
		return
	}

	userID, _ := h.Auth.GetSessionUserID(r)

	var body struct {
		MangadexID string `json:"mangadex_id"`
		Volume     string `json:"volume"`
		Page       int    `json:"page"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MangadexID == "" || body.Volume == "" || body.Page < 1 {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	err := h.Repository.ReadingProgress.Upsert(database.ReadingProgress{
		UserID:     userID,
		MangadexID: body.MangadexID,
		Volume:     body.Volume,
		Page:       body.Page,
	})
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProgressComplete(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.RequireCSRFHeader(r) {
		http.Error(w, "Invalid CSRF", http.StatusForbidden)
		return
	}

	userID, _ := h.Auth.GetSessionUserID(r)
	mangadexID := r.PathValue("mangadex_id")

	if err := h.Repository.ReadingProgress.MarkCompleted(userID, mangadexID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ProgressDelete(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.RequireCSRFHeader(r) {
		http.Error(w, "Invalid CSRF", http.StatusForbidden)
		return
	}

	userID, _ := h.Auth.GetSessionUserID(r)
	mangadexID := r.PathValue("mangadex_id")

	if err := h.Repository.ReadingProgress.DeleteByUserAndManga(userID, mangadexID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
