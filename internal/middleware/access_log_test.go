package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLog(t *testing.T) {
	var buf bytes.Buffer

	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
	})

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	mux.HandleFunc("GET /api/implicit", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := RequestID(AccessLog(mux, mux))

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantRoute  string
	}{
		{
			name:       "explicit status",
			path:       "/api/items/123?token=secret-query",
			wantStatus: http.StatusCreated,
			wantRoute:  "GET /api/items/{id}",
		},
		{
			name:       "implicit status",
			path:       "/api/implicit",
			wantStatus: http.StatusOK,
			wantRoute:  "GET /api/implicit",
		},
		{
			name:       "unmatched route",
			path:       "/unknown?token=secret-query",
			wantStatus: http.StatusNotFound,
			wantRoute:  "unmatched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Authorization", "Bearer secret-auth-token")
			req.Header.Set("Cookie", "access_token=secret-cookie")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("decode log: %v", err)
			}

			if entry["msg"] != "http request" {
				t.Errorf("message = %v", entry["msg"])
			}

			if entry["method"] != http.MethodGet {
				t.Errorf("method = %v", entry["method"])
			}

			if entry["route"] != tt.wantRoute {
				t.Errorf("route = %v, want %q", entry["route"], tt.wantRoute)
			}

			if entry["status"] != float64(tt.wantStatus) {
				t.Errorf("status in log = %v", entry["status"])
			}

			id := rec.Header().Get("X-Request-ID")
			if id == "" || entry["request_id"] != id {
				t.Errorf("request_id = %v, header = %q", entry["request_id"], id)
			}

			duration, ok := entry["duration_ms"].(float64)
			if !ok || duration < 0 {
				t.Errorf("invalid duration_ms: %v", entry["duration_ms"])
			}

			if strings.Contains(buf.String(), "secret-") {
				t.Error("sensitive data leaked into access log")
			}
		})
	}
}

func TestStatusWriter(t *testing.T) {
	tests := []struct {
		name       string
		write      func(http.ResponseWriter)
		wantStatus int
	}{
		{
			name: "explicit status",
			write: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusCreated)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "implicit status",
			write: func(w http.ResponseWriter) {
				_, _ = w.Write([]byte("hello"))
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "first status wins",
			write: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusNotFound)
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			sw := &statusWriter{ResponseWriter: rec}

			tt.write(sw)

			if sw.status != tt.wantStatus {
				t.Errorf("recorded status = %d, want %d", sw.status, tt.wantStatus)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("response status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if sw.Unwrap() != rec {
				t.Error("Unwrap returned wrong ResponseWriter")
			}
		})
	}
}
