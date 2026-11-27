package profile

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	seekerprofile "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/usecase/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/usecase/company"
)

//go:generate mockgen -source=profile.go -destination=mock_repository.go -package=profile
type ProfileRepository interface {
	GetSeekerByID(ctx context.Context, userID int) (models.SeekerProfile, error)
	UpsertSeeker(ctx context.Context, profile models.SeekerProfile) (models.SeekerProfile, error)
	UpdateSeekerAvatarKey(ctx context.Context, userID int, key *string) error
	GetEmployerByID(ctx context.Context, userID int) (models.EmployerProfile, error)
	UpsertEmployer(ctx context.Context, profile models.EmployerProfile) (models.EmployerProfile, error)
	UpdateEmployerLogoKey(ctx context.Context, userID int, key *string) error
}

type ImageStorage interface {
	Upload(ctx context.Context, key string, body io.ReadSeeker, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
}

type seekerUsecase = seekerprofile.ProfileUsecase

type ProfileUsecase struct {
	*company.CompanyUsecase
	*seekerUsecase
}

func NewProfileUsecase(profiles ProfileRepository, storage ImageStorage) *ProfileUsecase {
	return &ProfileUsecase{
		CompanyUsecase: company.NewCompanyUsecase(profiles, storage),
		seekerUsecase:  seekerprofile.NewProfileUsecase(profiles, storage),
	}
}
