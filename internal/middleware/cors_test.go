package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_RegularRequest(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(next)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8000" {
		t.Errorf("unexpected Access-Control-Allow-Origin: %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods is empty")
	}

	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers is empty")
	}
}

func TestCORS_Preflight(t *testing.T) {
	called := false

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	handler := CORS(next)

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("next handler must not be called for OPTIONS request")
	}

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8000" {
		t.Errorf("unexpected Access-Control-Allow-Origin: %q", got)
	}
}
