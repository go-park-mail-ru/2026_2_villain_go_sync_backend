package profile_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	profile "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/profile"
	"go.uber.org/mock/gomock"
)

func imageKey(s string) *string { return &s }

func TestProfileImageUsecase_Get(t *testing.T) {
	errRepo := errors.New("database unavailable")
	errStorage := errors.New("storage unavailable")
	tests := []struct {
		name       string
		employer   bool
		key        *string
		repoErr    error
		storageErr error
		wantErr    error
	}{
		{name: "seeker success", key: imageKey("avatars/12/a.jpg")},
		{name: "seeker missing image", wantErr: apperrors.ErrNotFound},
		{name: "seeker repo error", repoErr: errRepo, wantErr: errRepo},
		{name: "seeker storage error", key: imageKey("avatars/12/a.jpg"), storageErr: errStorage, wantErr: errStorage},
		{name: "employer success", employer: true, key: imageKey("logos/12/a.jpg")},
		{name: "employer missing image", employer: true, wantErr: apperrors.ErrNotFound},
		{name: "employer repo error", employer: true, repoErr: errRepo, wantErr: errRepo},
		{name: "employer storage error", employer: true, key: imageKey("logos/12/a.jpg"), storageErr: errStorage, wantErr: errStorage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := profile.NewMockProfileRepository(ctrl)
			storage := profile.NewMockImageStorage(ctrl)
			uc := profile.NewProfileUsecase(repo, storage)
			ctx := context.Background()
			if tt.employer {
				repo.EXPECT().GetEmployerByID(gomock.Any(), 12).
					Return(models.EmployerProfile{LogoKey: tt.key}, tt.repoErr)
			} else {
				repo.EXPECT().GetSeekerByID(gomock.Any(), 12).
					Return(models.SeekerProfile{AvatarKey: tt.key}, tt.repoErr)
			}
			if tt.repoErr == nil && tt.key != nil {
				if tt.storageErr != nil {
					storage.EXPECT().Download(gomock.Any(), *tt.key).
						Return(nil, "", tt.storageErr)
				} else {
					storage.EXPECT().Download(gomock.Any(), *tt.key).
						Return(io.NopCloser(strings.NewReader("test-image")), "image/jpeg", nil)
				}
			}
			var body io.ReadCloser
			var ct string
			var err error
			if tt.employer {
				body, ct, err = uc.GetEmployerLogo(ctx, 12)
			} else {
				body, ct, err = uc.GetSeekerAvatar(ctx, 12)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
				if body != nil {
					body.Close()
					t.Error("body must be nil on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if body == nil {
				t.Fatal("nil image body")
			}
			defer body.Close()
			b, readErr := io.ReadAll(body)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if string(b) != "test-image" {
				t.Errorf("unexpected body %q", b)
			}
			if ct != "image/jpeg" {
				t.Errorf("unexpected content type %q", ct)
			}
		})
	}
}

func TestProfileImageUsecase_Delete(t *testing.T) {
	errRepo := errors.New("repo error")
	errUpdate := errors.New("update error")
	errDelete := errors.New("storage delete error")
	tests := []struct {
		name       string
		employer   bool
		forbidden  bool
		oldKey     *string
		repoErr    error
		updateErr  error
		storageErr error
		wantErr    error
	}{
		{name: "seeker success", oldKey: imageKey("avatars/12/old.jpg")},
		{name: "seeker forbidden", forbidden: true, wantErr: apperrors.ErrForbidden},
		{name: "seeker no image", wantErr: apperrors.ErrNotFound},
		{name: "seeker repo error", repoErr: errRepo, wantErr: errRepo},
		{name: "seeker update error", oldKey: imageKey("avatars/12/old.jpg"), updateErr: errUpdate, wantErr: errUpdate},
		{name: "seeker delete failure ignored", oldKey: imageKey("avatars/12/old.jpg"), storageErr: errDelete},
		{name: "employer success", employer: true, oldKey: imageKey("logos/12/old.jpg")},
		{name: "employer forbidden", employer: true, forbidden: true, wantErr: apperrors.ErrForbidden},
		{name: "employer no image", employer: true, wantErr: apperrors.ErrNotFound},
		{name: "employer repo error", employer: true, repoErr: errRepo, wantErr: errRepo},
		{name: "employer update error", employer: true, oldKey: imageKey("logos/12/old.jpg"), updateErr: errUpdate, wantErr: errUpdate},
		{name: "employer delete failure ignored", employer: true, oldKey: imageKey("logos/12/old.jpg"), storageErr: errDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := profile.NewMockProfileRepository(ctrl)
			storage := profile.NewMockImageStorage(ctrl)
			uc := profile.NewProfileUsecase(repo, storage)
			role := models.RoleSeeker
			if tt.employer {
				role = models.RoleEmployer
			}
			if tt.forbidden {
				if tt.employer {
					role = models.RoleSeeker
				} else {
					role = models.RoleEmployer
				}
			} else {
				if tt.employer {
					repo.EXPECT().GetEmployerByID(gomock.Any(), 12).
						Return(models.EmployerProfile{LogoKey: tt.oldKey}, tt.repoErr)
				} else {
					repo.EXPECT().GetSeekerByID(gomock.Any(), 12).
						Return(models.SeekerProfile{AvatarKey: tt.oldKey}, tt.repoErr)
				}
				if tt.repoErr == nil && tt.oldKey != nil {
					if tt.employer {
						repo.EXPECT().UpdateEmployerLogoKey(gomock.Any(), 12, (*string)(nil)).Return(tt.updateErr)
					} else {
						repo.EXPECT().UpdateSeekerAvatarKey(gomock.Any(), 12, (*string)(nil)).Return(tt.updateErr)
					}
					if tt.updateErr == nil {
						storage.EXPECT().Delete(gomock.Any(), *tt.oldKey).Return(tt.storageErr)
					}
				}
			}
			var err error
			if tt.employer {
				err = uc.DeleteEmployerLogo(context.Background(), 12, role)
			} else {
				err = uc.DeleteSeekerAvatar(context.Background(), 12, role)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestProfileImageUsecase_Upload(t *testing.T) {
	errRepo := errors.New("repo error")
	errUpload := errors.New("upload error")
	errUpdate := errors.New("update error")
	errDelete := errors.New("cleanup error")
	tests := []struct {
		name      string
		employer  bool
		forbidden bool
		oldKey    *string
		repoErr   error
		uploadErr error
		updateErr error
		deleteErr error
		wantErr   error
	}{
		{name: "seeker new image"},
		{name: "seeker replace image", oldKey: imageKey("avatars/12/old.png")},
		{name: "seeker forbidden", forbidden: true, wantErr: apperrors.ErrForbidden},
		{name: "seeker repo error", repoErr: errRepo, wantErr: errRepo},
		{name: "seeker upload error", uploadErr: errUpload, wantErr: errUpload},
		{name: "seeker update error cleans new image", updateErr: errUpdate, wantErr: errUpdate},
		{name: "seeker cleanup failure", updateErr: errUpdate, deleteErr: errDelete, wantErr: errUpdate},
		{name: "seeker old image delete failure ignored", oldKey: imageKey("avatars/12/old.png"), deleteErr: errDelete},
		{name: "employer new image", employer: true},
		{name: "employer replace image", employer: true, oldKey: imageKey("logos/12/old.png")},
		{name: "employer forbidden", employer: true, forbidden: true, wantErr: apperrors.ErrForbidden},
		{name: "employer repo error", employer: true, repoErr: errRepo, wantErr: errRepo},
		{name: "employer upload error", employer: true, uploadErr: errUpload, wantErr: errUpload},
		{name: "employer update error cleans new image", employer: true, updateErr: errUpdate, wantErr: errUpdate},
		{name: "employer cleanup failure", employer: true, updateErr: errUpdate, deleteErr: errDelete, wantErr: errUpdate},
		{name: "employer old image delete failure ignored", employer: true, oldKey: imageKey("logos/12/old.png"), deleteErr: errDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := profile.NewMockProfileRepository(ctrl)
			storage := profile.NewMockImageStorage(ctrl)
			uc := profile.NewProfileUsecase(repo, storage)
			role := models.RoleSeeker
			prefix := "avatars/12/"
			if tt.employer {
				role = models.RoleEmployer
				prefix = "logos/12/"
			}
			if tt.forbidden {
				if tt.employer {
					role = models.RoleSeeker
				} else {
					role = models.RoleEmployer
				}
			} else {
				if tt.employer {
					repo.EXPECT().GetEmployerByID(gomock.Any(), 12).
						Return(models.EmployerProfile{LogoKey: tt.oldKey}, tt.repoErr)
				} else {
					repo.EXPECT().GetSeekerByID(gomock.Any(), 12).
						Return(models.SeekerProfile{AvatarKey: tt.oldKey}, tt.repoErr)
				}
				if tt.repoErr == nil {
					var generatedKey string
					storage.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any(), "image/png").
						DoAndReturn(func(_ context.Context, key string, body io.ReadSeeker, ct string) error {
							generatedKey = key
							if !strings.HasPrefix(key, prefix) {
								t.Errorf("unexpected key %q", key)
							}
							if ct != "image/png" {
								t.Errorf("unexpected content type %q", ct)
							}
							data, err := io.ReadAll(body)
							if err != nil || string(data) != "fake-image" {
								t.Errorf("unexpected uploaded body %q; err=%v", data, err)
							}
							return tt.uploadErr
						})
					if tt.uploadErr == nil {
						update := func(_ context.Context, userID int, key *string) error {
							if userID != 12 || key == nil || *key != generatedKey {
								t.Errorf("unexpected updated key: id=%d key=%v", userID, key)
							}
							return tt.updateErr
						}
						if tt.employer {
							repo.EXPECT().UpdateEmployerLogoKey(gomock.Any(), 12, gomock.Any()).DoAndReturn(update)
						} else {
							repo.EXPECT().UpdateSeekerAvatarKey(gomock.Any(), 12, gomock.Any()).DoAndReturn(update)
						}
						if tt.updateErr != nil {
							storage.EXPECT().Delete(gomock.Any(), gomock.Any()).
								DoAndReturn(func(_ context.Context, key string) error {
									if key != generatedKey {
										t.Errorf("wrong cleanup key %q", key)
									}
									return tt.deleteErr
								})
						} else if tt.oldKey != nil {
							storage.EXPECT().Delete(gomock.Any(), *tt.oldKey).Return(tt.deleteErr)
						}
					}
				}
			}
			body := strings.NewReader("fake-image")
			var err error
			if tt.employer {
				err = uc.UploadEmployerLogo(context.Background(), 12, role, body, "image/png")
			} else {
				err = uc.UploadSeekerAvatar(context.Background(), 12, role, body, "image/png")
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if tt.updateErr != nil && tt.deleteErr != nil && !errors.Is(err, tt.deleteErr) {
				t.Errorf("cleanup error should be included: %v", err)
			}
		})
	}
}
