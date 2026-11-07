package vacancy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	dto "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
	"go.uber.org/mock/gomock"
)

func TestVacancyListExtra(t *testing.T) {
	tests := []struct {
		name     string
		employer bool
		repoErr  error
		forbid   bool
	}{
		{name: "list all success"},
		{name: "list all error", repoErr: errors.New("db failure")},
		{name: "list employer success", employer: true},
		{name: "list employer repo error", employer: true, repoErr: errors.New("db failure")},
		{name: "list employer forbidden", employer: true, forbid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			vr := vacancy.NewMockVacancyRepository(ctrl)
			ur := vacancy.NewMockUserRepository(ctrl)
			uc := vacancy.NewVacancyUsecase(vr, ur)

			expected := []models.Vacancy{{ID: 8, Title: "Go"}}

			if tt.employer {
				if tt.forbid {
					ur.EXPECT().GetRoleByID(
						gomock.Any(), int64(10),
					).Return("seeker", nil)
				} else {
					ur.EXPECT().GetRoleByID(
						gomock.Any(), int64(10),
					).Return("employer", nil)

					vr.EXPECT().ListByEmployer(
						gomock.Any(), int64(10),
					).Return(expected, tt.repoErr)
				}
			} else {
				vr.EXPECT().ListVacancies(
					gomock.Any(),
				).Return(expected, tt.repoErr)
			}

			var got []models.Vacancy
			var err error

			if tt.employer {
				got, err = uc.ListByEmployer(context.Background(), 10)
			} else {
				got, err = uc.ListVacancies(context.Background())
			}

			switch {
			case tt.forbid:
				if !errors.Is(err, apperrors.ErrForbidden) {
					t.Fatalf("expected forbidden, got %v", err)
				}
			case tt.repoErr != nil:
				if !errors.Is(err, tt.repoErr) {
					t.Fatalf("expected DB error, got %v", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 1 || got[0].Title != "Go" {
					t.Errorf("unexpected vacancies: %+v", got)
				}
			}
		})
	}
}

func TestVacancyUpdateExtra(t *testing.T) {
	title := "Updated Go"
	emptyTitle := " "
	invalidCurrency := "BTC"

	tests := []struct {
		name       string
		input      dto.UpdateVacancyInput
		old        models.Vacancy
		getErr     error
		updateErr  error
		wantErr    error
		callUpdate bool
	}{
		{
			name:    "invalid ID",
			input:   dto.UpdateVacancyInput{EmployerID: 10},
			wantErr: apperrors.ErrInvalidInput,
		},
		{
			name:    "get not found",
			input:   dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7},
			getErr:  apperrors.ErrNotFound,
			wantErr: apperrors.ErrNotFound,
		},
		{
			name:    "not owner",
			input:   dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7},
			old:     models.Vacancy{EmployerID: 99, Title: "Go", Currency: "RUB"},
			wantErr: apperrors.ErrForbidden,
		},
		{
			name:    "empty title",
			input:   dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7, Title: &emptyTitle},
			old:     models.Vacancy{EmployerID: 10, Title: "Go", Currency: "RUB"},
			wantErr: apperrors.ErrInvalidInput,
		},
		{
			name:    "invalid currency",
			input:   dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7, Currency: &invalidCurrency},
			old:     models.Vacancy{EmployerID: 10, Title: "Go", Currency: "RUB"},
			wantErr: apperrors.ErrInvalidInput,
		},
		{
			name:       "success",
			input:      dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7, Title: &title},
			old:        models.Vacancy{ID: 7, EmployerID: 10, Title: "Go", Currency: "RUB"},
			callUpdate: true,
		},
		{
			name:       "repository error",
			input:      dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7, Title: &title},
			old:        models.Vacancy{ID: 7, EmployerID: 10, Title: "Go", Currency: "RUB"},
			callUpdate: true,
			updateErr:  errors.New("update failed"),
		},
		{
			name:       "category not found",
			input:      dto.UpdateVacancyInput{EmployerID: 10, VacancyID: 7, Title: &title},
			old:        models.Vacancy{ID: 7, EmployerID: 10, Title: "Go", Currency: "RUB"},
			callUpdate: true,
			updateErr:  apperrors.ErrCategoryNotFound,
			wantErr:    apperrors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			vr := vacancy.NewMockVacancyRepository(ctrl)
			ur := vacancy.NewMockUserRepository(ctrl)
			uc := vacancy.NewVacancyUsecase(vr, ur)

			ur.EXPECT().GetRoleByID(
				gomock.Any(), int64(10),
			).Return("employer", nil)

			ur.EXPECT().HasEmployerProfile(
				gomock.Any(), int64(10),
			).Return(true, nil)

			if tt.input.VacancyID > 0 {
				vr.EXPECT().GetByID(
					gomock.Any(), tt.input.VacancyID,
				).Return(tt.old, tt.getErr)
			}

			if tt.callUpdate {
				returned := tt.old
				returned.Title = title

				vr.EXPECT().Update(
					gomock.Any(), gomock.Any(), gomock.Any(),
				).Return(returned, tt.updateErr)
			}

			got, err := uc.Update(context.Background(), tt.input)

			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			case tt.updateErr != nil:
				if !errors.Is(err, tt.updateErr) {
					t.Fatalf("expected %v, got %v", tt.updateErr, err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}

			if tt.name == "success" && got.Title != title {
				t.Errorf("wrong updated title: %q", got.Title)
			}
		})
	}
}

func TestVacancyListEmployerApplicantsExtra(t *testing.T) {
	tests := []struct {
		name    string
		role    models.Role
		repoErr error
		wantErr error
	}{
		{
			name:    "forbidden",
			role:    models.RoleSeeker,
			wantErr: apperrors.ErrForbidden,
		},
		{
			name: "success",
			role: models.RoleEmployer,
		},
		{
			name:    "repository error",
			role:    models.RoleEmployer,
			repoErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			vr := vacancy.NewMockVacancyRepository(ctrl)
			ur := vacancy.NewMockUserRepository(ctrl)
			uc := vacancy.NewVacancyUsecase(vr, ur)

			want := []models.Applicant{
				{ApplicationID: 3, FirstName: "Ivan"},
			}

			if tt.role == models.RoleEmployer {
				vr.EXPECT().ListByEmployerID(
					gomock.Any(), int64(10),
				).Return(want, tt.repoErr)
			}

			got, err := uc.ListEmployerApplicants(
				context.Background(), 10, tt.role,
			)

			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			case tt.repoErr != nil:
				if !errors.Is(err, tt.repoErr) {
					t.Fatalf("expected %v, got %v", tt.repoErr, err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 1 || got[0].ApplicationID != 3 {
					t.Errorf("unexpected applicants: %+v", got)
				}
			}
		})
	}
}
