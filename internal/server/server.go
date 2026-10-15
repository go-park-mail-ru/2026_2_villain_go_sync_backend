package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"
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

	userRepo := repository.NewUserPostgresRepository(pool)
	vacancyRepo := repository.NewDemoVacancyRepository()

	authUsecase := usecase.NewAuthUsecase(userRepo, tokens)
	vacancyUsecase := usecase.NewVacancyUsecase(vacancyRepo)

	h := delivery.NewHandler(authUsecase, vacancyUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/refresh", h.Refresh)
	mux.HandleFunc("GET /api/health", delivery.Health)
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)
	mux.HandleFunc("POST /api/logout", h.Logout)
	mux.HandleFunc("GET /api/vacancies", h.ListVacancies)

	return middleware.CORS(mux)
}
