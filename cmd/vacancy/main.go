package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/s3"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/user"
	companydelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/delivery/company"
	vacancydelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/delivery/vacancy"
	companyrepo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/repository/company"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/repository/vacancy"
	companyusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/usecase/company"
	vacancyusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/usecase/vacancy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadVacancy()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	s3Config, err := config.LoadS3()
	if err != nil {
		log.Fatalf("load s3 config: %v", err)
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

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(s3Config.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s3Config.AccessKey, s3Config.SecretKey, "")),
	)
	if err != nil {
		log.Fatalf("load s3 SDK config: %v", err)
	}
	s3Client := awss3.NewFromConfig(sdkConfig, func(options *awss3.Options) {
		options.BaseEndpoint = aws.String(s3Config.Endpoint)
		options.UsePathStyle = true
	})
	storage := s3.NewAdapter(s3Client, s3Config.Bucket)
	companies := companyrepo.NewPostgresRepository(loggingPool)
	companyHandler := companydelivery.NewHandler(companyusecase.NewCompanyUsecase(companies, storage))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("net listen: %v", err)
	}
	server := grpc.NewServer()
	vacancyv1.RegisterVacancyServiceServer(server, vacancyHandler)
	vacancyv1.RegisterCompanyServiceServer(server, companyHandler)
	reflection.Register(server)
	log.Printf("vacancy gRPC server listening on %s", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
