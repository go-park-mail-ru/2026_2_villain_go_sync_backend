package company

import (
	"context"
	"errors"
	"io"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxLogoSize = 10 << 20

type CompanyUsecase interface {
	GetEmployerByID(ctx context.Context, employerID int) (models.EmployerProfile, error)
	GetEmployerLogo(ctx context.Context, userID int) (io.ReadCloser, string, error)
}

type Handler struct {
	vacancyv1.UnimplementedCompanyServiceServer
	usecase CompanyUsecase
}

var _ vacancyv1.CompanyServiceServer = (*Handler)(nil)

func NewHandler(usecase CompanyUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) GetCompanyProfile(ctx context.Context, req *vacancyv1.GetCompanyProfileRequest) (*vacancyv1.GetCompanyProfileResponse, error) {
	id, err := employerID(req.GetEmployerId())
	if err != nil {
		return nil, err
	}
	profile, err := h.usecase.GetEmployerByID(ctx, id)
	if err != nil {
		return nil, companyRPCError(err, "company profile not found")
	}
	return &vacancyv1.GetCompanyProfileResponse{Profile: toProtoCompanyProfile(profile)}, nil
}

func (h *Handler) GetCompanyLogo(ctx context.Context, req *vacancyv1.GetCompanyLogoRequest) (*vacancyv1.GetCompanyLogoResponse, error) {
	id, err := employerID(req.GetEmployerId())
	if err != nil {
		return nil, err
	}
	body, contentType, err := h.usecase.GetEmployerLogo(ctx, id)
	if err != nil {
		return nil, companyRPCError(err, "company logo not found")
	}

	content, readErr := io.ReadAll(io.LimitReader(body, maxLogoSize+1))
	closeErr := body.Close()
	if readErr != nil {
		return nil, companyRPCError(readErr, "company logo not found")
	}
	if len(content) > maxLogoSize {
		return nil, status.Error(codes.ResourceExhausted, "company logo exceeds 10 MiB")
	}
	if closeErr != nil {
		return nil, companyRPCError(closeErr, "company logo not found")
	}

	return &vacancyv1.GetCompanyLogoResponse{Content: content, ContentType: contentType}, nil
}

func employerID(value int64) (int, error) {
	id := int(value)
	if value <= 0 || int64(id) != value {
		return 0, status.Error(codes.InvalidArgument, "invalid employer id")
	}
	return id, nil
}

func companyRPCError(err error, notFoundMessage string) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, apperrors.ErrNotFound):
		return status.Error(codes.NotFound, notFoundMessage)
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
