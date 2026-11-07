package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

var csrfExempt = map[string]bool{
	"/api/login":    true,
	"/api/register": true,
	"/api/refresh":  true,
}

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		if csrfExempt[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil {
			httputil.WriteError(w, http.StatusForbidden, "csrf token missing")
			return
		}

		header := r.Header.Get("X-CSRF-Token")
		if header == "" {
			httputil.WriteError(w, http.StatusForbidden, "csrf header missing")
			return
		}

		if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
			httputil.WriteError(w, http.StatusForbidden, "csrf token mismatch")
			return
		}

		next.ServeHTTP(w, r)
	})
}
