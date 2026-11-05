package profile

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

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

type ProfileUsecase struct {
	profiles ProfileRepository
	storage  ImageStorage
}

func NewProfileUsecase(profiles ProfileRepository, storage ImageStorage) *ProfileUsecase {
	return &ProfileUsecase{
		profiles: profiles,
		storage:  storage,
	}
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

func (u *ProfileUsecase) UploadSeekerAvatar(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error {
	if !role.IsSeeker() {
		return apperrors.ErrForbidden
	}
	p, err := u.profiles.GetSeekerByID(ctx, userID)
	if err != nil {
		return err
	}
	return u.replaceProfileImage(
		ctx,
		userID,
		"avatars",
		p.AvatarKey,
		body,
		contentType,
		u.profiles.UpdateSeekerAvatarKey,
	)
}

func (u *ProfileUsecase) GetSeekerAvatar(ctx context.Context, userID int) (io.ReadCloser, string, error) {
	p, err := u.profiles.GetSeekerByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if p.AvatarKey == nil {
		return nil, "", apperrors.ErrNotFound
	}
	body, contentType, err := u.storage.Download(ctx, *p.AvatarKey)
	if err != nil {
		return nil, "", fmt.Errorf("get avatar from s3 storage: %w", err)
	}
	return body, contentType, nil
}

func (u *ProfileUsecase) DeleteSeekerAvatar(ctx context.Context, userID int, role models.Role) error {
	if !role.IsSeeker() {
		return apperrors.ErrForbidden
	}
	p, err := u.profiles.GetSeekerByID(ctx, userID)
	if err != nil {
		return err
	}
	return u.deleteProfileImage(
		ctx,
		userID,
		p.AvatarKey,
		u.profiles.UpdateSeekerAvatarKey,
	)
}

func (u *ProfileUsecase) UploadEmployerLogo(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error {
	if !role.IsEmployer() {
		return apperrors.ErrForbidden
	}
	p, err := u.profiles.GetEmployerByID(ctx, userID)
	if err != nil {
		return err
	}
	return u.replaceProfileImage(
		ctx,
		userID,
		"logos",
		p.LogoKey,
		body,
		contentType,
		u.profiles.UpdateEmployerLogoKey,
	)
}

func (u *ProfileUsecase) GetEmployerLogo(ctx context.Context, userID int) (io.ReadCloser, string, error) {
	p, err := u.profiles.GetEmployerByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if p.LogoKey == nil {
		return nil, "", apperrors.ErrNotFound
	}
	body, contentType, err := u.storage.Download(ctx, *p.LogoKey)
	if err != nil {
		return nil, "", fmt.Errorf("get logo from s3 storage: %w", err)
	}
	return body, contentType, nil
}

func (u *ProfileUsecase) DeleteEmployerLogo(ctx context.Context, userID int, role models.Role) error {
	if !role.IsEmployer() {
		return apperrors.ErrForbidden
	}
	p, err := u.profiles.GetEmployerByID(ctx, userID)
	if err != nil {
		return err
	}
	return u.deleteProfileImage(
		ctx,
		userID,
		p.LogoKey,
		u.profiles.UpdateEmployerLogoKey,
	)
}
