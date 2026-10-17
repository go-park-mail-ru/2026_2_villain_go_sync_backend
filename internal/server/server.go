package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	authDelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/auth"
	healthDelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/health"
	vacancyDelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
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

	authHandler := authDelivery.NewHandler(authUsecase)
	vacancyHandler := vacancyDelivery.NewHandler(vacancyUsecase)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", healthDelivery.Health)

	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/logout", authHandler.Logout)

	mux.Handle("GET /api/vacancies", middleware.Auth(tokens, http.HandlerFunc(vacancyHandler.ListVacancies)))

	return middleware.CORS(mux)
}
