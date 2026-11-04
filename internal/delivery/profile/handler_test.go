package profile_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/mocks"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestGetSeekerByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		GetSeekerByID(gomock.Any(), 1).
		Return(models.SeekerProfile{FirstName: "Ivan"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/seekers/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.GetSeekerByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestGetSeekerByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		GetSeekerByID(gomock.Any(), 1).
		Return(models.SeekerProfile{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/seekers/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.GetSeekerByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateEmployer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		UpdateEmployer(gomock.Any(), 10, models.RoleEmployer, gomock.Any()).
		Return(models.EmployerProfile{CompanyName: "Test"}, nil)

	body := `{"company_name":"Test","description":"Desc","website":"site.com"}`
	req := httptest.NewRequest(http.MethodPut, "/api/profile/employer", bytes.NewBufferString(body))

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleEmployer)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateEmployer(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateEmployer_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		UpdateEmployer(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).
		Return(models.EmployerProfile{}, apperrors.ErrForbidden)

	body := `{"company_name":"Test"}`
	req := httptest.NewRequest(http.MethodPut, "/api/profile/employer", bytes.NewBufferString(body))

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateEmployer(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
