package main

// тут пока тестово, т.к. я пока не знаю как конфиги будут собираться

import (
	"context"
	"log"
	"net"
	"time"

	pb "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/api/gen/auth/v1"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/database"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/tokenauth"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/auth/delivery"
	repository "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/auth/repository"
	usecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/auth/usecase"

	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	tokens := tokenauth.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	userRepo := repository.NewPostgresRepository(pool)
	authUC := usecase.NewAuthUsecase(userRepo, tokens)
	handler := delivery.NewAuthGRPCHandler(authUC)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, handler)

	log.Println("Auth gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
