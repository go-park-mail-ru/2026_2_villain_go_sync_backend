package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/password"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/storage"
)

const (
	maxEmailLength    = 254
	minPasswordLength = 8
	maxPasswordLength = 72
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	Storage storage.UserRepository
}

func isValidEmail(email string) bool {
	if len(email) > maxEmailLength {
		return false
	}

	return emailRegex.MatchString(email)
}

func isValidPassword(password string) bool {
	passwordLength := len(password)

	return minPasswordLength <= passwordLength && passwordLength <= maxPasswordLength
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !isValidEmail(request.Email) {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	if !isValidPassword(request.Password) {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	hash, err := password.Hash(request.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = h.Storage.Create(storage.User{
		Email:        request.Email,
		PasswordHash: hash,
	})
	if errors.Is(err, storage.ErrEmailTaken) {
		http.Error(w, "email already taken", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
