package resume

import (
	"errors"
	"strings"
)

type UpserRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Experience  string `json:"experience"`
	Education   string `json:"education"`
}

func (r UpserRequest) IsValid() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}
