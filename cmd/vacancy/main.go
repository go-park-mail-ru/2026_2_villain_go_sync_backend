package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	vacancydelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/grpc/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/user"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/vacancy"
	vacancyusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadVacancy()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := repository.NewPostgresPool(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()
	loggingPool := repository.NewLoggingPool(pool)

	vacancyRepo := vacancy.NewPostgresRepository(loggingPool)
	userRepo := user.NewPostgresRepository(loggingPool)
	usecase := vacancyusecase.NewVacancyUsecase(vacancyRepo, userRepo)
	vacancyHandler := vacancydelivery.NewHandler(usecase)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("net listen: %v", err)
	}
	server := grpc.NewServer()
	vacancyv1.RegisterVacancyServiceServer(server, vacancyHandler)
	reflection.Register(server)
	log.Printf("vacancy gRPC server listening on %s", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
