package middleware

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/metrics"
)

func TestAccessLogMetrics(t *testing.T) {
	observer := metrics.NewHTTPMetrics()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	wrapped := AccessLog(mux, mux, observer)

	for _, path := range []string{"/api/items/123?token=secret-value", "/api/items/456", "/missing"} {
		wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	rec := httptest.NewRecorder()
	observer.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	lines := strings.Split(body, "\n")
	for _, want := range []string{
		`http_server_requests_total{method="GET",route="GET /api/items/{id}",status="201"} 2`,
		`http_server_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`http_server_request_duration_seconds_count{method="GET",route="GET /api/items/{id}",status="201"} 2`,
		`http_server_request_duration_seconds_count{method="GET",route="unmatched",status="404"} 1`,
	} {
		if !slices.Contains(lines, want) {
			t.Errorf("missing metric %q", want)
		}
	}
	for _, value := range []string{"/api/items/123", "/api/items/456", "/missing", "secret-value"} {
		if strings.Contains(body, value) {
			t.Errorf("metrics expose %q", value)
		}
	}
}
