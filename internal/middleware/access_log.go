package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	if code >= 100 && code < 200 && code != http.StatusSwitchingProtocols {
		sw.ResponseWriter.WriteHeader(code)
		return
	}
	if sw.status != 0 {
		return
	}
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(p []byte) (int, error) {
	if sw.status == 0 {
		sw.WriteHeader(http.StatusOK)
	}
	return sw.ResponseWriter.Write(p)
}

func (sw *statusWriter) Unwrap() http.ResponseWriter {
	return sw.ResponseWriter
}

func AccessLog(mux *http.ServeMux, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{
			ResponseWriter: w,
		}
		next.ServeHTTP(sw, r)
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		_, pattern := mux.Handler(r)

		switch {
		case pattern == "":
			pattern = "unmatched"
		case !strings.HasPrefix(pattern, r.Method+" ") &&
			!(r.Method == http.MethodHead &&
				strings.HasPrefix(pattern, http.MethodGet+" ")):
			pattern = "redacted"
		}

		slog.InfoContext(
			r.Context(),
			"http request",
			"request_id", requestid.FromContext(r.Context()),
			"method", r.Method,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"route", pattern,
		)
	})
}
