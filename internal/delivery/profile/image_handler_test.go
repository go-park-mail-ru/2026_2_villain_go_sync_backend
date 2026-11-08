package profile_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/profile"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"go.uber.org/mock/gomock"
)

func TestGetProfileImage(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		id         string
		usecaseErr error
		wantStatus int
	}{
		{"seeker avatar success", "avatar", "12", nil, http.StatusOK},
		{"seeker avatar not found", "avatar", "12", apperrors.ErrNotFound, http.StatusNotFound},
		{"seeker avatar invalid ID", "avatar", "abc", nil, http.StatusBadRequest},
		{"employer logo success", "logo", "11", nil, http.StatusOK},
		{"employer logo not found", "logo", "11", apperrors.ErrNotFound, http.StatusNotFound},
		{"employer logo invalid ID", "logo", "abc", nil, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := profile.NewMockProfileUsecase(ctrl)
			h := profile.NewHandler(mockUC)

			var imageBody io.ReadCloser
			contentType := ""

			if tt.usecaseErr == nil {
				imageBody = io.NopCloser(strings.NewReader("image-data"))
				contentType = "image/png"
			}

			if tt.id != "abc" {
				switch tt.kind {
				case "avatar":
					mockUC.EXPECT().
						GetSeekerAvatar(gomock.Any(), 12).
						Return(imageBody, contentType, tt.usecaseErr)
				case "logo":
					mockUC.EXPECT().
						GetEmployerLogo(gomock.Any(), 11).
						Return(imageBody, contentType, tt.usecaseErr)
				}
			}

			url := "/api/seekers/" + tt.id + "/avatar"
			if tt.kind == "logo" {
				url = "/api/employers/" + tt.id + "/logo"
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()

			if tt.kind == "avatar" {
				h.GetSeekerAvatar(rec, req)
			} else {
				h.GetEmployerLogo(rec, req)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s",
					tt.wantStatus, rec.Code, rec.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				if rec.Body.String() != "image-data" {
					t.Errorf("unexpected image body: %s", rec.Body.String())
				}
				if rec.Header().Get("Content-Type") != "image/png" {
					t.Errorf("unexpected content type: %s",
						rec.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestDeleteProfileImage(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		authorized bool
		role       models.Role
		usecaseErr error
		wantStatus int
	}{
		{"delete seeker avatar", "avatar", true, models.RoleSeeker, nil, http.StatusNoContent},
		{"delete seeker avatar forbidden", "avatar", true, models.RoleEmployer, apperrors.ErrForbidden, http.StatusForbidden},
		{"delete seeker avatar unauthorized", "avatar", false, models.RoleSeeker, nil, http.StatusUnauthorized},
		{"delete employer logo", "logo", true, models.RoleEmployer, nil, http.StatusNoContent},
		{"delete employer logo forbidden", "logo", true, models.RoleSeeker, apperrors.ErrForbidden, http.StatusForbidden},
		{"delete employer logo unauthorized", "logo", false, models.RoleEmployer, nil, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := profile.NewMockProfileUsecase(ctrl)
			h := profile.NewHandler(mockUC)

			url := "/api/me/avatar"
			if tt.kind == "logo" {
				url = "/api/me/logo"
			}

			req := httptest.NewRequest(http.MethodDelete, url, nil)

			if tt.authorized {
				ctx := auth.WithUserID(req.Context(), 12)
				ctx = auth.WithRole(ctx, tt.role)
				req = req.WithContext(ctx)

				if tt.kind == "avatar" {
					mockUC.EXPECT().
						DeleteSeekerAvatar(gomock.Any(), 12, tt.role).
						Return(tt.usecaseErr)
				} else {
					mockUC.EXPECT().
						DeleteEmployerLogo(gomock.Any(), 12, tt.role).
						Return(tt.usecaseErr)
				}
			}

			rec := httptest.NewRecorder()

			if tt.kind == "avatar" {
				h.DeleteSeekerAvatar(rec, req)
			} else {
				h.DeleteEmployerLogo(rec, req)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d: %s",
					tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUploadProfileImage(t *testing.T) {
	pngData := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0}

	tests := []struct {
		name       string
		kind       string
		field      string
		authorized bool
		role       models.Role
		data       []byte
		usecaseErr error
		wantCall   bool
		wantStatus int
	}{
		{"upload avatar", "avatar", "avatar", true, models.RoleSeeker, pngData, nil, true, http.StatusNoContent},
		{"avatar unauthorized", "avatar", "avatar", false, models.RoleSeeker, pngData, nil, false, http.StatusUnauthorized},
		{"avatar forbidden", "avatar", "avatar", true, models.RoleEmployer, pngData, apperrors.ErrForbidden, true, http.StatusForbidden},
		{"avatar wrong field", "avatar", "wrong", true, models.RoleSeeker, pngData, nil, false, http.StatusBadRequest},
		{"avatar empty", "avatar", "avatar", true, models.RoleSeeker, nil, nil, false, http.StatusBadRequest},
		{"avatar unsupported format", "avatar", "avatar", true, models.RoleSeeker, []byte("not an image"), nil, false, http.StatusUnsupportedMediaType},
		{"upload logo", "logo", "logo", true, models.RoleEmployer, pngData, nil, true, http.StatusNoContent},
		{"logo unauthorized", "logo", "logo", false, models.RoleEmployer, pngData, nil, false, http.StatusUnauthorized},
		{"logo forbidden", "logo", "logo", true, models.RoleSeeker, pngData, apperrors.ErrForbidden, true, http.StatusForbidden},
		{"logo wrong field", "logo", "wrong", true, models.RoleEmployer, pngData, nil, false, http.StatusBadRequest},
		{"logo empty", "logo", "logo", true, models.RoleEmployer, nil, nil, false, http.StatusBadRequest},
		{"logo unsupported format", "logo", "logo", true, models.RoleEmployer, []byte("not an image"), nil, false, http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockUC := profile.NewMockProfileUsecase(ctrl)
			h := profile.NewHandler(mockUC)

			var body bytes.Buffer
			writer := multipart.NewWriter(&body)

			part, err := writer.CreateFormFile(tt.field, "image.png")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := part.Write(tt.data); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}

			url := "/api/me/avatar"
			if tt.kind == "logo" {
				url = "/api/me/logo"
			}

			req := httptest.NewRequest(http.MethodPut, url, &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			if tt.authorized {
				ctx := auth.WithUserID(req.Context(), 12)
				ctx = auth.WithRole(ctx, tt.role)
				req = req.WithContext(ctx)
			}

			if tt.wantCall {
				if tt.kind == "avatar" {
					mockUC.EXPECT().
						UploadSeekerAvatar(
							gomock.Any(), 12, tt.role,
							gomock.Any(), "image/png",
						).
						Return(tt.usecaseErr)
				} else {
					mockUC.EXPECT().
						UploadEmployerLogo(
							gomock.Any(), 12, tt.role,
							gomock.Any(), "image/png",
						).
						Return(tt.usecaseErr)
				}
			}

			rec := httptest.NewRecorder()

			if tt.kind == "avatar" {
				h.UploadSeekerAvatar(rec, req)
			} else {
				h.UploadEmployerLogo(rec, req)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d: %s",
					tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
