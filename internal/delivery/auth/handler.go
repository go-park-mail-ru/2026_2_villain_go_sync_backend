package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	id "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	authUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/auth"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password, role string) (*authUsecase.Tokens, error)
	Login(ctx context.Context, email, password string) (*authUsecase.Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*authUsecase.Tokens, error)
}

type Handler struct {
	auth AuthUsecase
}

func NewHandler(auth AuthUsecase) *Handler {
	return &Handler{auth: auth}
}

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

	if err := validateCredentials(request.Email, request.Password); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if !isValidRole(request.Role) {
		httputil.WriteError(w, http.StatusBadRequest, ErrInvalidRole.Error())
		return
	}

	tokens, err := h.auth.Register(
		r.Context(),
		request.Email,
		request.Password,
		request.Role,
	)

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

	if !isValidEmail(request.Email) || request.Password == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "invalid request")
		return
	}

	tokens, err := h.auth.Login(r.Context(), request.Email, request.Password)

	if errors.Is(err, authUsecase.ErrInvalidCredentials) {
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

	tokens, err := h.auth.Refresh(r.Context(), cookie.Value)
	if errors.Is(err, authUsecase.ErrInvalidToken) {
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

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := id.UserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	data := map[string]int64{
		"id": userID,
	}
	httputil.WriteOK(w, http.StatusOK, data)
}
