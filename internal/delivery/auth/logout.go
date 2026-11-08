package auth

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

func (h *Handler) Logout(w http.ResponseWriter, _ *http.Request) {
	for _, cookie := range []http.Cookie{
		{Name: "access_token", Path: "/", HttpOnly: true},
		{Name: "refresh_token", Path: "/api/refresh", HttpOnly: true},
		{Name: "csrf_token", Path: "/", HttpOnly: false},
	} {
		cookie.SameSite = http.SameSiteLaxMode
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0)
		http.SetCookie(w, &cookie)
	}

	httputil.WriteOK(w, http.StatusOK, nil)
}
