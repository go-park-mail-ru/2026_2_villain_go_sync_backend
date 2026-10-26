package delivery

import (
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/applicant"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase"
)

type Config struct {
	Auth      *auth.Handler
	Vacancy   *vacancy.Handler
	Resume    *resume.Handler
	Profile   *profile.Handler
	Applicant *applicant.Handler
}

func CreateConfig(usecases *usecase.Config) *Config {
	return &Config{
		Auth:      auth.NewHandler(usecases.Auth),
		Vacancy:   vacancy.NewHandler(usecases.Vacancy),
		Resume:    resume.NewHandler(usecases.Resume),
		Profile:   profile.NewHandler(usecases.Profile),
		Applicant: applicant.NewHandler(usecases.Applicant),
	}
}
