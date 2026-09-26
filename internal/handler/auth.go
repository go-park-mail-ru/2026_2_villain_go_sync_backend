package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
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

func Register(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusCreated)
}
