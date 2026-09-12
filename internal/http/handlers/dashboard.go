package handlers

import (
	"net/http"

	"kogane/internal/database"
)

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request, session database.Session) {
	h.render(w, "dashboard.html", map[string]any{
		"Mangas":    h.Library,
		"CSRFToken": session.CSRFToken,
	})
}
