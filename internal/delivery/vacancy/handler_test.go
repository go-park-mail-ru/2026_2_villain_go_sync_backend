package vacancy_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	vacancydto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestListVacancies_Success(t *testing.T) {
	stamp := time.Date(2026, 11, 21, 10, 0, 0, 123456000, time.UTC)
	items := []models.Vacancy{{ID: 1, CreatedAt: stamp}, {ID: 2, CreatedAt: stamp}}
	position := func(id int64) *pagination.Cursor {
		return &pagination.Cursor{CreatedAt: stamp, ID: id}
	}
	for _, tc := range []struct {
		name   string
		query  string
		cursor *pagination.Cursor
		limit  int
		items  []models.Vacancy
		ids    []int64
		next   *pagination.Cursor
	}{
		{name: "defaults", limit: 12, items: items, ids: []int64{1, 2}},
		{name: "first batch", query: "?limit=1", limit: 1, items: items, ids: []int64{1}, next: position(1)},
		{name: "last batch", query: "?limit=1", cursor: position(1), limit: 1, items: items[1:], ids: []int64{2}},
		{name: "empty batch", query: "?limit=1", cursor: position(2), limit: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mockUC := vacancy.NewMockVacancyUsecase(gomock.NewController(t))
			mockUC.EXPECT().ListVacancies(gomock.Any(), pagination.Params{Cursor: tc.cursor, Limit: tc.limit + 1}).Return(tc.items, int64(2), nil)
			url := "/api/vacancies" + tc.query
			if tc.cursor != nil {
				raw, err := json.Marshal(tc.cursor)
				if err != nil {
					t.Fatal(err)
				}
				url += "&cursor=" + base64.RawURLEncoding.EncodeToString(raw)
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
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
			actual := response.Data.Pagination
			want := pagination.Metadata{Limit: tc.limit, Total: 2}
			if tc.next != nil {
				if actual.NextCursor == nil {
					t.Fatal("missing next_cursor")
				}
				raw, err := base64.RawURLEncoding.DecodeString(*actual.NextCursor)
				if err != nil {
					t.Fatal(err)
				}
				var next pagination.Cursor
				if err := json.Unmarshal(raw, &next); err != nil || next != *tc.next {
					t.Fatalf("cursor = %+v, want %+v, err = %v", next, *tc.next, err)
				}
				want.NextCursor = actual.NextCursor
			}
			if actual != want {
				t.Errorf("pagination = %+v, want %+v", actual, want)
			}
			got := response.Data.Items
			if got == nil || len(got) != len(tc.ids) {
				t.Fatalf("unexpected items: %+v", got)
			}
			for i, item := range got {
				if item.ID != tc.ids[i] {
					t.Errorf("item %d ID = %d, want %d", i, item.ID, tc.ids[i])
				}
			}
		})
	}
}

func TestListVacancies_InvalidPagination(t *testing.T) {
	for _, query := range []string{"page=2", "cursor=", "cursor=%25", "cursor=e30", "cursor=e30&cursor=e30", "limit=", "limit=0", "limit=101"} {
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

	mockUC.EXPECT().ListVacancies(gomock.Any(), pagination.Params{Limit: 13}).Return(nil, int64(0), errors.New("db error"))

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
		ListApplicants(gomock.Any(), int64(10), models.RoleEmployer, int64(1), pagination.Params{Limit: 6}).
		Return([]models.Applicant{{Seeker: models.ApplicantSeeker{FirstName: "Ivan"}}}, int64(6), nil)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/api/vacancies/1/applicants?limit=5", nil)
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
