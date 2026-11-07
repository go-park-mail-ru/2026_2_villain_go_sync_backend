package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	uc "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/application"
	"go.uber.org/mock/gomock"
)

func TestApplicationUsecase_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := uc.NewMockApplicationRepository(ctrl)
	mockRes := uc.NewMockResumeRepository(ctrl)
	mockVac := uc.NewMockVacancyRepository(ctrl)
	mockUser := uc.NewMockUserRepository(ctrl)

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser)

	input := dto.CreateApplicationInput{
		SeekerID:  10,
		VacancyID: 5,
		ResumeID:  3,
	}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return("seeker", nil)

	mockRes.EXPECT().GetByID(gomock.Any(), 3).Return(models.Resume{ID: 3, SeekerID: 10}, nil)

	mockVac.EXPECT().GetByID(gomock.Any(), int64(5)).Return(models.Vacancy{ID: 5, Version: 1}, nil)

	mockApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(models.Application{ID: 100}, nil)

	res, err := usecase.Create(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 100 {
		t.Errorf("expected ID 100, got %d", res.ID)
	}
}

func TestApplicationUsecase_Create_NotSeeker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := uc.NewMockApplicationRepository(ctrl)
	mockRes := uc.NewMockResumeRepository(ctrl)
	mockVac := uc.NewMockVacancyRepository(ctrl)
	mockUser := uc.NewMockUserRepository(ctrl)

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser)

	input := dto.CreateApplicationInput{SeekerID: 10, VacancyID: 5, ResumeID: 3}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return("employer", nil)

	_, err := usecase.Create(context.Background(), input)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestApplicationUsecase_Create_ResumeBelongsToAnother(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := uc.NewMockApplicationRepository(ctrl)
	mockRes := uc.NewMockResumeRepository(ctrl)
	mockVac := uc.NewMockVacancyRepository(ctrl)
	mockUser := uc.NewMockUserRepository(ctrl)

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser)

	input := dto.CreateApplicationInput{SeekerID: 10, VacancyID: 5, ResumeID: 3}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return("seeker", nil)

	mockRes.EXPECT().GetByID(gomock.Any(), 3).Return(models.Resume{ID: 3, SeekerID: 99}, nil)

	_, err := usecase.Create(context.Background(), input)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
