package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/app"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery"
	healthDelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/health"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/middleware"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/s3"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/runtimeconfig"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
)

const startupTimeout = 5 * time.Second

func New(ctx context.Context) (_ http.Handler, _ func(), err error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, errors.New("load backend configuration")
	}

	startup, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	pool, err := repository.NewPostgresPool(startup, cfg.DSN())
	if err != nil {
		return nil, nil, errors.New("connect product database")
	}

	defer func() {
		if err != nil {
			pool.Close()
		}
	}()
	sdkCfg, err := awsconfig.LoadDefaultConfig(startup, awsconfig.WithRegion(cfg.S3Region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")))
	if err != nil {
		return nil, nil, errors.New("configure S3")
	}

	s3Client := awss3.NewFromConfig(sdkCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = true
	})

	s3Adapter := s3.NewAdapter(s3Client, cfg.S3Bucket)

	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	settings := runtimeconfig.New(nil)
	closeResources := pool.Close
	if cfg.OnlineConfURL != "" {
		client, err := onlineconf.NewClient(cfg.OnlineConfURL, cfg.OnlineConfSnapshotSecret)
		if err != nil {
			return nil, nil, err
		}
		refreshCtx, stopRefresh := context.WithCancel(ctx)
		cache := onlineconf.StartCache(refreshCtx, client)
		settings = runtimeconfig.New(cache)
		closeResources = func() {
			stopRefresh()
			<-cache.Done()
			pool.Close()
		}
	}
	appConfig := app.CreateConfig(pool, tokens, s3Adapter, settings)
	return routes(appConfig.Handlers, tokens), closeResources, nil
}

func routes(handlers *delivery.Config, tokens *auth.TokenManager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", healthDelivery.Health)

	mux.HandleFunc("POST /api/register", handlers.Auth.Register)
	mux.HandleFunc("POST /api/login", handlers.Auth.Login)
	mux.HandleFunc("POST /api/refresh", handlers.Auth.Refresh)
	mux.HandleFunc("POST /api/logout", handlers.Auth.Logout)

	mux.Handle("GET /api/me", middleware.Auth(tokens, http.HandlerFunc(handlers.Auth.Me)))

	runtimeConfig := middleware.Auth(tokens, http.HandlerFunc(handlers.RuntimeConfig.Get))
	mux.HandleFunc("GET /api/runtime-config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Add("Vary", "Cookie")
		runtimeConfig.ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /api/vacancies", handlers.Vacancy.ListVacancies)
	mux.HandleFunc("GET /api/vacancies/{id}", handlers.Vacancy.GetByID)
	mux.Handle("POST /api/vacancies", middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.CreateVacancy)))
	mux.Handle("GET /api/employer/vacancies", middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.ListVacanciesByEmployer)))
	mux.Handle("PATCH /api/vacancies/{id}", middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.UpdateVacancy)))
	mux.Handle("DELETE /api/vacancies/{id}", middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.Delete)))
	mux.Handle("GET /api/vacancies/{id}/applicants", middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.ListApplicants)))
	mux.HandleFunc("GET /api/vacancy-categories", handlers.Vacancy.ListCategories)

	mux.Handle("POST /api/resumes",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Resume.Create)))
	mux.Handle("PUT /api/resumes/{id}",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Resume.Update)))
	mux.Handle("DELETE /api/resumes/{id}",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Resume.Delete)))

	mux.Handle("GET /api/resumes/{id}",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Resume.GetByID)))
	mux.HandleFunc("GET /api/resumes", handlers.Resume.ListBySeekerID)

	mux.HandleFunc("GET /api/seekers/{id}", handlers.Profile.GetSeekerByID)
	mux.HandleFunc("GET /api/seekers/{id}/avatar", handlers.Profile.GetSeekerAvatar)
	mux.Handle("POST /api/profile/seeker",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UpdateSeeker)))
	mux.Handle("PUT /api/profile/seeker",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UpdateSeeker)))
	mux.Handle(
		"PUT /api/me/avatar",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UploadSeekerAvatar)),
	)
	mux.Handle(
		"DELETE /api/me/avatar",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.DeleteSeekerAvatar)),
	)

	mux.HandleFunc("GET /api/employers/{id}", handlers.Profile.GetPublicEmployer)
	mux.HandleFunc("GET /api/employers/{id}/logo", handlers.Profile.GetEmployerLogo)
	mux.Handle("POST /api/profile/employer",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UpdateEmployer)))
	mux.Handle("PUT /api/profile/employer",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UpdateEmployer)))
	mux.Handle(
		"PUT /api/me/logo",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.UploadEmployerLogo)),
	)
	mux.Handle(
		"DELETE /api/me/logo",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Profile.DeleteEmployerLogo)),
	)
	mux.Handle(
		"POST /api/vacancies/{id}/applications",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Application.Create)),
	)
	mux.Handle(
		"GET /api/employer/applications",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Vacancy.ListEmployerApplicants)),
	)

	mux.Handle(
		"PATCH /api/employer/applications/{id}/status",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Application.UpdateStatus)),
	)
	mux.Handle(
		"GET /api/seeker/applications",
		middleware.Auth(tokens, http.HandlerFunc(handlers.Application.ListMine)),
	)

	return middleware.RequestID(
		middleware.AccessLog(
			mux,
			middleware.CORS(middleware.CSRF(mux)),
		),
	)
}
