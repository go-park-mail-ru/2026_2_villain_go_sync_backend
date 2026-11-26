package company

import (
	"context"
	"errors"
	"io"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type CompanyRepository interface {
	GetEmployerByID(ctx context.Context, userID int) (models.EmployerProfile, error)
	UpsertEmployer(ctx context.Context, profile models.EmployerProfile) (models.EmployerProfile, error)
	UpdateEmployerLogoKey(ctx context.Context, userID int, key *string) error
}

type ImageStorage interface {
	Upload(ctx context.Context, key string, body io.ReadSeeker, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, key string) error
}

type CompanyUsecase struct {
	companies CompanyRepository
	storage   ImageStorage
}

func NewCompanyUsecase(companies CompanyRepository, storage ImageStorage) *CompanyUsecase {
	return &CompanyUsecase{
		companies: companies,
		storage:   storage,
	}
}

func (u *CompanyUsecase) GetEmployerByID(ctx context.Context, employerID int) (models.EmployerProfile, error) {
	p, err := u.companies.GetEmployerByID(ctx, employerID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return models.EmployerProfile{}, apperrors.ErrNotFound
		}
		return models.EmployerProfile{}, err
	}
	return p, nil
}

func (u *CompanyUsecase) UpdateEmployer(ctx context.Context, userID int, role models.Role, p models.EmployerProfile) (models.EmployerProfile, error) {
	if !role.IsEmployer() || (p.Profile.UserID != 0 && p.Profile.UserID != userID) {
		return models.EmployerProfile{}, apperrors.ErrForbidden
	}
	p.Profile.UserID = userID
	return u.companies.UpsertEmployer(ctx, p)
}
