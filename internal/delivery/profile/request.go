package profile

import (
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/structured"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/sanitizer"
)

type UpdateSeekerRequest struct {
	structured.CareerDetails
	FirstName         sanitizer.SafeString  `json:"first_name"`
	LastName          sanitizer.SafeString  `json:"last_name"`
	Phone             sanitizer.SafeString  `json:"phone"`
	Description       sanitizer.SafeString  `json:"description"`
	City              *sanitizer.SafeString `json:"city"`
	Messenger         *sanitizer.SafeString `json:"messenger"`
	ContactEmail      *sanitizer.SafeString `json:"contact_email"`
	PreferredPosition *sanitizer.SafeString `json:"preferred_position"`
	PreferredCities   *sanitizer.SafeString `json:"preferred_cities"`
	WorkFormat        *sanitizer.SafeString `json:"work_format"`
}

func (r UpdateSeekerRequest) IsValid() error {
	if strings.TrimSpace(r.FirstName.String()) == "" {
		return errors.New("first_name is required")
	}
	if strings.TrimSpace(r.LastName.String()) == "" {
		return errors.New("last_name is required")
	}
	switch r.WorkFormat.String() {
	case "", "office", "hybrid", "remote":
		return nil
	default:
		return errors.New("work_format must be office, hybrid or remote")
	}
}

type UpdateEmployerRequest struct {
	structured.CompanyDetails
	CompanyName   sanitizer.SafeString  `json:"company_name"`
	Description   sanitizer.SafeString  `json:"description"`
	Website       sanitizer.SafeString  `json:"website"`
	Industry      *sanitizer.SafeString `json:"industry"`
	TeamSize      *sanitizer.SafeString `json:"team_size"`
	MainCity      *sanitizer.SafeString `json:"main_city"`
	Benefits      *sanitizer.SafeString `json:"benefits"`
	HiringEmail   *sanitizer.SafeString `json:"hiring_email"`
	OfficeAddress *sanitizer.SafeString `json:"office_address"`
}

func (r UpdateEmployerRequest) IsValid() error {
	if strings.TrimSpace(r.CompanyName.String()) == "" {
		return errors.New("company_name is required")
	}
	return nil
}
