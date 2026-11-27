package profile

import (
	"context"
	"errors"
	"io"

	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxAvatarSize = 10 << 20

type ProfileUsecase interface {
	GetSeekerByID(ctx context.Context, seekerID int) (models.SeekerProfile, error)
	GetSeekerAvatar(ctx context.Context, userID int) (io.ReadCloser, string, error)
}

type Handler struct {
	seekerv1.UnimplementedSeekerServiceServer
	usecase ProfileUsecase
}

var _ seekerv1.SeekerServiceServer = (*Handler)(nil)

func NewHandler(usecase ProfileUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) GetSeekerProfile(ctx context.Context, req *seekerv1.GetSeekerProfileRequest) (*seekerv1.GetSeekerProfileResponse, error) {
	id, err := seekerID(req.GetSeekerId())
	if err != nil {
		return nil, err
	}
	profile, err := h.usecase.GetSeekerByID(ctx, id)
	if err != nil {
		return nil, profileRPCError(err, "seeker profile not found")
	}
	return &seekerv1.GetSeekerProfileResponse{Profile: toProtoSeekerProfile(profile)}, nil
}

func (h *Handler) GetSeekerAvatar(ctx context.Context, req *seekerv1.GetSeekerAvatarRequest) (*seekerv1.GetSeekerAvatarResponse, error) {
	id, err := seekerID(req.GetSeekerId())
	if err != nil {
		return nil, err
	}
	body, contentType, err := h.usecase.GetSeekerAvatar(ctx, id)
	if err != nil {
		return nil, profileRPCError(err, "seeker avatar not found")
	}

	content, readErr := io.ReadAll(io.LimitReader(body, maxAvatarSize+1))
	closeErr := body.Close()
	if readErr != nil {
		return nil, profileRPCError(readErr, "seeker avatar not found")
	}
	if len(content) > maxAvatarSize {
		return nil, status.Error(codes.ResourceExhausted, "seeker avatar exceeds 10 MiB")
	}
	if closeErr != nil {
		return nil, profileRPCError(closeErr, "seeker avatar not found")
	}

	return &seekerv1.GetSeekerAvatarResponse{Content: content, ContentType: contentType}, nil
}

func seekerID(value int64) (int, error) {
	id := int(value)
	if value <= 0 || int64(id) != value {
		return 0, status.Error(codes.InvalidArgument, "invalid seeker id")
	}
	return id, nil
}

func profileRPCError(err error, notFoundMessage string) error {
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
