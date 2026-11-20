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

func TestUpdateStatusValidation(t *testing.T) {
	reason := "skills_mismatch"
	unknown := "unknown"
	for _, tc := range []struct {
		name   string
		role   models.Role
		status string
		reason *string
		want   error
	}{
		{name: "seeker cannot change status", role: models.RoleSeeker, status: "rejected", reason: &reason, want: apperrors.ErrForbidden},
		{name: "rejection needs reason", role: models.RoleEmployer, status: "rejected", want: apperrors.ErrInvalidInput},
		{name: "unknown reason", role: models.RoleEmployer, status: "rejected", reason: &unknown, want: apperrors.ErrInvalidInput},
		{name: "reason on offer", role: models.RoleEmployer, status: "offer", reason: &reason, want: apperrors.ErrInvalidInput},
		{name: "viewed is forbidden", role: models.RoleEmployer, status: "viewed", want: apperrors.ErrInvalidInput},
		{name: "accepted is forbidden", role: models.RoleEmployer, status: "accepted", want: apperrors.ErrInvalidInput},
		{name: "created is automatic", role: models.RoleEmployer, status: "created", want: apperrors.ErrInvalidInput},
		{name: "valid reviewing", role: models.RoleEmployer, status: "reviewing"},
		{name: "valid interview", role: models.RoleEmployer, status: "interview"},
		{name: "valid offer", role: models.RoleEmployer, status: "offer"},
		{name: "valid rejection", role: models.RoleEmployer, status: "rejected", reason: &reason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := uc.NewMockApplicationRepository(gomock.NewController(t))
			usecase := uc.NewApplicationUsecase(repo, nil, nil, nil)
			input := dto.UpdateStatusInput{ApplicationID: 1, EmployerID: 2, Status: tc.status, ReasonCode: tc.reason}
			if tc.want == nil {
				repo.EXPECT().UpdateStatus(gomock.Any(), input).Return(models.ApplicationStatusUpdate{ApplicationID: 1}, nil)
			}
			_, err := usecase.UpdateStatus(context.Background(), tc.role, input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestListBySeekerRequiresSeekerRole(t *testing.T) {
	repo := uc.NewMockApplicationRepository(gomock.NewController(t))
	usecase := uc.NewApplicationUsecase(repo, nil, nil, nil)
	if _, _, err := usecase.ListBySeeker(context.Background(), 1, models.RoleEmployer, 1, 12); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
