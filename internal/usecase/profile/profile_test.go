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

func TestProfileUsecase_UpdateSeeker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	p := models.SeekerProfile{FirstName: "Ivan"}

	mockRepo.EXPECT().
		UpsertSeeker(gomock.Any(), p).
		Return(p, nil)

	res, err := uc.UpdateSeeker(context.Background(), 10, models.RoleSeeker, p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FirstName != "Ivan" {
		t.Errorf("expected Ivan, got %v", res.FirstName)
	}
}

func TestProfileUsecase_UpdateSeeker_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	_, err := uc.UpdateSeeker(context.Background(), 10, models.RoleEmployer, models.SeekerProfile{})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestProfileUsecase_GetSeekerByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	p := models.SeekerProfile{FirstName: "Ivan"}

	mockRepo.EXPECT().
		GetSeekerByID(gomock.Any(), 1).
		Return(p, nil)

	res, err := uc.GetSeekerByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FirstName != "Ivan" {
		t.Errorf("expected Ivan, got %v", res.FirstName)
	}
}

func TestProfileUsecase_GetSeekerByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	mockRepo.EXPECT().
		GetSeekerByID(gomock.Any(), 1).
		Return(models.SeekerProfile{}, apperrors.ErrNotFound)

	_, err := uc.GetSeekerByID(context.Background(), 1)

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProfileUsecase_UpdateEmployer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := profile.NewMockProfileRepository(ctrl)
	uc := profileUsecase.NewProfileUsecase(mockRepo, nil)

	p := models.EmployerProfile{CompanyName: "Yandex"}

	mockRepo.EXPECT().
		UpsertEmployer(gomock.Any(), p).
		Return(p, nil)

	res, err := uc.UpdateEmployer(context.Background(), 10, models.RoleEmployer, p)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CompanyName != "Yandex" {
		t.Errorf("expected Yandex, got %v", res.CompanyName)
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

	p := models.EmployerProfile{CompanyName: "Yandex"}

	mockRepo.EXPECT().
		GetEmployerByID(gomock.Any(), 1).
		Return(p, nil)

	res, err := uc.GetEmployerByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CompanyName != "Yandex" {
		t.Errorf("expected Yandex, got %v", res.CompanyName)
	}
}
