package vacancy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	mocks "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/mocks/usecase_vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	vacancyUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
	"go.uber.org/mock/gomock"
)

func setupUsecase(t *testing.T) (*vacancyUsecase.VacancyUsecase, *mocks.MockVacancyRepository, *mocks.MockUserRepository) {
	ctrl := gomock.NewController(t)

	mockVacancyRepo := mocks.NewMockVacancyRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := vacancyUsecase.NewVacancyUsecase(mockVacancyRepo, mockUserRepo)
	return uc, mockVacancyRepo, mockUserRepo
}

func TestVacancyUsecase_ListApplicants_NotEmployer(t *testing.T) {
	uc, _, _ := setupUsecase(t)

	_, err := uc.ListApplicants(context.Background(), 10, models.RoleSeeker, 1)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestVacancyUsecase_ListApplicants_NotOwner(t *testing.T) {
	uc, mockVacancyRepo, _ := setupUsecase(t)

	mockVacancyRepo.EXPECT().
		IsOwner(gomock.Any(), int64(1), int64(10)).
		Return(false, nil)

	_, err := uc.ListApplicants(context.Background(), 10, models.RoleEmployer, 1)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestVacancyUsecase_ListApplicants_Success(t *testing.T) {
	uc, mockVacancyRepo, _ := setupUsecase(t)

	mockVacancyRepo.EXPECT().
		IsOwner(gomock.Any(), int64(1), int64(10)).
		Return(true, nil)
	mockVacancyRepo.EXPECT().
		ListByVacancyID(gomock.Any(), int64(1)).
		Return([]models.Applicant{{FirstName: "Ivan"}}, nil)

	res, err := uc.ListApplicants(context.Background(), 10, models.RoleEmployer, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Errorf("expected 1 applicant, got %d", len(res))
	}
}

func TestVacancyUsecase_Create_NotEmployer(t *testing.T) {
	uc, _, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("seeker", nil)

	_, err := uc.Create(context.Background(), dto.CreateVacancyInput{EmployerID: 10})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestVacancyUsecase_Create_NoProfile(t *testing.T) {
	uc, _, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("employer", nil)

	mockUserRepo.EXPECT().
		HasEmployerProfile(gomock.Any(), int64(10)).
		Return(false, nil)

	_, err := uc.Create(context.Background(), dto.CreateVacancyInput{EmployerID: 10})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestVacancyUsecase_Create_InvalidTitle(t *testing.T) {
	uc, _, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("employer", nil)
	mockUserRepo.EXPECT().
		HasEmployerProfile(gomock.Any(), int64(10)).
		Return(true, nil)

	_, err := uc.Create(context.Background(), dto.CreateVacancyInput{
		EmployerID: 10,
		Title:      "   ",
	})

	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestVacancyUsecase_Create_Success(t *testing.T) {
	uc, mockVacancyRepo, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("employer", nil)
	mockUserRepo.EXPECT().
		HasEmployerProfile(gomock.Any(), int64(10)).
		Return(true, nil)

	input := dto.CreateVacancyInput{
		EmployerID: 10,
		Title:      "Go dev",
		Currency:   "RUB",
	}

	mockVacancyRepo.EXPECT().
		Create(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(models.Vacancy{ID: 1, Title: "Go dev"}, nil)

	res, err := uc.Create(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Errorf("expected ID 1, got %d", res.ID)
	}
}

func TestVacancyUsecase_Update_WrongOwner(t *testing.T) {
	uc, mockVacancyRepo, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("employer", nil)
	mockUserRepo.EXPECT().
		HasEmployerProfile(gomock.Any(), int64(10)).
		Return(true, nil)

	mockVacancyRepo.EXPECT().
		GetByID(gomock.Any(), int64(1)).
		Return(models.Vacancy{ID: 1, EmployerID: 99}, nil)

	_, err := uc.Update(context.Background(), dto.UpdateVacancyInput{
		VacancyID:  1,
		EmployerID: 10,
	})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestVacancyUsecase_Delete_Success(t *testing.T) {
	uc, mockVacancyRepo, mockUserRepo := setupUsecase(t)

	mockUserRepo.EXPECT().
		GetRoleByID(gomock.Any(), int64(10)).
		Return("employer", nil)
	mockUserRepo.EXPECT().
		HasEmployerProfile(gomock.Any(), int64(10)).
		Return(true, nil)

	mockVacancyRepo.EXPECT().
		Delete(gomock.Any(), int64(1), int64(10)).
		Return(nil)

	err := uc.Delete(context.Background(), 1, 10)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
