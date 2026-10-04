package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"strings"
	"testing"
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

// --- helpers ---

func newTestTokens() *auth.TokenManager {
	return auth.NewTokenManager(
		[]byte("test-secret"),
		15*time.Minute,
		168*time.Hour,
	)
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// --- Register ---

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
			h := NewHandler(repo, nil, newTestTokens())

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/register",
				strings.NewReader(tt.body),
			)
			recorder := httptest.NewRecorder()

			h.Register(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s",
					recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if repo.createCalls != 0 {
				t.Errorf("Create() called %d times, want 0", repo.createCalls)
			}
			if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			// Проверяем, что тело — валидный JSON с полем error
			var resp map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
				t.Errorf("body is not JSON: %v", err)
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("body has no error field: %s", recorder.Body.String())
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
	tokens := newTestTokens()
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
	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
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
			if cookie.Expires.IsZero() {
				t.Error("cookie must have an expiration date")
			} else if !cookie.Expires.After(time.Now()) {
				t.Error("cookie expiration must be in the future")
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

// --- Login ---

func TestLogin_InvalidRequest(t *testing.T) {
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
			body:       `{"email":"invalid","password":"Password123"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid password",
			body:       `{"email":"test@example.com","password":"short"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{}
			h := NewHandler(repo, nil, newTestTokens())

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/login",
				strings.NewReader(tt.body),
			)
			recorder := httptest.NewRecorder()

			h.Login(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s",
					recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{}
	h := NewHandler(repo, nil, newTestTokens())

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{"email":"test@example.com","password":"Password123"}`),
	)
	recorder := httptest.NewRecorder()

	h.Login(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if len(recorder.Result().Cookies()) != 0 {
		t.Error("no cookies should be set on failed login")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := password.Hash("Password123")
	repo := &mockUserRepository{
		user: models.User{
			ID:           42,
			Email:        "test@example.com",
			Role:         "seeker",
			PasswordHash: hash,
		},
	}
	h := NewHandler(repo, nil, newTestTokens())

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{"email":"test@example.com","password":"WrongPass1"}`),
	)
	recorder := httptest.NewRecorder()

	h.Login(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if len(recorder.Result().Cookies()) != 0 {
		t.Error("no cookies should be set on failed login")
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := password.Hash("Password123")
	repo := &mockUserRepository{
		user: models.User{
			ID:           42,
			Email:        "test@example.com",
			Role:         "seeker",
			PasswordHash: hash,
		},
	}
	tokens := newTestTokens()
	h := NewHandler(repo, nil, tokens)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{"email":"test@example.com","password":"Password123"}`),
	)
	recorder := httptest.NewRecorder()

	h.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s",
			recorder.Code, http.StatusOK, recorder.Body.String())
	}

	cookies := recorder.Result().Cookies()
	if findCookie(cookies, "access_token") == nil {
		t.Error("access_token cookie missing")
	}
	if findCookie(cookies, "refresh_token") == nil {
		t.Error("refresh_token cookie missing")
	}
}

// --- Refresh ---

func TestRefresh_NoCookie(t *testing.T) {
	h := NewHandler(&mockUserRepository{}, nil, newTestTokens())

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	recorder := httptest.NewRecorder()

	h.Refresh(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRefresh_GarbageToken(t *testing.T) {
	h := NewHandler(&mockUserRepository{}, nil, newTestTokens())

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "garbage"})
	recorder := httptest.NewRecorder()

	h.Refresh(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRefresh_WrongTokenType(t *testing.T) {
	tokens := newTestTokens()
	h := NewHandler(&mockUserRepository{}, nil, tokens)

	access, _ := tokens.Generate(42, "seeker", auth.TokenTypeAccess)

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: access})
	recorder := httptest.NewRecorder()

	h.Refresh(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRefresh_Success(t *testing.T) {
	tokens := newTestTokens()
	h := NewHandler(&mockUserRepository{}, nil, tokens)

	refresh, _ := tokens.Generate(42, "seeker", auth.TokenTypeRefresh)

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: refresh})
	recorder := httptest.NewRecorder()

	h.Refresh(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s",
			recorder.Code, http.StatusOK, recorder.Body.String())
	}

	cookies := recorder.Result().Cookies()
	if findCookie(cookies, "access_token") == nil {
		t.Error("access_token cookie missing")
	}
	if findCookie(cookies, "refresh_token") == nil {
		t.Error("refresh_token cookie missing")
	}
}
