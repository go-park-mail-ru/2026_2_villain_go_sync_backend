package auth

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/password"
)

type UserRepository interface {
	Create(ctx context.Context, user models.User) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type Tokens struct {
	Access  string
	Refresh string
}

type AuthUsecase struct {
	users  UserRepository
	tokens *auth.TokenManager
}

func NewAuthUsecase(users UserRepository, tokens *auth.TokenManager) *AuthUsecase {
	return &AuthUsecase{
		users:  users,
		tokens: tokens,
	}
}

func (u *AuthUsecase) Register(ctx context.Context, email, pass, role string) (*Tokens, error) {
	hash, err := password.Hash(pass)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := u.users.Create(ctx, models.User{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	})
	if err != nil {
		return nil, err
	}

	return u.issueTokens(int64(user.ID), user.Role)
}

func (u *AuthUsecase) Login(ctx context.Context, email, pass string) (*Tokens, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := password.Check(pass, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u.issueTokens(int64(user.ID), user.Role)
}

func (u *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	claims, err := u.tokens.Parse(refreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return u.issueTokens(claims.UserID, claims.Role)
}

func (u *AuthUsecase) issueTokens(userID int64, role string) (*Tokens, error) {
	access, err := u.tokens.Generate(userID, role, auth.TokenTypeAccess)
	if err != nil {
		return nil, fmt.Errorf("generate access: %w", err)
	}

	refresh, err := u.tokens.Generate(userID, role, auth.TokenTypeRefresh)
	if err != nil {
		return nil, fmt.Errorf("generate refresh: %w", err)
	}

	return &Tokens{Access: access, Refresh: refresh}, nil
}
