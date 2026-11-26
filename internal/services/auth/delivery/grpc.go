package delivery

import (
	"context"
	"errors"

	pb "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	authUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/auth/usecase"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	auth *authUsecase.AuthUsecase
}

func NewAuthGRPCHandler(auth *authUsecase.AuthUsecase) *AuthGRPCHandler {
	return &AuthGRPCHandler{auth: auth}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := validateCredentials(req.Email, req.Password); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	role := models.Role(req.Role)
	if !role.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	tokens, err := h.auth.Register(ctx, req.Email, req.Password, role)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmailTaken) {
			return nil, status.Error(codes.AlreadyExists, "email already taken")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RegisterResponse{
		Tokens: &pb.AuthTokens{
			AccessToken:  tokens.Access,
			RefreshToken: tokens.Refresh,
		},
	}, nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if !isValidEmail(req.Email) || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	tokens, err := h.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, authUsecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.LoginResponse{
		Tokens: &pb.AuthTokens{
			AccessToken:  tokens.Access,
			RefreshToken: tokens.Refresh,
		},
	}, nil
}

func (h *AuthGRPCHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	tokens, err := h.auth.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, authUsecase.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.RefreshResponse{
		Tokens: &pb.AuthTokens{
			AccessToken:  tokens.Access,
			RefreshToken: tokens.Refresh,
		},
	}, nil
}
