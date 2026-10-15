package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	tokens, err := h.Auth.Register(
		r.Context(),
		request.Email,
		request.Password,
		request.Role,
	)

	if errors.Is(err, usecase.ErrInvalidCredentials) {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if errors.Is(err, usecase.ErrInvalidRole) {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, apperrors.ErrEmailTaken) {
		httputil.WriteError(w, http.StatusConflict, "email already taken")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	setAuthCookies(w, tokens.Access, tokens.Refresh)
	httputil.WriteOK(w, http.StatusCreated, nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	tokens, err := h.Auth.Login(r.Context(), request.Email, request.Password)
	if errors.Is(err, usecase.ErrInvalidCredentials) {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	setAuthCookies(w, tokens.Access, tokens.Refresh)
	httputil.WriteOK(w, http.StatusOK, nil)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tokens, err := h.Auth.Refresh(r.Context(), cookie.Value)
	if errors.Is(err, usecase.ErrInvalidToken) {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	setAuthCookies(w, tokens.Access, tokens.Refresh)
	httputil.WriteOK(w, http.StatusOK, nil)
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}
