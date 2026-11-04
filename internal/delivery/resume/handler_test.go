package resume_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/resume"
	mocks "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/mocks/delivery_resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestGetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockResumeUsecase(ctrl)
	h := resume.NewHandler(mockUC)

	mockUC.EXPECT().
		GetByID(gomock.Any(), 1).
		Return(models.Resume{Title: "Go Dev"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/resumes/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockResumeUsecase(ctrl)
	h := resume.NewHandler(mockUC)

	body := `{invalid_json`
	req := httptest.NewRequest(http.MethodPost, "/api/resumes", bytes.NewBufferString(body))

	ctx := auth.WithUserID(req.Context(), 10)
	ctx = auth.WithRole(ctx, models.RoleSeeker)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
