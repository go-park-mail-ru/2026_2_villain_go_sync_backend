package vacancy

import (
	"context"
	"errors"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/dto/pagination"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VacancyUsecase interface {
	ListVacancies(ctx context.Context, params pagination.Params) ([]models.Vacancy, int64, error)
	GetByID(ctx context.Context, id int64) (models.Vacancy, error)
	ListCategories(ctx context.Context) ([]models.Category, error)
}

type Handler struct {
	vacancyv1.UnimplementedVacancyServiceServer
	usecase VacancyUsecase
}

var _ vacancyv1.VacancyServiceServer = (*Handler)(nil)

func NewHandler(usecase VacancyUsecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) GetVacancy(ctx context.Context, req *vacancyv1.GetVacancyRequest) (*vacancyv1.GetVacancyResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid vacancy id")
	}
	vacancy, err := h.usecase.GetByID(ctx, req.GetId())
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			return nil, status.Error(codes.NotFound, "vacancy not found")
		case errors.Is(err, context.Canceled):
			return nil, status.Error(codes.Canceled, "request canceled")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}
	return &vacancyv1.GetVacancyResponse{Vacancy: toProtoVacancy(vacancy)}, nil
}

func (h *Handler) ListVacancies(ctx context.Context, req *vacancyv1.ListVacanciesRequest) (*vacancyv1.ListVacanciesResponse, error) {
	var limit *int
	var cursor *string
	if req != nil {
		if req.Limit != nil {
			value := int(*req.Limit)
			limit = &value
		}
		cursor = req.Cursor
	}
	params, err := pagination.NewParams(limit, cursor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	vacancies, total, err := h.usecase.ListVacancies(ctx, params.WithLookahead())
	if err != nil {
		return nil, listRPCError(err)
	}

	page, err := pagination.NewPage(vacancies, params.Limit, total, func(v models.Vacancy) pagination.Cursor {
		return pagination.Cursor{CreatedAt: v.CreatedAt, ID: v.ID}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}
	items := make([]*vacancyv1.Vacancy, len(page.Items))
	for i, vacancy := range page.Items {
		items[i] = toProtoVacancy(vacancy)
	}
	return &vacancyv1.ListVacanciesResponse{
		Items: items,
		Pagination: &vacancyv1.Pagination{
			Limit:      int32(page.Pagination.Limit),
			Total:      page.Pagination.Total,
			NextCursor: page.Pagination.NextCursor,
		},
	}, nil
}

func (h *Handler) ListCategories(ctx context.Context, req *emptypb.Empty) (*vacancyv1.ListCategoriesResponse, error) {
	categories, err := h.usecase.ListCategories(ctx)
	if err != nil {
		return nil, listRPCError(err)
	}
	protoCategories := make([]*vacancyv1.Category, 0, len(categories))
	for _, category := range categories {
		protoCategories = append(protoCategories, &vacancyv1.Category{
			Id:   category.ID,
			Name: category.Name,
		})
	}
	return &vacancyv1.ListCategoriesResponse{Categories: protoCategories}, nil
}

func listRPCError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
