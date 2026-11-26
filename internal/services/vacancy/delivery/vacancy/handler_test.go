package vacancy

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type vacancyStub struct {
	get        func(context.Context, int64) (models.Vacancy, error)
	list       func(context.Context, pagination.Params) ([]models.Vacancy, int64, error)
	categories func(context.Context) ([]models.Category, error)
}

func (s vacancyStub) GetByID(ctx context.Context, id int64) (models.Vacancy, error) {
	return s.get(ctx, id)
}

func (s vacancyStub) ListVacancies(ctx context.Context, params pagination.Params) ([]models.Vacancy, int64, error) {
	return s.list(ctx, params)
}

func (s vacancyStub) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.categories(ctx)
}

func TestGetVacancy(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 400, time.UTC)
	updated := created.Add(time.Hour)
	changed := updated.Add(time.Minute)
	experience := models.ExperienceOneToThree
	format := models.WorkFormatRemote
	employment := models.EmploymentFull
	schedule := models.WorkScheduleFlexible
	model := models.Vacancy{
		ID: 42, Version: 3, Title: "Developer", ApplyUntil: proto.String("2025-02-01"),
		CreatedAt: created, UpdatedAt: updated, ChangedAt: changed, Categories: []string{"IT"},
		Company:      models.VacancyCompany{EmployerID: 5, Name: "Example", LogoURL: proto.String("/api/employers/5/logo")},
		Content:      models.VacancyContent{Description: "Job", Responsibilities: "Development"},
		Requirements: models.VacancyRequirements{Description: "Go", ExperienceLevel: &experience},
		Salary:       models.VacancySalary{SalaryFrom: proto.Float64(0), Currency: models.CurrencyRUB},
		Conditions: models.VacancyConditions{
			Offer: proto.String(""), EmploymentType: &employment, WorkFormat: &format,
			WorkSchedule: &schedule, WorkHours: proto.String("8"),
		},
		Location: models.VacancyLocation{City: proto.String("Moscow"), OfficeAddress: proto.String("Office")},
	}
	h := NewHandler(vacancyStub{get: func(_ context.Context, id int64) (models.Vacancy, error) {
		if id != 42 {
			t.Fatalf("use case received id %d", id)
		}
		return model, nil
	}})
	want := &vacancyv1.Vacancy{
		Id: 42, Version: 3, Title: "Developer", ApplyUntil: proto.String("2025-02-01"),
		CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(updated), ChangedAt: timestamppb.New(changed),
		Categories:   []string{"IT"},
		Company:      &vacancyv1.Vacancy_Company{EmployerId: 5, Name: "Example", LogoUrl: proto.String("/api/employers/5/logo")},
		Content:      &vacancyv1.Vacancy_Content{Description: "Job", Responsibilities: "Development"},
		Requirements: &vacancyv1.Vacancy_Requirements{Description: "Go", ExperienceLevel: proto.String("1_3")},
		Salary:       &vacancyv1.Vacancy_Salary{SalaryFrom: proto.Float64(0), Currency: "RUB"},
		Conditions: &vacancyv1.Vacancy_Conditions{
			Offer: proto.String(""), EmploymentType: proto.String("full"), WorkFormat: proto.String("remote"),
			WorkSchedule: proto.String("flexible"), WorkHours: proto.String("8"),
		},
		Location: &vacancyv1.Vacancy_Location{City: proto.String("Moscow"), OfficeAddress: proto.String("Office")},
	}
	got, err := h.GetVacancy(t.Context(), &vacancyv1.GetVacancyRequest{Id: 42})
	if err != nil || !proto.Equal(got.GetVacancy(), want) {
		t.Fatalf("GetVacancy() = %v, %v, want %v", got, err, want)
	}
}

func TestVacancyRPCInvalidRequests(t *testing.T) {
	h := NewHandler(vacancyStub{})
	for _, req := range []*vacancyv1.GetVacancyRequest{nil, {Id: 0}, {Id: -1}} {
		if _, err := h.GetVacancy(t.Context(), req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("GetVacancy(%v) error = %v", req, err)
		}
	}
	for _, req := range []*vacancyv1.ListVacanciesRequest{
		{Limit: proto.Int32(0)}, {Limit: proto.Int32(-1)}, {Limit: proto.Int32(101)},
		{Cursor: proto.String("")}, {Cursor: proto.String("invalid!")},
	} {
		if _, err := h.ListVacancies(t.Context(), req); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("ListVacancies(%v) error = %v", req, err)
		}
	}
}

