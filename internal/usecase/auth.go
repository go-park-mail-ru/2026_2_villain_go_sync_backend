package usecase

import (
	"context"
	"fmt"
	"regexp"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/password"
)

const (
	maxEmailLength    = 254
	minPasswordLength = 8
	maxPasswordLength = 72
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var (
	reUpper = regexp.MustCompile(`[A-Z]`)
	reLower = regexp.MustCompile(`[a-z]`)
	reDigit = regexp.MustCompile(`[0-9]`)
	reChars = regexp.MustCompile("^[A-Za-z0-9!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>/?`~]+$")
)

var allowedRoles = map[string]bool{
	"employer": true,
	"seeker":   true,
}

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
	if err := validateCredentials(email, pass); err != nil {
		return nil, err
	}

	if !isValidRole(role) {
		return nil, ErrInvalidRole
	}

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
		return nil, err // ErrEmailTaken пробросится наверх
	}

	return u.issueTokens(int64(user.ID), user.Role)
}

func (u *AuthUsecase) Login(ctx context.Context, email, pass string) (*Tokens, error) {
	if err := validateCredentials(email, pass); err != nil {
		return nil, err
	}

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

func validateCredentials(email, password string) error {
	if !isValidEmail(email) || !isValidPassword(password) {
		return ErrInvalidCredentials
	}

	return nil
}

func isValidEmail(email string) bool {
	if len(email) > maxEmailLength {
		return false
	}

	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return false
	}

	if !reUpper.MatchString(password) {
		return false
	}
	if !reLower.MatchString(password) {
		return false
	}
	if !reDigit.MatchString(password) {
		return false
	}
	if !reChars.MatchString(password) {
		return false
	}

	return true
}

func isValidRole(role string) bool {
	return allowedRoles[role]
}
