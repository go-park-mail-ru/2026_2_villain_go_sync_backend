package server

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/handler"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("POST /api/register", handler.Register)

	return mux
}
