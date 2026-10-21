package resume

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	repo "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/repository/resume"
)

type ResumeRepository interface {
	Create(ctx context.Context, resume models.Resume) (models.Resume, error)
	GetByID(ctx context.Context, id int) (models.Resume, error)
	ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error)
	Update(ctx context.Context, resume models.Resume, seekerID int) (models.Resume, error)
	Delete(ctx context.Context, id int, seekerID int) error
}

type ResumeUsecase struct {
	resumes ResumeRepository
}

func NewResumeUsecase(resumes ResumeRepository) *ResumeUsecase {
	return &ResumeUsecase{resumes: resumes}
}

func (u *ResumeUsecase) Create(ctx context.Context, userID int, role string, resume models.Resume) (models.Resume, error) {
	if role != "seeker" {
		return models.Resume{}, ErrForbidden
	}
	resume.SeekerID = userID
	return u.resumes.Create(ctx, resume)
}

func (u *ResumeUsecase) GetByID(ctx context.Context, id int) (models.Resume, error) {
	res, err := u.resumes.GetByID(ctx, id)
	if err != nil {
		if err == repo.ErrNotFound {
			return models.Resume{}, ErrNotFound
		}
		return models.Resume{}, err
	}
	return res, nil
}

func (u *ResumeUsecase) ListBySeekerID(ctx context.Context, seekerID int) ([]models.Resume, error) {
	return u.resumes.ListBySeekerID(ctx, seekerID)
}

func (u *ResumeUsecase) Update(ctx context.Context, userID int, role string, resume models.Resume) (models.Resume, error) {
	if role != "seeker" {
		return models.Resume{}, ErrForbidden
	}

	res, err := u.resumes.Update(ctx, resume, userID)
	if err != nil {
		if err == repo.ErrNotFound {
			return models.Resume{}, ErrNotFound
		}
		return models.Resume{}, err
	}
	return res, nil
}

func (u *ResumeUsecase) Delete(ctx context.Context, userID int, role string, id int) error {
	if role != "seeker" {
		return ErrForbidden
	}

	err := u.resumes.Delete(ctx, id, userID)
	if err != nil {
		if err == repo.ErrNotFound {
			return ErrNotFound
		}
		return err
	}
	return nil
}
