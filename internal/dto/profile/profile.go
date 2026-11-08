package profile

import (
	"fmt"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type SeekerResponse struct {
	models.SeekerProfile
	AvatarURL *string `json:"avatar_url"`
}

type EmployerResponse struct {
	models.EmployerProfile
	LogoURL *string `json:"logo_url"`
}

func NewSeekerResponse(p models.SeekerProfile) SeekerResponse {
	response := SeekerResponse{
		SeekerProfile: p,
	}
	if p.AvatarKey != nil && *p.AvatarKey != "" {
		url := fmt.Sprintf("/api/seekers/%d/avatar", p.UserID)
		response.AvatarURL = &url
	}
	return response
}

func NewEmployerResponse(p models.EmployerProfile) EmployerResponse {
	response := EmployerResponse{
		EmployerProfile: p,
	}
	if p.LogoKey != nil && *p.LogoKey != "" {
		url := fmt.Sprintf("/api/employers/%d/logo", p.UserID)
		response.LogoURL = &url
	}
	return response
}
