package resume

import (
	"testing"

	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/services/seeker/usecase/resume"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestResumeRPCsAreUnimplemented(t *testing.T) {
	h := NewHandler(resume.NewResumeUsecase(nil))
	ctx := t.Context()
	_, createErr := h.CreateResume(ctx, &seekerv1.CreateResumeRequest{})
	_, updateErr := h.UpdateResume(ctx, &seekerv1.UpdateResumeRequest{})
	_, deleteErr := h.DeleteResume(ctx, &seekerv1.DeleteResumeRequest{})
	_, getErr := h.GetResume(ctx, &seekerv1.GetResumeRequest{})
	_, listErr := h.ListResumes(ctx, &seekerv1.ListResumesRequest{})
	for _, err := range []error{createErr, updateErr, deleteErr, getErr, listErr} {
		if status.Code(err) != codes.Unimplemented {
			t.Fatalf("got %v, want Unimplemented", err)
		}
	}
}
