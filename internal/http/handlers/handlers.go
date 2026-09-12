package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/config"
	"kogane/internal/database"
	"kogane/internal/library"
	"kogane/internal/storage"
	apptemplates "kogane/internal/templates"
	"kogane/internal/turnstile"
)

type Handler struct {
	Config     config.Config
	Auth       *auth.Service
	Renderer   *apptemplates.Renderer
	Turnstile  *turnstile.Client
	Storage    *storage.Client
	Library    library.Library
	Repository *database.Repository
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	if err := h.Renderer.Render(w, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}
