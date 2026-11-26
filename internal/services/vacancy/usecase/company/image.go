package company

import (
	"context"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/shared/profileimage"
)

func (u *CompanyUsecase) UploadEmployerLogo(ctx context.Context, userID int, role models.Role, body io.ReadSeeker, contentType string) error {
	if !role.IsEmployer() {
		return apperrors.ErrForbidden
	}
	p, err := u.companies.GetEmployerByID(ctx, userID)
	if err != nil {
		return err
	}
	return profileimage.Replace(
		ctx,
		u.storage,
		userID,
		"logos",
		p.Company.LogoKey,
		body,
		contentType,
		u.companies.UpdateEmployerLogoKey,
	)
}

func (u *CompanyUsecase) GetEmployerLogo(ctx context.Context, userID int) (io.ReadCloser, string, error) {
	p, err := u.companies.GetEmployerByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if p.Company.LogoKey == nil {
		return nil, "", apperrors.ErrNotFound
	}
	body, contentType, err := u.storage.Download(ctx, *p.Company.LogoKey)
	if err != nil {
		return nil, "", fmt.Errorf("get logo from s3 storage: %w", err)
	}
	return body, contentType, nil
}

func (u *CompanyUsecase) DeleteEmployerLogo(ctx context.Context, userID int, role models.Role) error {
	if !role.IsEmployer() {
		return apperrors.ErrForbidden
	}
	p, err := u.companies.GetEmployerByID(ctx, userID)
	if err != nil {
		return err
	}
	return profileimage.Delete(
		ctx,
		u.storage,
		userID,
		p.Company.LogoKey,
		u.companies.UpdateEmployerLogoKey,
	)
}
