package runtimeconfig

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/runtimeconfig"
)

type Handler struct {
	settings runtimeconfig.Settings
}

func NewHandler(settings runtimeconfig.Settings) *Handler {
	return &Handler{settings: settings}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	config := h.settings.ForUser(userID)
	httputil.WriteOK(w, http.StatusOK, struct {
		Configs map[string]int  `json:"configs"`
		Flags   map[string]bool `json:"flags"`
	}{Configs: config.Configs, Flags: config.Flags})
}
