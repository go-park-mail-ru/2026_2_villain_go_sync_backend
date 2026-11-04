package resume_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	mocks "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/mocks/usecase_resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	resumeUsecase "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/resume"
	"go.uber.org/mock/gomock"
)

func TestResumeUsecase_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	input := models.Resume{Title: "Go dev"}
	expectedInRepo := models.Resume{SeekerID: 10, Title: "Go dev"}
	returnedFromRepo := models.Resume{ID: 1, SeekerID: 10, Title: "Go dev"}

	mockRepo.EXPECT().
		Create(gomock.Any(), expectedInRepo).
		Return(returnedFromRepo, nil)

	res, err := uc.Create(context.Background(), 10, models.RoleSeeker, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Errorf("expected ID 1, got %d", res.ID)
	}
}

func TestResumeUsecase_Create_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	_, err := uc.Create(context.Background(), 10, models.RoleEmployer, models.Resume{Title: "Go dev"})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestResumeUsecase_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	expected := models.Resume{ID: 1, Title: "Go dev"}

	mockRepo.EXPECT().
		GetByID(gomock.Any(), 1).
		Return(expected, nil)

	res, err := uc.GetByID(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Go dev" {
		t.Errorf("expected 'Go dev', got %v", res.Title)
	}
}

func TestResumeUsecase_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), 1).
		Return(models.Resume{}, apperrors.ErrNotFound)

	_, err := uc.GetByID(context.Background(), 1)

	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestResumeUsecase_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	input := models.Resume{ID: 1, Title: "Updated"}
	expected := models.Resume{ID: 1, SeekerID: 10, Title: "Updated"}

	mockRepo.EXPECT().
		Update(gomock.Any(), input, 10).
		Return(expected, nil)

	res, err := uc.Update(context.Background(), 10, models.RoleSeeker, input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Title != "Updated" {
		t.Errorf("expected 'Updated', got %v", res.Title)
	}
}

func TestResumeUsecase_Update_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	_, err := uc.Update(context.Background(), 10, models.RoleEmployer, models.Resume{ID: 1})

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestResumeUsecase_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	mockRepo.EXPECT().
		Delete(gomock.Any(), 1, 10).
		Return(nil)

	err := uc.Delete(context.Background(), 10, models.RoleSeeker, 1)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestResumeUsecase_Delete_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	err := uc.Delete(context.Background(), 10, models.RoleEmployer, 1)

	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestResumeUsecase_Delete_NoContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockResumeRepository(ctrl)
	uc := resumeUsecase.NewResumeUsecase(mockRepo)

	mockRepo.EXPECT().
		Delete(gomock.Any(), 1, 10).
		Return(apperrors.ErrNoContent)

	err := uc.Delete(context.Background(), 10, models.RoleSeeker, 1)

	if !errors.Is(err, apperrors.ErrNoContent) {
		t.Errorf("expected ErrNoContent, got %v", err)
	}
}
