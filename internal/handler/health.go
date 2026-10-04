package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
)

func Health(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteOK(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
