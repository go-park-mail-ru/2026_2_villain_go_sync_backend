package profile

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

type ProfileRepository interface {
	GetSeekerByID(ctx context.Context, userID int) (models.SeekerProfile, error)
	UpsertSeeker(ctx context.Context, profile models.SeekerProfile) (models.SeekerProfile, error)
	GetEmployerByID(ctx context.Context, userID int) (models.EmployerProfile, error)
	UpsertEmployer(ctx context.Context, profile models.EmployerProfile) (models.EmployerProfile, error)
}

type ProfileUsecase struct {
	profiles ProfileRepository
}

func NewProfileUsecase(profiles ProfileRepository) *ProfileUsecase {
	return &ProfileUsecase{profiles: profiles}
}

func (u *ProfileUsecase) GetSeekerByID(ctx context.Context, seekerID int) (models.SeekerProfile, error) {
	p, err := u.profiles.GetSeekerByID(ctx, seekerID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return models.SeekerProfile{}, apperrors.ErrNotFound
	}
	return p, err
}

func (u *ProfileUsecase) UpdateSeeker(ctx context.Context, userID int, role models.Role, p models.SeekerProfile) (models.SeekerProfile, error) {
	if !role.IsSeeker() {
		return models.SeekerProfile{}, apperrors.ErrForbidden
	}

	return u.profiles.UpsertSeeker(ctx, p)
}

func (u *ProfileUsecase) GetEmployerByID(ctx context.Context, employerID int) (models.EmployerProfile, error) {
	p, err := u.profiles.GetEmployerByID(ctx, employerID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return models.EmployerProfile{}, apperrors.ErrNotFound
		}
		return models.EmployerProfile{}, err
	}
	return p, nil
}

func (u *ProfileUsecase) UpdateEmployer(ctx context.Context, userID int, role models.Role, p models.EmployerProfile) (models.EmployerProfile, error) {
	if !role.IsEmployer() {
		return models.EmployerProfile{}, apperrors.ErrForbidden
	}

	return u.profiles.UpsertEmployer(ctx, p)
}
