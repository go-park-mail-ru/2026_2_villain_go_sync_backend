package vacancy_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestMoreVacancyHandler_GetByID(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		err    error
		status int
	}{
		{"found", "7", nil, http.StatusOK},
		{"missing", "7", apperrors.ErrNotFound, http.StatusNotFound},
		{"invalid ID", "abc", nil, http.StatusBadRequest},
		{"zero ID", "0", nil, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			if tt.status != http.StatusBadRequest {
				uc.EXPECT().GetByID(gomock.Any(), int64(7)).
					Return(models.Vacancy{ID: 7, Title: "Go developer"}, tt.err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/vacancies/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()
			h.GetByID(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, rec.Code, rec.Body.String())
			}
			if tt.status == http.StatusOK && !bytes.Contains(rec.Body.Bytes(), []byte(`"title":"Go developer"`)) {
				t.Errorf("title missing from response: %s", rec.Body.String())
			}
		})
	}
}

func TestMoreVacancyHandler_ListByEmployer(t *testing.T) {
	tests := []struct {
		name       string
		authorized bool
		err        error
		status     int
	}{
		{"unauthorized", false, nil, http.StatusUnauthorized},
		{"success", true, nil, http.StatusOK},
		{"forbidden", true, apperrors.ErrForbidden, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			req := httptest.NewRequest(http.MethodGet, "/api/employer/vacancies", nil)
			if tt.authorized {
				req = req.WithContext(auth.WithUserID(req.Context(), 12))
				uc.EXPECT().ListByEmployer(gomock.Any(), int64(12)).
					Return([]models.Vacancy{{ID: 7, Title: "Go developer"}}, tt.err)
			}
			rec := httptest.NewRecorder()
			h.ListVacanciesByEmployer(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMoreVacancyHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		authorized bool
		ucErr      error
		status     int
		call       bool
	}{
		{"invalid JSON", "{oops", true, nil, http.StatusBadRequest, false},
		{"unauthorized", `{"title":"Go dev"}`, false, nil, http.StatusUnauthorized, false},
		{"success", `{"title":"Go dev","description":"desc","currency":"RUB","category_ids":[1]}`, true, nil, http.StatusCreated, true},
		{"forbidden", `{"title":"Go dev"}`, true, apperrors.ErrForbidden, http.StatusForbidden, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			req := httptest.NewRequest(http.MethodPost, "/api/vacancies", bytes.NewBufferString(tt.body))
			if tt.authorized {
				req = req.WithContext(auth.WithUserID(req.Context(), 12))
			}
			if tt.call {
				uc.EXPECT().Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, input dto.CreateVacancyInput) (models.Vacancy, error) {
						if input.EmployerID != 12 {
							t.Errorf("wrong employer id: %d", input.EmployerID)
						}
						return models.Vacancy{ID: 7, Title: input.Title}, tt.ucErr
					})
			}
			rec := httptest.NewRecorder()
			h.CreateVacancy(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMoreVacancyHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           string
		authorized     bool
		wantStatus     int
		call           bool
		ucErr          error
		salaryProvided bool
		salaryNil      bool
	}{
		{name: "unauthorized", id: "7", body: `{}`, wantStatus: http.StatusUnauthorized},
		{name: "invalid JSON", id: "7", body: "{oops", authorized: true, wantStatus: http.StatusBadRequest},
		{name: "invalid ID", id: "abc", body: `{}`, authorized: true, wantStatus: http.StatusBadRequest},
		{name: "invalid salary", id: "7", body: `{"salary_from":"bad"}`, authorized: true, wantStatus: http.StatusBadRequest},
		{name: "success without salary update", id: "7", body: `{"title":"Updated"}`, authorized: true, wantStatus: http.StatusOK, call: true},
		{name: "salary set", id: "7", body: `{"salary_from":35000}`, authorized: true, wantStatus: http.StatusOK, call: true, salaryProvided: true},
		{name: "salary cleared", id: "7", body: `{"salary_from":null}`, authorized: true, wantStatus: http.StatusOK, call: true, salaryProvided: true, salaryNil: true},
		{name: "forbidden", id: "7", body: `{}`, authorized: true, wantStatus: http.StatusForbidden, call: true, ucErr: apperrors.ErrForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			req := httptest.NewRequest(http.MethodPut, "/api/vacancies/"+tt.id, bytes.NewBufferString(tt.body))
			req.SetPathValue("id", tt.id)
			if tt.authorized {
				req = req.WithContext(auth.WithUserID(req.Context(), 12))
			}
			if tt.call {
				uc.EXPECT().Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, input dto.UpdateVacancyInput) (models.Vacancy, error) {
						if input.VacancyID != 7 || input.EmployerID != 12 {
							t.Errorf("wrong IDs: %+v", input)
						}
						if input.SalaryFrom.Provided != tt.salaryProvided {
							t.Errorf("wrong salary provided flag: %+v", input.SalaryFrom)
						}
						if tt.salaryProvided {
							if tt.salaryNil && input.SalaryFrom.Value != nil {
								t.Errorf("expected nil salary")
							}
							if !tt.salaryNil && (input.SalaryFrom.Value == nil || *input.SalaryFrom.Value != 35000) {
								t.Errorf("expected salary 35000")
							}
						}
						return models.Vacancy{ID: 7, Title: "Updated"}, tt.ucErr
					})
			}
			rec := httptest.NewRecorder()
			h.UpdateVacancy(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d: %s", tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMoreVacancyHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		authorized bool
		id         string
		err        error
		status     int
		call       bool
	}{
		{"unauthorized", false, "7", nil, http.StatusUnauthorized, false},
		{"invalid ID", true, "abc", nil, http.StatusBadRequest, false},
		{"zero ID", true, "0", nil, http.StatusBadRequest, false},
		{"success", true, "7", nil, http.StatusNoContent, true},
		{"forbidden", true, "7", apperrors.ErrForbidden, http.StatusForbidden, true},
		{"not found", true, "7", apperrors.ErrNotFound, http.StatusNotFound, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			req := httptest.NewRequest(http.MethodDelete, "/api/vacancies/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			if tt.authorized {
				req = req.WithContext(auth.WithUserID(req.Context(), 12))
			}
			if tt.call {
				uc.EXPECT().Delete(gomock.Any(), int64(7), int64(12)).Return(tt.err)
			}
			rec := httptest.NewRecorder()
			h.Delete(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMoreVacancyHandler_ListEmployerApplicants(t *testing.T) {
	tests := []struct {
		name     string
		withUser bool
		withRole bool
		role     models.Role
		err      error
		status   int
		call     bool
	}{
		{name: "no user", status: http.StatusUnauthorized},
		{name: "no role", withUser: true, status: http.StatusUnauthorized},
		{name: "success", withUser: true, withRole: true, role: models.RoleEmployer, status: http.StatusOK, call: true},
		{name: "forbidden", withUser: true, withRole: true, role: models.RoleSeeker, err: apperrors.ErrForbidden, status: http.StatusForbidden, call: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := vacancy.NewMockVacancyUsecase(ctrl)
			h := vacancy.NewHandler(uc)
			req := httptest.NewRequest(http.MethodGet, "/api/employer/applications", nil)
			if tt.withUser {
				ctx := auth.WithUserID(req.Context(), 12)
				if tt.withRole {
					ctx = auth.WithRole(ctx, tt.role)
				}
				req = req.WithContext(ctx)
			}
			if tt.call {
				uc.EXPECT().ListEmployerApplicants(gomock.Any(), int64(12), tt.role).
					Return([]models.Applicant{{ApplicationID: 4, FirstName: "Ivan"}}, tt.err)
			}
			rec := httptest.NewRecorder()
			h.ListEmployerApplicants(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("want %d, got %d: %s", tt.status, rec.Code, rec.Body.String())
			}
			if tt.status == http.StatusOK && !bytes.Contains(rec.Body.Bytes(), []byte(`"application_id":4`)) {
				t.Errorf("application missing from response: %s", rec.Body.String())
			}
		})
	}
}
