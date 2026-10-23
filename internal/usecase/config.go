package usecase

import (
	tokenauth "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
)

type Config struct {
	Auth    *auth.AuthUsecase
	Vacancy *vacancy.VacancyUsecase
	Resume  *resume.ResumeUsecase
	Profile *profile.ProfileUsecase
}

func CreateConfig(repos *repository.Config, tokens *tokenauth.TokenManager) *Config {
	return &Config{
		Auth:    auth.NewAuthUsecase(repos.User, tokens),
		Vacancy: vacancy.NewVacancyUsecase(repos.Vacancy),
		Resume:  resume.NewResumeUsecase(repos.Resume),
		Profile: profile.NewProfileUsecase(repos.Profile),
	}
}