func TestListVacancies(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	first := models.Vacancy{ID: 42, Title: "First", CreatedAt: created}
	second := models.Vacancy{ID: 41, Title: "Second", CreatedAt: created}
	h := NewHandler(vacancyStub{list: func(_ context.Context, params pagination.Params) ([]models.Vacancy, int64, error) {
		if params.Limit != 2 {
			t.Fatalf("use case received limit %d, want 2", params.Limit)
		}
		if params.Cursor == nil {
			return []models.Vacancy{first, second}, 2, nil
		}
		if params.Cursor.ID != first.ID || !params.Cursor.CreatedAt.Equal(created) {
			t.Fatalf("use case received cursor %+v", params.Cursor)
		}
		return []models.Vacancy{second}, 2, nil
	}})
	req := &vacancyv1.ListVacanciesRequest{Limit: proto.Int32(1)}
	got, err := h.ListVacancies(t.Context(), req)
	if err != nil || len(got.GetItems()) != 1 || got.GetItems()[0].GetId() != first.ID ||
		got.GetItems()[0].GetTitle() != first.Title || got.GetPagination().GetLimit() != 1 ||
		got.GetPagination().GetTotal() != 2 || got.GetPagination().GetNextCursor() == "" {
		t.Fatalf("first page: %v, %v", got, err)
	}

	req.Cursor = got.GetPagination().NextCursor
	got, err = h.ListVacancies(t.Context(), req)
	if err != nil || len(got.GetItems()) != 1 || got.GetItems()[0].GetId() != second.ID ||
		got.GetPagination() == nil || got.GetPagination().NextCursor != nil {
		t.Fatalf("second page: %v, %v", got, err)
	}
}

func TestListVacanciesDefaultLimit(t *testing.T) {
	h := NewHandler(vacancyStub{list: func(_ context.Context, params pagination.Params) ([]models.Vacancy, int64, error) {
		if params.Limit != 13 || params.Cursor != nil {
			t.Fatalf("unexpected params: %+v", params)
		}
		return nil, 0, nil
	}})
	got, err := h.ListVacancies(t.Context(), &vacancyv1.ListVacanciesRequest{})
	want := &vacancyv1.ListVacanciesResponse{Pagination: &vacancyv1.Pagination{Limit: 12}}
	if err != nil || !proto.Equal(got, want) {
		t.Fatalf("ListVacancies() = %v, %v, want %v", got, err, want)
	}
}

func TestListCategories(t *testing.T) {
	h := NewHandler(vacancyStub{categories: func(context.Context) ([]models.Category, error) {
		return []models.Category{{ID: 2, Name: "IT"}, {ID: 1, Name: "Design"}}, nil
	}})
	got, err := h.ListCategories(t.Context(), &emptypb.Empty{})
	want := &vacancyv1.ListCategoriesResponse{Categories: []*vacancyv1.Category{
		{Id: 2, Name: "IT"}, {Id: 1, Name: "Design"},
	}}
	if err != nil || !proto.Equal(got, want) {
		t.Fatalf("ListCategories() = %v, %v, want %v", got, err, want)
	}
}

func TestVacancyRPCDependencyErrors(t *testing.T) {
	for _, tc := range []struct {
		err      error
		getCode  codes.Code
		listCode codes.Code
	}{
		{context.Canceled, codes.Canceled, codes.Canceled},
		{context.DeadlineExceeded, codes.DeadlineExceeded, codes.DeadlineExceeded},
		{apperrors.ErrNotFound, codes.NotFound, codes.Internal},
		{errors.New("database error"), codes.Internal, codes.Internal},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			wrapped := fmt.Errorf("dependency: %w", tc.err)
			h := NewHandler(vacancyStub{
				get:        func(context.Context, int64) (models.Vacancy, error) { return models.Vacancy{}, wrapped },
				list:       func(context.Context, pagination.Params) ([]models.Vacancy, int64, error) { return nil, 0, wrapped },
				categories: func(context.Context) ([]models.Category, error) { return nil, wrapped },
			})
			_, getErr := h.GetVacancy(t.Context(), &vacancyv1.GetVacancyRequest{Id: 42})
			_, listErr := h.ListVacancies(t.Context(), &vacancyv1.ListVacanciesRequest{})
			_, categoriesErr := h.ListCategories(t.Context(), &emptypb.Empty{})
			if status.Code(getErr) != tc.getCode || status.Code(listErr) != tc.listCode || status.Code(categoriesErr) != tc.listCode {
				t.Fatalf("GetVacancy: %v, ListVacancies: %v, ListCategories: %v", getErr, listErr, categoriesErr)
			}
		})
	}
}
