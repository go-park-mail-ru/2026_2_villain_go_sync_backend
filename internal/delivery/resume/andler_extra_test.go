package resume_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/resume"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestResumeHandlers_Table(t *testing.T) {
	tests := []struct {
		name, action, method, url, id, body string
		auth                                bool
		wantStatus                          int
		wantBody                            string
		setup                               func(*resume.MockResumeUsecase)
	}{
		{name: "create unauthorized", action: "create", method: http.MethodPost, url: "/api/resumes", body: `{"title":"CV"}`, wantStatus: 401},
		{name: "create bad json", action: "create", method: http.MethodPost, url: "/api/resumes", body: "{", auth: true, wantStatus: 400},
		{name: "create missing title", action: "create", method: http.MethodPost, url: "/api/resumes", body: `{"title":" "}`, auth: true, wantStatus: 422},
		{name: "create success", action: "create", method: http.MethodPost, url: "/api/resumes", body: `{"title":"CV","description":"Go dev"}`, auth: true, wantStatus: 201, wantBody: `"title":"CV"`, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Create(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).
				Return(models.Resume{ID: 31, Title: "CV"}, nil)
		}},
		{name: "create forbidden", action: "create", method: http.MethodPost, url: "/api/resumes", body: `{"title":"CV"}`, auth: true, wantStatus: 403, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Create(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).Return(models.Resume{}, apperrors.ErrForbidden)
		}},

		{name: "get invalid id", action: "get", method: http.MethodGet, url: "/api/resumes/abc", id: "abc", wantStatus: 400},
		{name: "get success", action: "get", method: http.MethodGet, url: "/api/resumes/31", id: "31", wantStatus: 200, wantBody: `"title":"CV"`, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().GetByID(gomock.Any(), 31).Return(models.Resume{ID: 31, Title: "CV"}, nil)
		}},
		{name: "get not found", action: "get", method: http.MethodGet, url: "/api/resumes/31", id: "31", wantStatus: 404, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().GetByID(gomock.Any(), 31).Return(models.Resume{}, apperrors.ErrNotFound)
		}},

		{name: "list missing seeker id", action: "list", method: http.MethodGet, url: "/api/resumes", wantStatus: 400},
		{name: "list invalid seeker id", action: "list", method: http.MethodGet, url: "/api/resumes?seeker_id=abc", wantStatus: 400},
		{name: "list success", action: "list", method: http.MethodGet, url: "/api/resumes?seeker_id=10", wantStatus: 200, wantBody: `"title":"CV"`, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().ListBySeekerID(gomock.Any(), 10).Return([]models.Resume{{ID: 31, Title: "CV"}}, nil)
		}},
		{name: "list empty", action: "list", method: http.MethodGet, url: "/api/resumes?seeker_id=10", wantStatus: 200, wantBody: `"data":[]`, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().ListBySeekerID(gomock.Any(), 10).Return([]models.Resume{}, nil)
		}},
		{name: "list repository error", action: "list", method: http.MethodGet, url: "/api/resumes?seeker_id=10", wantStatus: 500, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().ListBySeekerID(gomock.Any(), 10).Return(nil, errors.New("db down"))
		}},

		{name: "update unauthorized", action: "update", method: http.MethodPut, url: "/api/resumes/31", id: "31", body: `{"title":"Updated"}`, wantStatus: 401},
		{name: "update invalid id", action: "update", method: http.MethodPut, url: "/api/resumes/bad", id: "bad", body: `{"title":"Updated"}`, auth: true, wantStatus: 400},
		{name: "update invalid json", action: "update", method: http.MethodPut, url: "/api/resumes/31", id: "31", body: "{", auth: true, wantStatus: 400},
		{name: "update missing title", action: "update", method: http.MethodPut, url: "/api/resumes/31", id: "31", body: `{}`, auth: true, wantStatus: 422},
		{name: "update success", action: "update", method: http.MethodPut, url: "/api/resumes/31", id: "31", body: `{"title":"Updated"}`, auth: true, wantStatus: 200, wantBody: `"title":"Updated"`, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Update(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).
				Return(models.Resume{ID: 31, Title: "Updated"}, nil)
		}},
		{name: "update forbidden", action: "update", method: http.MethodPut, url: "/api/resumes/31", id: "31", body: `{"title":"Updated"}`, auth: true, wantStatus: 403, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Update(gomock.Any(), 10, models.RoleSeeker, gomock.Any()).Return(models.Resume{}, apperrors.ErrForbidden)
		}},

		{name: "delete unauthorized", action: "delete", method: http.MethodDelete, url: "/api/resumes/31", id: "31", wantStatus: 401},
		{name: "delete invalid id", action: "delete", method: http.MethodDelete, url: "/api/resumes/wrong", id: "wrong", auth: true, wantStatus: 400},
		{name: "delete success", action: "delete", method: http.MethodDelete, url: "/api/resumes/31", id: "31", auth: true, wantStatus: 204, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Delete(gomock.Any(), 10, models.RoleSeeker, 31).Return(nil)
		}},
		{name: "delete not found", action: "delete", method: http.MethodDelete, url: "/api/resumes/31", id: "31", auth: true, wantStatus: 404, setup: func(m *resume.MockResumeUsecase) {
			m.EXPECT().Delete(gomock.Any(), 10, models.RoleSeeker, 31).Return(apperrors.ErrNotFound)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := resume.NewMockResumeUsecase(ctrl)
			h := resume.NewHandler(mockUC)

			if tt.setup != nil {
				tt.setup(mockUC)
			}

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			req.SetPathValue("id", tt.id)

			if tt.auth {
				ctx := auth.WithUserID(req.Context(), 10)
				ctx = auth.WithRole(ctx, models.RoleSeeker)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()

			switch tt.action {
			case "create":
				h.Create(rec, req)
			case "get":
				h.GetByID(rec, req)
			case "list":
				h.ListBySeekerID(rec, req)
			case "update":
				h.Update(rec, req)
			case "delete":
				h.Delete(rec, req)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("status: got %d, want %d; body: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s",
					tt.wantBody, rec.Body.String())
			}
		})
	}
}
