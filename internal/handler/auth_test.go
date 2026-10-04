package handler

import (
	"errors"
	"strings"
	"testing"

	"net/http"
	"net/http/httptest"

	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/password"
)

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"minimum length", "Abcdefg1", true},
		{"maximum length", "Aa1" + strings.Repeat("a", 69), true},
		{"empty", "", false},
		{"too short", "Aa1", false},
		{"too long", "Aa1" + strings.Repeat("a", 70), false},
		{"no uppercase", "password123", false},
		{"no lowercase", "PASSWORD123", false},
		{"no digit", "PasswordABC", false},
		{"forbidden character", "Password123 ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPassword(tt.password); got != tt.want {
				t.Errorf("isValidPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid", "test@example.com", true},
		{"maximum length", strings.Repeat("a", 242) + "@example.com", true},
		{"too long", strings.Repeat("a", 243) + "@example.com", false},
		{"empty", "", false},
		{"missing at", "test.example.com", false},
		{"missing local part", "@example.com", false},
		{"missing domain", "test@", false},
		{"contains space", "test user@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmail(tt.email); got != tt.want {
				t.Errorf("isValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{"seeker", true},
		{"employer", true},
		{"admin", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("role="+tt.role, func(t *testing.T) {
			if got := isValidRole(tt.role); got != tt.want {
				t.Errorf("isValidRole(%q) = %v, want %v",
					tt.role, got, tt.want)
			}
		})
	}
}

func TestValidateCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{"valid", "test@example.com", "Password123", false},
		{"invalid email", "invalid", "Password123", true},
		{"invalid password", "test@example.com", "short", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCredentials(tt.email, tt.password)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidCredentials) {
					t.Errorf("error = %v, want ErrInvalidCredentials", err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRegister_InvalidRequest(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email",
			body:       `{"email":"invalid","password":"Password123","role":"seeker"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid password",
			body:       `{"email":"test@example.com","password":"short","role":"seeker"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid role",
			body:       `{"email":"test@example.com","password":"Password123","role":"admin"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{}
			h := NewHandler(repo, nil, nil)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/register",
				strings.NewReader(tt.body),
			)
			recorder := httptest.NewRecorder()

			h.Register(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if repo.createCalls != 0 {
				t.Errorf("Create() called %d times, want 0", repo.createCalls)
			}
		})
	}
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepository{
		user: models.User{
			ID:    42,
			Email: "test@example.com",
			Role:  "seeker",
		},
	}
	tokens := auth.NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		168*time.Hour,
	)
	h := NewHandler(repo, nil, tokens)

	body := `{"email":"test@example.com","password":"Password123","role":"seeker"}`
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/register",
		strings.NewReader(body),
	)
	recorder := httptest.NewRecorder()

	h.Register(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s",
			recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if repo.createCalls != 1 {
		t.Errorf("Create() called %d times, want 1", repo.createCalls)
	}
	if repo.createdUser.Email != "test@example.com" {
		t.Errorf("saved email = %q, want test@example.com", repo.createdUser.Email)
	}
	if repo.createdUser.Role != "seeker" {
		t.Errorf("saved role = %q, want seeker", repo.createdUser.Role)
	}
	if err := password.Check("Password123", repo.createdUser.PasswordHash); err != nil {
		t.Errorf("saved password hash is invalid: %v", err)
	}

	cookies := make(map[string]*http.Cookie)
	for _, cookie := range recorder.Result().Cookies() {
		cookies[cookie.Name] = cookie
	}

	tests := []struct {
		name string
		typ  auth.TokenType
		path string
	}{
		{"access_token", auth.TokenTypeAccess, "/"},
		{"refresh_token", auth.TokenTypeRefresh, "/api/refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie, ok := cookies[tt.name]
			if !ok {
				t.Fatalf("missing cookie %q", tt.name)
			}
			if !cookie.HttpOnly {
				t.Error("cookie must be HttpOnly")
			}
			if cookie.Path != tt.path {
				t.Errorf("Path = %q, want %q", cookie.Path, tt.path)
			}
			if cookie.SameSite != http.SameSiteLaxMode {
				t.Errorf("SameSite = %v, want Lax", cookie.SameSite)
			}

			claims, err := tokens.Parse(cookie.Value, tt.typ)
			if err != nil {
				t.Fatalf("invalid token: %v", err)
			}
			if claims.UserID != 42 {
				t.Errorf("UserID = %d, want 42", claims.UserID)
			}
			if claims.Role != "seeker" {
				t.Errorf("role = %q, want seeker", claims.Role)
			}
		})
	}
}
