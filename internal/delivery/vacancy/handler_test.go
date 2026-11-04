package vacancy_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	mocks "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/mocks/delivery_vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestListVacancies_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockVacancyUsecase(ctrl)
	h := vacancy.NewHandler(mockUC)

	want := []models.Vacancy{{ID: 1, Title: "Go"}}
	mockUC.EXPECT().ListVacancies(gomock.Any()).Return(want, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/vacancies", nil)
	rec := httptest.NewRecorder()
	h.ListVacancies(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListVacancies_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockVacancyUsecase(ctrl)
	h := vacancy.NewHandler(mockUC)

	mockUC.EXPECT().ListVacancies(gomock.Any()).Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/vacancies", nil)
	rec := httptest.NewRecorder()
	h.ListVacancies(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestListApplicants_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockVacancyUsecase(ctrl)
	h := vacancy.NewHandler(mockUC)

	// Настраиваем ожидание мока
	mockUC.EXPECT().
		ListApplicants(gomock.Any(), int64(10), models.RoleEmployer, int64(1)).
		Return([]models.Applicant{{FirstName: "Ivan"}}, nil)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/api/vacancies/1/applicants", nil)
	req.SetPathValue("id", "1")

	// Имитируем мидлварь
	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleEmployer)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.ListApplicants(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
