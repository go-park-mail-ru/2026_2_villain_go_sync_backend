package vacancy_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	vacancydto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestListVacancies_Success(t *testing.T) {
	for _, tc := range []struct {
		name, query       string
		page, limit       int
		items             []models.Vacancy
		total, totalPages int64
	}{
		{name: "defaults", page: 1, limit: 12, items: []models.Vacancy{{ID: 1, Title: "Go"}}, total: 13, totalPages: 2},
		{name: "explicit page", query: "?page=2&limit=5", page: 2, limit: 5, items: []models.Vacancy{{ID: 6, Title: "Go"}}, total: 6, totalPages: 2},
		{name: "empty page", query: "?page=3&limit=5", page: 3, limit: 5, total: 6, totalPages: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := vacancy.NewMockVacancyUsecase(gomock.NewController(t))
			mockUC.EXPECT().ListVacancies(gomock.Any(), tc.page, tc.limit).Return(tc.items, tc.total, nil)
			req := httptest.NewRequest(http.MethodGet, "/api/vacancies"+tc.query, nil)
			rec := httptest.NewRecorder()
			vacancy.NewHandler(mockUC).ListVacancies(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			var response struct {
				Data pagination.Response[vacancydto.VacancyResponse] `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			want := pagination.Metadata{
				Page:       tc.page,
				Limit:      tc.limit,
				Total:      tc.total,
				TotalPages: tc.totalPages,
			}
			if response.Data.Pagination != want {
				t.Errorf("pagination = %+v, want %+v", response.Data.Pagination, want)
			}
			items := response.Data.Items
			if items == nil || len(items) != len(tc.items) {
				t.Fatalf("unexpected items: %+v", items)
			}
			for i, item := range items {
				if item.ID != tc.items[i].ID {
					t.Errorf("item %d ID = %d, want %d", i, item.ID, tc.items[i].ID)
				}
			}
		})
	}
}

func TestListVacancies_InvalidPagination(t *testing.T) {
	for _, query := range []string{"page=0", "page=x", "limit=", "limit=0", "limit=101"} {
		t.Run(query, func(t *testing.T) {
			mockUC := vacancy.NewMockVacancyUsecase(gomock.NewController(t))
			req := httptest.NewRequest(http.MethodGet, "/api/vacancies?"+query, nil)
			rec := httptest.NewRecorder()
			vacancy.NewHandler(mockUC).ListVacancies(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListVacancies_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := vacancy.NewMockVacancyUsecase(ctrl)
	h := vacancy.NewHandler(mockUC)

	mockUC.EXPECT().ListVacancies(gomock.Any(), 1, 12).Return(nil, int64(0), errors.New("db error"))

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

	mockUC := vacancy.NewMockVacancyUsecase(ctrl)
	h := vacancy.NewHandler(mockUC)

	// Настраиваем ожидание мока
	mockUC.EXPECT().
		ListApplicants(gomock.Any(), int64(10), models.RoleEmployer, int64(1), 2, 5).
		Return([]models.Applicant{{Seeker: models.ApplicantSeeker{FirstName: "Ivan"}}}, int64(6), nil)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/api/vacancies/1/applicants?page=2&limit=5", nil)
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
