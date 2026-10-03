package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManager_GenerateAndParse_Access(t *testing.T) {
	accessTTL := 15 * time.Minute
	manager := NewTokenManager(
		[]byte("test-secret"),
		accessTTL,
		168*time.Hour,
	)

	token, err := manager.Generate(42, "seeker", TokenTypeAccess)
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}
	if token == "" {
		t.Fatal("Generate() returned an empty token")
	}

	claims, err := manager.Parse(token, TokenTypeAccess)
	if err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if claims == nil {
		t.Fatal("Parse() returned nil claims")
	}

	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Role != "seeker" {
		t.Errorf("Role = %q, want seeker", claims.Role)
	}
	if claims.TokenType != TokenTypeAccess {
		t.Errorf("TokenType = %q, want access", claims.TokenType)
	}
	if claims.Subject != "42" {
		t.Errorf("Subject = %q, want 42", claims.Subject)
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("token timestamps are missing")
	}
	if ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); ttl != accessTTL {
		t.Errorf("token TTL = %v, want %v", ttl, accessTTL)
	}
}

func TestTokenManager_GenerateAndParse_Refresh(t *testing.T) {
	refreshTTL := 168 * time.Hour
	manager := NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		refreshTTL,
	)

	token, err := manager.Generate(42, "seeker", TokenTypeRefresh)
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	claims, err := manager.Parse(token, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}
	if claims == nil {
		t.Fatal("Parse() returned nil claims")
	}

	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Role != "seeker" {
		t.Errorf("Role = %q, want seeker", claims.Role)
	}
	if claims.TokenType != TokenTypeRefresh {
		t.Errorf("TokenType = %q, want refresh", claims.TokenType)
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		t.Fatal("token timestamps are missing")
	}
	if ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); ttl != refreshTTL {
		t.Errorf("token TTL = %v, want %v", ttl, refreshTTL)
	}
}

func TestTokenManager_Parse_WrongType(t *testing.T) {
	tests := []struct {
		name      string
		generated TokenType
		expected  TokenType
	}{
		{
			name:      "access as refresh",
			generated: TokenTypeAccess,
			expected:  TokenTypeRefresh,
		},
		{
			name:      "refresh as access",
			generated: TokenTypeRefresh,
			expected:  TokenTypeAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTokenManager(
				[]byte("test-secret"),
				15*time.Minute,
				168*time.Hour,
			)

			token, err := manager.Generate(42, "seeker", tt.generated)
			if err != nil {
				t.Fatalf("Generate() returned an error: %v", err)
			}

			claims, err := manager.Parse(token, tt.expected)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("Parse() error = %v, want ErrInvalidToken", err)
			}
			if claims != nil {
				t.Error("Parse() returned claims for the wrong token type")
			}
		})
	}
}

func TestTokenManager_Parse_InvalidToken(t *testing.T) {
	manager := NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		168*time.Hour,
	)
	otherManager := NewTokenManager(
		[]byte("other-secret"),
		15*time.Minute,
		168*time.Hour,
	)

	foreignToken, err := otherManager.Generate(42, "seeker", TokenTypeAccess)
	if err != nil {
		t.Fatalf("Generate() returned an error: %v", err)
	}

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "malformed", token: "not-a-jwt"},
		{name: "wrong signature", token: foreignToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.Parse(tt.token, TokenTypeAccess)
			if !errors.Is(err, ErrInvalidToken) {
				t.Errorf("Parse() error = %v, want ErrInvalidToken", err)
			}
			if claims != nil {
				t.Error("Parse() returned claims for an invalid token")
			}
		})
	}
}

func TestTokenManager_Parse_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret")
	manager := NewTokenManager(secret, 15*time.Minute, 168*time.Hour)
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    42,
		Role:      "seeker",
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "42",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() returned an error: %v", err)
	}

	claims, err := manager.Parse(tokenString, TokenTypeAccess)
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("Parse() error = %v, want ErrExpiredToken", err)
	}
	if claims != nil {
		t.Error("Parse() returned claims for an expired token")
	}
}

func TestTokenManager_Generate_UnknownType(t *testing.T) {
	manager := NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		168*time.Hour,
	)

	token, err := manager.Generate(42, "seeker", TokenType("unknown"))
	if err == nil {
		t.Error("Generate() returned no error for an unknown token type")
	}
	if token != "" {
		t.Error("Generate() returned a token for an unknown token type")
	}
}

func TestTokenManager_DefaultTTL(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		typ     TokenType
		wantTTL time.Duration
	}{
		{
			name:    "access zero",
			ttl:     0,
			typ:     TokenTypeAccess,
			wantTTL: 15 * time.Minute,
		},
		{
			name:    "refresh zero",
			ttl:     0,
			typ:     TokenTypeRefresh,
			wantTTL: 168 * time.Hour,
		},
		{
			name:    "access negative",
			ttl:     -time.Hour,
			typ:     TokenTypeAccess,
			wantTTL: 15 * time.Minute,
		},
		{
			name:    "refresh negative",
			ttl:     -time.Hour,
			typ:     TokenTypeRefresh,
			wantTTL: 168 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewTokenManager([]byte("test-secret"), tt.ttl, tt.ttl)

			token, err := manager.Generate(42, "seeker", tt.typ)
			if err != nil {
				t.Fatalf("Generate() returned an error: %v", err)
			}

			claims, err := manager.Parse(token, tt.typ)
			if err != nil {
				t.Fatalf("Parse() returned an error: %v", err)
			}
			if claims == nil || claims.IssuedAt == nil || claims.ExpiresAt == nil {
				t.Fatal("token claims or timestamps are missing")
			}

			if ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); ttl != tt.wantTTL {
				t.Errorf("token TTL = %v, want %v", ttl, tt.wantTTL)
			}
		})
	}
}
