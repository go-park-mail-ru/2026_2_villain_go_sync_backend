package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/password"
	authUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/auth"
	gomock "go.uber.org/mock/gomock"
)

func newTestTokens() *auth.TokenManager {
	return auth.NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		168*time.Hour,
	)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	tm := newTestTokens()
	uc := authUsecase.NewAuthUsecase(mockRepo, tm)

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(models.User{ID: 1, Role: models.RoleSeeker}, nil)

	tokens, err := uc.Register(context.Background(), "test@example.com", "Password123", models.RoleSeeker)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil || tokens.Access == "" || tokens.Refresh == "" {
		t.Error("expected valid tokens, got empty")
	}
}

func TestAuthUsecase_Register_EmailTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	uc := authUsecase.NewAuthUsecase(mockRepo, newTestTokens())

	repoErr := errors.New("email already taken")
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(models.User{}, repoErr)

	_, err := uc.Register(context.Background(), "test@example.com", "Password123", models.RoleSeeker)

	if !errors.Is(err, repoErr) {
		t.Errorf("expected error %v, got %v", repoErr, err)
	}
}

func TestAuthUsecase_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	tm := newTestTokens()
	uc := authUsecase.NewAuthUsecase(mockRepo, tm)

	hash, _ := password.Hash("Password123")

	mockRepo.EXPECT().
		GetByEmail(gomock.Any(), "test@example.com").
		Return(models.User{ID: 1, Role: models.RoleSeeker, PasswordHash: hash}, nil)

	tokens, err := uc.Login(context.Background(), "test@example.com", "Password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil || tokens.Access == "" {
		t.Error("expected valid tokens, got empty")
	}
}

func TestAuthUsecase_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	uc := authUsecase.NewAuthUsecase(mockRepo, newTestTokens())

	mockRepo.EXPECT().
		GetByEmail(gomock.Any(), "test@example.com").
		Return(models.User{}, errors.New("db error"))

	_, err := uc.Login(context.Background(), "test@example.com", "Password123")

	if !errors.Is(err, authUsecase.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUsecase_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	uc := authUsecase.NewAuthUsecase(mockRepo, newTestTokens())

	hash, _ := password.Hash("CorrectPassword123")

	mockRepo.EXPECT().
		GetByEmail(gomock.Any(), "test@example.com").
		Return(models.User{ID: 1, Role: models.RoleSeeker, PasswordHash: hash}, nil)

	_, err := uc.Login(context.Background(), "test@example.com", "WrongPassword123")

	if !errors.Is(err, authUsecase.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUsecase_Refresh_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	tm := newTestTokens()
	uc := authUsecase.NewAuthUsecase(mockRepo, tm)

	validRefresh, _ := tm.Generate(1, models.RoleSeeker, auth.TokenTypeRefresh)

	tokens, err := uc.Refresh(context.Background(), validRefresh)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens == nil || tokens.Access == "" {
		t.Error("expected valid tokens, got empty")
	}
}

func TestAuthUsecase_Refresh_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authUsecase.NewMockUserRepository(ctrl)
	uc := authUsecase.NewAuthUsecase(mockRepo, newTestTokens())

	_, err := uc.Refresh(context.Background(), "garbage-token")

	if !errors.Is(err, authUsecase.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}
