package resume

import (
	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/usecase/resume"
)

type Handler struct {
	seekerv1.UnimplementedResumeServiceServer
	usecase *resume.ResumeUsecase
}

var _ seekerv1.ResumeServiceServer = (*Handler)(nil)

func NewHandler(usecase *resume.ResumeUsecase) *Handler {
	return &Handler{usecase: usecase}
}
