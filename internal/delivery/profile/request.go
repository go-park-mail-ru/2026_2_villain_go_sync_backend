package profile

import (
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type UpdateSeekerRequest struct {
	FirstName   sanitizer.SafeString `json:"first_name"`
	LastName    sanitizer.SafeString `json:"last_name"`
	Phone       sanitizer.SafeString `json:"phone"`
	Description sanitizer.SafeString `json:"description"`
}

func (r UpdateSeekerRequest) IsValid() error {
	if strings.TrimSpace(r.FirstName.String()) == "" {
		return errors.New("first_name is required")
	}
	return nil
}

type UpdateEmployerRequest struct {
	CompanyName sanitizer.SafeString `json:"company_name"`
	Description sanitizer.SafeString `json:"description"`
	Website     sanitizer.SafeString `json:"website"`
}

func (r UpdateEmployerRequest) IsValid() error {
	if strings.TrimSpace(r.CompanyName.String()) == "" {
		return errors.New("company_name is required")
	}
	return nil
}
