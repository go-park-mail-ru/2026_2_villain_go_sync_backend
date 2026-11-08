package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/requestid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		id := hex.EncodeToString(b)

		ctx := requestid.WithID(r.Context(), id)

		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
