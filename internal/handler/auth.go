package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
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

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var allowedRoles = map[string]bool{
	"employer": true,
	"seeker":   true,
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
		httputil.WriteError(w, http.StatusBadRequest, "invalid role")
		return
	}

	hash, err := password.Hash(request.Password)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := h.Storage.Create(models.User{
		Email:        request.Email,
		PasswordHash: hash,
		Role:         request.Role,
	})
	if errors.Is(err, apperrors.ErrEmailTaken) {
		httputil.WriteError(w, http.StatusConflict, "email already taken")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	accessToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeAccess)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	refreshToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeRefresh)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)

	httputil.WriteOK(w, http.StatusCreated, nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := validateCredentials(request.Email, request.Password); err != nil {
		httputil.WriteError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.Storage.GetByEmail(request.Email)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := password.Check(request.Password, user.PasswordHash); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeAccess)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	refreshToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeRefresh)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)

	httputil.WriteOK(w, http.StatusOK, nil)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	claims, err := h.Tokens.Parse(cookie.Value, auth.TokenTypeRefresh)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	accessToken, err := h.Tokens.Generate(claims.UserID, claims.Role, auth.TokenTypeAccess)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	refreshToken, err := h.Tokens.Generate(claims.UserID, claims.Role, auth.TokenTypeRefresh)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.setAuthCookies(w, accessToken, refreshToken)

	httputil.WriteOK(w, http.StatusOK, nil)
}

func (h *Handler) setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	now := time.Now()

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Expires:  now.Add(h.Tokens.AccessTTL()),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		Expires:  now.Add(h.Tokens.RefreshTTL()),
	})
}
