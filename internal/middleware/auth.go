package middleware

import (
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
)

const bearerPrefix = "Bearer "

func Auth(tm *auth.TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(header, bearerPrefix)

		claims, err := tm.Parse(tokenString, auth.TokenTypeAccess)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := auth.WithUserID(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
