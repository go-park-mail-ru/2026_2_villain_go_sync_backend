package server

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/handler"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/storage"
)

func New() http.Handler {
	mux := http.NewServeMux()

	h := &handler.Handler{
		Storage: storage.NewMemoryRepository(),
	}

	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("POST /api/register", h.Register)

	return mux
}
