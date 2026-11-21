package resume

import "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"

type UpdateInput struct {
	Resume          models.Resume
	ClearSalaryFrom bool
}
