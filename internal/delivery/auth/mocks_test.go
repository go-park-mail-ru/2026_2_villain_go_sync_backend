package auth

import (
	"context"

	authUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/auth"
)

type mockAuthUsecase struct {
	registerFn func(ctx context.Context, email, password, role string) (*authUsecase.Tokens, error)
	loginFn    func(ctx context.Context, email, password string) (*authUsecase.Tokens, error)
	refreshFn  func(ctx context.Context, refreshToken string) (*authUsecase.Tokens, error)
}

func (m *mockAuthUsecase) Register(ctx context.Context, email, password, role string) (*authUsecase.Tokens, error) {
	return m.registerFn(ctx, email, password, role)
}

func (m *mockAuthUsecase) Login(ctx context.Context, email, password string) (*authUsecase.Tokens, error) {
	return m.loginFn(ctx, email, password)
}

func (m *mockAuthUsecase) Refresh(ctx context.Context, refreshToken string) (*authUsecase.Tokens, error) {
	return m.refreshFn(ctx, refreshToken)
}
