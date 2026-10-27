package resume

import (
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type UpserRequest struct {
	Title       sanitizer.SafeString `json:"title"`
	Description sanitizer.SafeString `json:"description"`
	Experience  sanitizer.SafeString `json:"experience"`
	Education   sanitizer.SafeString `json:"education"`
}

func (r UpserRequest) IsValid() error {
	if strings.TrimSpace(r.Title.String()) == "" {
		return errors.New("title is required")
	}
	return nil
}
