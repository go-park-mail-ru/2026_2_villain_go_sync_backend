package httputil

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteOK(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, response{Data: data})
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, response{Error: msg})
}
