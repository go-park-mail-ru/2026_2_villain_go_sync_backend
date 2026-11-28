package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	uc "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/runtimeconfig"
	"go.uber.org/mock/gomock"
)

func TestApplicationUsecase_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := uc.NewMockApplicationRepository(ctrl)
	mockRes := uc.NewMockResumeRepository(ctrl)
	mockVac := uc.NewMockVacancyRepository(ctrl)
	mockUser := uc.NewMockUserRepository(ctrl)

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser, runtimeconfig.New(nil))

	input := dto.CreateApplicationInput{
		SeekerID:  10,
		VacancyID: 5,
		ResumeID:  3,
	}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return(models.RoleSeeker, nil)

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

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser, runtimeconfig.New(nil))

	input := dto.CreateApplicationInput{SeekerID: 10, VacancyID: -5, ResumeID: 3}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return(models.RoleEmployer, nil)

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

	usecase := uc.NewApplicationUsecase(mockApp, mockRes, mockVac, mockUser, runtimeconfig.New(nil))

	input := dto.CreateApplicationInput{SeekerID: 10, VacancyID: 5, ResumeID: 3}

	mockUser.EXPECT().GetRoleByID(gomock.Any(), int64(10)).Return(models.RoleSeeker, nil)

	mockRes.EXPECT().GetByID(gomock.Any(), 3).Return(models.Resume{ID: 3, SeekerID: 99}, nil)

	_, err := usecase.Create(context.Background(), input)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateStatusValidation(t *testing.T) {
	reason := models.RejectionSkillsMismatch
	unknown := models.RejectionReasonCode("unknown")
	for _, tc := range []struct {
		name   string
		role   models.Role
		status models.ApplicationStatusValue
		reason *models.RejectionReasonCode
		want   error
	}{
		{name: "seeker cannot change status", role: models.RoleSeeker, status: "invalid", reason: &unknown, want: apperrors.ErrForbidden},
		{name: "rejection needs reason", role: models.RoleEmployer, status: models.ApplicationRejected, want: apperrors.ErrInvalidInput},
		{name: "unknown reason", role: models.RoleEmployer, status: models.ApplicationRejected, reason: &unknown, want: apperrors.ErrInvalidInput},
		{name: "reason on offer", role: models.RoleEmployer, status: models.ApplicationOffer, reason: &reason, want: apperrors.ErrInvalidInput},
		{name: "viewed is forbidden", role: models.RoleEmployer, status: "viewed", want: apperrors.ErrInvalidInput},
		{name: "accepted is forbidden", role: models.RoleEmployer, status: "accepted", want: apperrors.ErrInvalidInput},
		{name: "created is automatic", role: models.RoleEmployer, status: models.ApplicationCreated, want: apperrors.ErrInvalidInput},
		{name: "valid reviewing", role: models.RoleEmployer, status: models.ApplicationReviewing},
		{name: "valid interview", role: models.RoleEmployer, status: models.ApplicationInterview},
		{name: "valid offer", role: models.RoleEmployer, status: models.ApplicationOffer},
		{name: "valid rejection", role: models.RoleEmployer, status: models.ApplicationRejected, reason: &reason},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := uc.NewMockApplicationRepository(gomock.NewController(t))
			usecase := uc.NewApplicationUsecase(repo, nil, nil, nil, runtimeconfig.New(nil))
			input := dto.UpdateStatusInput{ApplicationID: 1, EmployerID: 2, Status: tc.status, ReasonCode: tc.reason}
			if tc.role.IsEmployer() {
				repo.EXPECT().GetVacancyEmployerID(gomock.Any(), int64(1)).Return(int64(2), nil)
			}
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
	usecase := uc.NewApplicationUsecase(repo, nil, nil, nil, runtimeconfig.New(nil))
	if _, _, err := usecase.ListBySeeker(context.Background(), 1, models.RoleEmployer, pagination.Params{Limit: 12}); !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestUpdateStatusChecksOwnershipBeforeStatus(t *testing.T) {
	for _, tc := range []struct {
		name            string
		ownerID         int64
		lookupErr, want error
	}{
		{"another employer", 3, nil, apperrors.ErrForbidden},
		{"missing application", 0, apperrors.ErrNotFound, apperrors.ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := uc.NewMockApplicationRepository(gomock.NewController(t))
			repo.EXPECT().GetVacancyEmployerID(gomock.Any(), int64(1)).Return(tc.ownerID, tc.lookupErr)
			input := dto.UpdateStatusInput{ApplicationID: 1, EmployerID: 2, Status: "invalid"}
			_, err := uc.NewApplicationUsecase(repo, nil, nil, nil, runtimeconfig.New(nil)).UpdateStatus(context.Background(), models.RoleEmployer, input)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}
