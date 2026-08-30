package handlers

import (
	"net/http"
	"strings"
	"time"

	"kogane/internal/library"
)

func (h *Handler) PDF(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !library.ValidComponent(title) ||
		!library.ValidComponent(vol) {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	url, err := h.Storage.PresignGetObject(
		r.Context(),
		title+"/"+vol,
		15*time.Minute,
	)
	if err != nil {
		http.Error(
			w,
			"Erro ao gerar link de download",
			http.StatusInternalServerError,
		)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
