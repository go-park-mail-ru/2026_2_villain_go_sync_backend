package resume

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
)

//go:generate mockgen -source=resume.go -destination=mock_repository.go -package=resume
type ResumeRepository interface {
	Create(ctx context.Context, resume models.Resume) (models.Resume, error)
	GetByID(ctx context.Context, id int) (models.Resume, error)
	GetOwnerID(ctx context.Context, id int) (int, error)
	ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error)
	Update(ctx context.Context, input dto.UpdateInput, seekerID int) (models.Resume, error)
	Delete(ctx context.Context, id int, seekerID int) error
	HasApplicantResume(ctx context.Context, resumeID int, employerID int) (bool, error)
}

type ResumeUsecase struct {
	resumes ResumeRepository
}

func NewResumeUsecase(resumes ResumeRepository) *ResumeUsecase {
	return &ResumeUsecase{resumes: resumes}
}

func (u *ResumeUsecase) Create(ctx context.Context, userID int, role models.Role, resume models.Resume) (models.Resume, error) {
	if !role.IsSeeker() {
		return models.Resume{}, apperrors.ErrForbidden
	}
	if err := validatePreferences(resume.Preferences); err != nil {
		return models.Resume{}, err
	}
	resume.SeekerID = userID
	return u.resumes.Create(ctx, resume)
}

func (u *ResumeUsecase) GetByID(ctx context.Context, resumeID int, userID int, role models.Role) (models.Resume, error) {
	res, err := u.resumes.GetByID(ctx, resumeID)
	if err != nil {
		return models.Resume{}, err
	}

	if role.IsSeeker() {
		if int64(res.SeekerID) != int64(userID) {
			return models.Resume{}, apperrors.ErrForbidden
		}
	} else if role.IsEmployer() {
		hasAccess, err := u.resumes.HasApplicantResume(ctx, resumeID, userID)
		if err != nil {
			return models.Resume{}, err
		}
		if !hasAccess {
			return models.Resume{}, apperrors.ErrForbidden
		}
	} else {
		return models.Resume{}, apperrors.ErrForbidden
	}
	return res, nil
}

func (u *ResumeUsecase) ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error) {
	return u.resumes.ListBySeekerID(ctx, seekerID)
}

func (u *ResumeUsecase) Update(ctx context.Context, userID int, role models.Role, input dto.UpdateInput) (models.Resume, error) {
	if !role.IsSeeker() {
		return models.Resume{}, apperrors.ErrForbidden
	}
	ownerID, err := u.resumes.GetOwnerID(ctx, input.Resume.ID)
	if err != nil {
		return models.Resume{}, err
	}
	if ownerID != userID {
		return models.Resume{}, apperrors.ErrNotOwner
	}
	if err := validatePreferences(input.Resume.Preferences); err != nil {
		return models.Resume{}, err
	}

	res, err := u.resumes.Update(ctx, input, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return models.Resume{}, apperrors.ErrNotFound
		}
		return models.Resume{}, err
	}
	return res, nil
}

func (u *ResumeUsecase) Delete(ctx context.Context, userID int, role models.Role, id int) error {
	if !role.IsSeeker() {
		return apperrors.ErrForbidden
	}

	err := u.resumes.Delete(ctx, id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNoContent) {
			return apperrors.ErrNoContent
		}
		return err
	}
	return nil
}

func validatePreferences(preferences models.ResumePreferences) error {
	if value := preferences.EmploymentType; value != nil && *value != "" && !value.IsValid() {
		return apperrors.ValidationError("employment_type must be full or part")
	}
	if value := preferences.WorkFormat; value != nil && *value != "" && !value.IsValid() {
		return apperrors.ValidationError("work_format must be office, hybrid or remote")
	}
	if value := preferences.WorkSchedule; value != nil && *value != "" && !value.IsValid() {
		return apperrors.ValidationError("invalid work_schedule")
	}
	return nil
}
