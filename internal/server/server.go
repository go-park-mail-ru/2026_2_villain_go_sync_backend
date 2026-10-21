package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/app"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	healthDelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/health"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
)

func New() http.Handler {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repository.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	appConfig := app.CreateConfig(pool, tokens)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", healthDelivery.Health)

	mux.HandleFunc("POST /api/register", appConfig.Handlers.Auth.Register)
	mux.HandleFunc("POST /api/login", appConfig.Handlers.Auth.Login)
	mux.HandleFunc("POST /api/refresh", appConfig.Handlers.Auth.Refresh)
	mux.HandleFunc("POST /api/logout", appConfig.Handlers.Auth.Logout)

	mux.Handle("GET /api/me", middleware.Auth(tokens, http.HandlerFunc(appConfig.Handlers.Auth.Me)))

	mux.HandleFunc("GET /api/vacancies", appConfig.Handlers.Vacancy.ListVacancies)
	mux.HandleFunc("GET /api/vacancies/{id}", appConfig.Handlers.Vacancy.GetByID)
	mux.Handle("POST /api/vacancies", middleware.Auth(tokens, http.HandlerFunc(appConfig.Handlers.Vacancy.CreateVacancy)))
	mux.Handle("GET /api/employer/vacancies", middleware.Auth(tokens, http.HandlerFunc(appConfig.Handlers.Vacancy.ListVacanciesByEmployer)))
	mux.Handle("PATCH /api/vacancies/{id}", middleware.Auth(tokens, http.HandlerFunc(appConfig.Handlers.Vacancy.UpdateVacancy)))
	mux.Handle("DELETE /api/vacancies/{id}", middleware.Auth(tokens, http.HandlerFunc(appConfig.Handlers.Vacancy.Delete)))

	return middleware.CORS(mux)
}
