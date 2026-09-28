package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess   TokenType = "access"
	TokenTypeRefresh  TokenType = "refresh"
	defaultAccessTTL            = 15 * time.Minute
	defaultRefreshTTL           = 168 * time.Hour
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID    int64     `json:"uid"`
	TokenType TokenType `json:"typ"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(secret []byte, accessTTL, refreshTTL time.Duration) *TokenManager {
	if accessTTL <= 0 {
		accessTTL = defaultAccessTTL
	}

	if refreshTTL <= 0 {
		refreshTTL = defaultRefreshTTL
	}

	return &TokenManager{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *TokenManager) ttlFor(typ TokenType) (time.Duration, error) {
	switch typ {
	case TokenTypeAccess:
		return m.accessTTL, nil

	case TokenTypeRefresh:
		return m.refreshTTL, nil

	default:
		return 0, fmt.Errorf("unknown token type: %s", typ)
	}
}

func (m *TokenManager) Generate(userID int64, typ TokenType) (string, error) {
	ttl, err := m.ttlFor(typ)
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := Claims{
		UserID:    userID,
		TokenType: typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(m.secret)
}

func (m *TokenManager) Parse(tokenString string, expected TokenType) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expected {
		return nil, fmt.Errorf("%w: expected %s, got %s",
			ErrInvalidToken, expected, claims.TokenType)
	}

	return claims, nil
}
