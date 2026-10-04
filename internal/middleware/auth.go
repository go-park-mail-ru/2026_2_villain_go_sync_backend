package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

func Auth(tm *auth.TokenManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		claims, err := tm.Parse(cookie.Value, auth.TokenTypeAccess)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := auth.WithUserID(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
