package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	authUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/auth"
)

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	mockUC := &mockAuthUsecase{
		registerFn: func(ctx context.Context, email, password string, role models.Role) (*authUsecase.Tokens, error) {
			return &authUsecase.Tokens{Access: "test-access", Refresh: "test-refresh"}, nil
		},
	}
	h := NewHandler(mockUC)

	body := `{"email":"test@example.com","password":"Password123","role":"seeker"}`
	request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	h.Register(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	cookies := recorder.Result().Cookies()
	access := findCookie(cookies, "access_token")
	refresh := findCookie(cookies, "refresh_token")

	if access == nil || access.Value != "test-access" {
		t.Error("access_token cookie missing or wrong")
	}
	if refresh == nil || refresh.Value != "test-refresh" {
		t.Error("refresh_token cookie missing or wrong")
	}
}

func TestRegister_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       `{`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid request format",
			body:       `{"email":"invalid","password":"short","role":"seeker"}`,
			mockErr:    nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid role",
			body:       `{"email":"test@example.com","password":"Password123","role":"admin"}`,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "email taken",
			body:       `{"email":"test@example.com","password":"Password123","role":"seeker"}`,
			mockErr:    apperrors.ErrEmailTaken,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{
				registerFn: func(ctx context.Context, email, password string, role models.Role) (*authUsecase.Tokens, error) {
					return nil, tt.mockErr
				},
			}
			h := NewHandler(mockUC)

			request := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			h.Register(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
		})
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	mockUC := &mockAuthUsecase{
		loginFn: func(ctx context.Context, email, password string) (*authUsecase.Tokens, error) {
			return &authUsecase.Tokens{Access: "test-access", Refresh: "test-refresh"}, nil
		},
	}
	h := NewHandler(mockUC)

	request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"email":"test@example.com","password":"Password123"}`))
	recorder := httptest.NewRecorder()

	h.Login(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	cookies := recorder.Result().Cookies()
	if findCookie(cookies, "access_token") == nil {
		t.Error("access_token cookie missing")
	}
	if findCookie(cookies, "refresh_token") == nil {
		t.Error("refresh_token cookie missing")
	}
}

func TestLogin_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "invalid credentials",
			body:       `{"email":"test@example.com","password":"Password123"}`,
			mockErr:    authUsecase.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid format",
			body:       `{"email":"invalid","password":"short"}`,
			mockErr:    nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{
				loginFn: func(ctx context.Context, email, password string) (*authUsecase.Tokens, error) {
					return nil, tt.mockErr
				},
			}
			h := NewHandler(mockUC)

			request := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			h.Login(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if len(recorder.Result().Cookies()) != 0 {
				t.Error("no cookies should be set on failed login")
			}
		})
	}
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	mockUC := &mockAuthUsecase{
		refreshFn: func(ctx context.Context, refreshToken string) (*authUsecase.Tokens, error) {
			return &authUsecase.Tokens{Access: "new-access", Refresh: "new-refresh"}, nil
		},
	}
	h := NewHandler(mockUC)

	request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "valid-refresh"})
	recorder := httptest.NewRecorder()

	h.Refresh(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	cookies := recorder.Result().Cookies()
	access := findCookie(cookies, "access_token")
	if access == nil || access.Value != "new-access" {
		t.Error("access_token cookie missing or wrong")
	}
}

func TestRefresh_Errors(t *testing.T) {
	tests := []struct {
		name       string
		cookieVal  string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "no cookie",
			cookieVal:  "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			cookieVal:  "garbage",
			mockErr:    authUsecase.ErrInvalidToken,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockAuthUsecase{
				refreshFn: func(ctx context.Context, refreshToken string) (*authUsecase.Tokens, error) {
					return nil, tt.mockErr
				},
			}
			h := NewHandler(mockUC)

			request := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
			if tt.cookieVal != "" {
				request.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.cookieVal})
			}
			recorder := httptest.NewRecorder()

			h.Refresh(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
