package server

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/handler"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/storage"
)

func New() http.Handler {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	mux := http.NewServeMux()

	h := handler.NewHandler(
		storage.NewMemoryRepository(),
		storage.NewMemoryVacancyRepository(),
		tokens,
	)

	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("GET /api/vacancies", h.ListVacancies)

	return middleware.CORS(mux)
}
