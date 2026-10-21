package resume

import (
	"errors"
	"strings"
)

type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Experience  string `json:"experience"`
	Education   string `json:"education"`
}

func (r CreateRequest) IsValid() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

type UpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Experience  string `json:"experience"`
	Education   string `json:"education"`
}

func (r UpdateRequest) IsValid() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}
