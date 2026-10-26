package profile

import (
	"errors"
	"strings"
)

type UpdateSeekerRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Description string `json:"description"`
}

func (r UpdateSeekerRequest) IsValid() error {
	if strings.TrimSpace(r.FirstName) == "" {
		return errors.New("first_name is required")
	}
	return nil
}

type UpdateEmployerRequest struct {
	CompanyName string `json:"company_name"`
	Description string `json:"description"`
	Website     string `json:"website"`
}

func (r UpdateEmployerRequest) IsValid() error {
	if strings.TrimSpace(r.CompanyName) == "" {
		return errors.New("company_name is required")
	}
	return nil
}
