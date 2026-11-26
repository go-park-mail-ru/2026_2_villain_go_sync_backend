package profile

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/vacancy/usecase/company"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/profileimage"
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

type ProfileUsecase struct {
	*company.CompanyUsecase
	profiles ProfileRepository
	storage  ImageStorage
}

func NewProfileUsecase(profiles ProfileRepository, storage ImageStorage) *ProfileUsecase {
	return &ProfileUsecase{
		CompanyUsecase: company.NewCompanyUsecase(profiles, storage),
		profiles:       profiles,
		storage:        storage,
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
	if !role.IsSeeker() || (p.Profile.UserID != 0 && p.Profile.UserID != userID) {
		return models.SeekerProfile{}, apperrors.ErrForbidden
	}
	if value := p.Preferences.WorkFormat; value != nil && *value != "" && !value.IsValid() {
		return models.SeekerProfile{}, apperrors.ValidationError("work_format must be office, hybrid or remote")
	}

	p.Profile.UserID = userID
	return u.profiles.UpsertSeeker(ctx, p)
}

func (u *ProfileUsecase) UploadSeekerAvatar(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error {
	if !role.IsSeeker() {
		return apperrors.ErrForbidden
	}
	p, err := u.profiles.GetSeekerByID(ctx, userID)
	if err != nil {
		return err
	}
	return profileimage.Replace(
		ctx,
		u.storage,
		userID,
		"avatars",
		p.PersonalInfo.AvatarKey,
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
	if p.PersonalInfo.AvatarKey == nil {
		return nil, "", apperrors.ErrNotFound
	}
	body, contentType, err := u.storage.Download(ctx, *p.PersonalInfo.AvatarKey)
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
	return profileimage.Delete(
		ctx,
		u.storage,
		userID,
		p.PersonalInfo.AvatarKey,
		u.profiles.UpdateSeekerAvatarKey,
	)
}
