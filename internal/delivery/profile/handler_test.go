package profile_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestGetSeekerByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := profile.NewMockProfileUsecase(ctrl)
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

	mockUC := profile.NewMockProfileUsecase(ctrl)
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

	mockUC := profile.NewMockProfileUsecase(ctrl)
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

	mockUC := profile.NewMockProfileUsecase(ctrl)
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

func TestUpdateSeeker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	avatarKey := "avatars/10.jpg"

	mockUC.EXPECT().
		UpdateSeeker(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).
		Return(models.SeekerProfile{
			Profile:   models.Profile{UserID: 10},
			FirstName: "Ivan",
			LastName:  "Ivanov",
			AvatarKey: &avatarKey,
		}, nil)

	body := `{"first_name":"Ivan","last_name":"Ivanov","description":"Go developer"}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/profile/seeker",
		bytes.NewBufferString(body),
	)

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateSeeker(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"avatar_url":"/api/seekers/10/avatar"`)) {
		t.Errorf("unexpected avatar URL: %s", rec.Body.String())
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"first_name":"Ivan"`)) {
		t.Errorf("missing first name: %s", rec.Body.String())
	}
}

func TestUpdateSeeker_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/profile/seeker",
		nil,
	)
	rec := httptest.NewRecorder()

	h.UpdateSeeker(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUpdateSeeker_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/profile/seeker",
		bytes.NewBufferString(`{invalid`),
	)

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateSeeker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateSeeker_InvalidProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	body := `{"last_name":"Ivanov"}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/profile/seeker",
		bytes.NewBufferString(body),
	)

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateSeeker(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rec.Code)
	}
}

func TestUpdateSeeker_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		UpdateSeeker(gomock.Any(), 10, models.RoleEmployer, gomock.Any()).
		Return(models.SeekerProfile{}, apperrors.ErrForbidden)

	body := `{"first_name":"Ivan"}`

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/profile/seeker",
		bytes.NewBufferString(body),
	)

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleEmployer)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.UpdateSeeker(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestGetPublicEmployer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	logoKey := "logos/11.jpg"

	mockUC.EXPECT().
		GetEmployerByID(gomock.Any(), 11).
		Return(models.EmployerProfile{
			Profile:     models.Profile{UserID: 11},
			CompanyName: "Test Company",
			LogoKey:     &logoKey,
		}, nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/employers/11",
		nil,
	)
	req.SetPathValue("id", "11")

	rec := httptest.NewRecorder()
	h.GetPublicEmployer(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"logo_url":"/api/employers/11/logo"`)) {
		t.Errorf("unexpected logo URL: %s", rec.Body.String())
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"company_name":"Test Company"`)) {
		t.Errorf("missing company name: %s", rec.Body.String())
	}
}

func TestGetPublicEmployer_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/employers/abc",
		nil,
	)
	req.SetPathValue("id", "abc")

	rec := httptest.NewRecorder()
	h.GetPublicEmployer(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetPublicEmployer_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUC := profile.NewMockProfileUsecase(ctrl)
	h := profile.NewHandler(mockUC)

	mockUC.EXPECT().
		GetEmployerByID(gomock.Any(), 99).
		Return(models.EmployerProfile{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/employers/99",
		nil,
	)
	req.SetPathValue("id", "99")

	rec := httptest.NewRecorder()
	h.GetPublicEmployer(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
