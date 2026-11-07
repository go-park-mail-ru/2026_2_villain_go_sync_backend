package application_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/application"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestHandler_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := application.NewMockApplicationUsecase(ctrl)
	h := application.NewHandler(mockUC)

	mockUC.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, input dto.CreateApplicationInput) (models.Application, error) {
			if input.SeekerID != 10 || input.VacancyID != 5 || input.ResumeID != 3 {
				t.Errorf("unexpected IDs in input: %+v", input)
			}
			if input.CoverLetter == nil || *input.CoverLetter != "Hello" {
				t.Errorf("expected cover letter 'Hello', got %v", input.CoverLetter)
			}
			return models.Application{ID: 1}, nil
		})

	body := `{"resume_id": 3, "cover_letter": "Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/vacancies/5/applications", bytes.NewBufferString(body))
	req.SetPathValue("id", "5")

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := application.NewMockApplicationUsecase(ctrl)
	h := application.NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/vacancies/5/applications", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandler_Create_InvalidVacancyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := application.NewMockApplicationUsecase(ctrl)
	h := application.NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/vacancies/abc/applications", nil)
	req.SetPathValue("id", "abc")

	ctx := auth.WithUserID(req.Context(), 10)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
