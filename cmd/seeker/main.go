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
	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/config"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/s3"
	profiledelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/delivery/profile"
	resumedelivery "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/delivery/resume"
	profilerepo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/repository/profile"
	resumerepo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/repository/resume"
	profileusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/usecase/profile"
	resumeusecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/usecase/resume"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.LoadSeeker()
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
	profiles := profilerepo.NewPostgresRepository(loggingPool)
	profileHandler := profiledelivery.NewHandler(profileusecase.NewProfileUsecase(profiles, storage))
	resumes := resumerepo.NewPostgresRepository(loggingPool)
	resumeHandler := resumedelivery.NewHandler(resumeusecase.NewResumeUsecase(resumes))

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("net listen: %v", err)
	}
	server := grpc.NewServer()
	seekerv1.RegisterSeekerServiceServer(server, profileHandler)
	seekerv1.RegisterResumeServiceServer(server, resumeHandler)
	reflection.Register(server)
	log.Printf("seeker gRPC server listening on %s", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}
