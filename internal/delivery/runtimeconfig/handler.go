package runtimeconfig

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/httputil"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/runtimeconfig"
)

type Settings interface {
	ForUser(userID int64) runtimeconfig.PublicConfig
}

type responseDTO struct {
	Configs map[string]int  `json:"configs"`
	Flags   map[string]bool `json:"flags"`
}
type Handler struct {
	settings Settings
}

func NewHandler(settings Settings) *Handler {
	return &Handler{settings: settings}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	config := h.settings.ForUser(userID)
	httputil.WriteOK(w, http.StatusOK, responseDTO{Configs: config.Configs, Flags: config.Flags})
}
