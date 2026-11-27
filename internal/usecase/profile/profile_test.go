package profile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/profile"
	profileUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/profile"
	"go.uber.org/mock/gomock"
)

func TestProfileUsecase_UpdateEmployer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	const userID = 10
	p := models.EmployerProfile{Company: models.EmployerCompanyInfo{CompanyName: "Yandex"}}

	want := p
	want.Profile.UserID = userID

	mockRepo.EXPECT().
		UpsertEmployer(gomock.Any(), want).
		Return(want, nil)

	res, err := uc.UpdateEmployer(context.Background(), userID, models.RoleEmployer, p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Company.CompanyName != "Yandex" {
		t.Errorf("expected Yandex, got %v", res.Company.CompanyName)
	}
}

func TestProfileUsecase_UpdateEmployer_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	_, err := uc.UpdateEmployer(context.Background(), 10, models.RoleSeeker, models.EmployerProfile{})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestProfileUsecase_GetEmployerByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	p := models.EmployerProfile{Company: models.EmployerCompanyInfo{CompanyName: "Yandex"}}

	mockRepo.EXPECT().
		GetEmployerByID(gomock.Any(), 1).
		Return(p, nil)

	res, err := uc.GetEmployerByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Company.CompanyName != "Yandex" {
		t.Errorf("expected Yandex, got %v", res.Company.CompanyName)
	}
}
