package handler

import (
	"encoding/json"
	"net/http"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&request)

	if err == nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
