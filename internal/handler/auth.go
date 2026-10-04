package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
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

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
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
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := validateCredentials(request.Email, request.Password); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !isValidRole(request.Role) {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	hash, err := password.Hash(request.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := h.Storage.Create(models.User{
		Email:        request.Email,
		PasswordHash: hash,
		Role:         request.Role,
	})
	if errors.Is(err, apperrors.ErrEmailTaken) {
		http.Error(w, "email already taken", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	accessToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeAccess)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeRefresh)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := validateCredentials(request.Email, request.Password); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	user, err := h.Storage.GetByEmail(request.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := password.Check(request.Password, user.PasswordHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeAccess)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.Tokens.Generate(int64(user.ID), user.Role, auth.TokenTypeRefresh)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	claims, err := h.Tokens.Parse(request.RefreshToken, auth.TokenTypeRefresh)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.Tokens.Generate(claims.UserID, claims.Role, auth.TokenTypeAccess)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.Tokens.Generate(claims.UserID, claims.Role, auth.TokenTypeRefresh)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
