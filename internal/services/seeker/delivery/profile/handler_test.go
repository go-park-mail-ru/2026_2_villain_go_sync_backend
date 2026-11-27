package profile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	seekerv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/seeker/v1"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type profileStub struct {
	profile func(context.Context, int) (models.SeekerProfile, error)
	avatar  func(context.Context, int) (io.ReadCloser, string, error)
}

func (s profileStub) GetSeekerByID(ctx context.Context, id int) (models.SeekerProfile, error) {
	return s.profile(ctx, id)
}

func (s profileStub) GetSeekerAvatar(ctx context.Context, id int) (io.ReadCloser, string, error) {
	return s.avatar(ctx, id)
}

func TestGetSeekerProfile(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 400, time.UTC)
	updated := created.Add(time.Hour)
	workFormat := models.WorkFormatRemote
	year := int16(2020)
	model := models.SeekerProfile{
		Profile: models.Profile{UserID: 42, Description: "Developer", CreatedAt: created, UpdatedAt: updated},
		PersonalInfo: models.SeekerPersonalInfo{
			FirstName: "Anna", LastName: "Ivanova", City: proto.String(""), AvatarKey: proto.String("private/storage/key"),
		},
		Contacts: models.SeekerContacts{Phone: "1234", Messenger: proto.String("@anna"), ContactEmail: proto.String("anna@example.org")},
		Preferences: models.SeekerPreferences{
			PreferredPosition: proto.String("Backend"), PreferredCities: proto.String("Moscow"), WorkFormat: &workFormat,
		},
		Career: models.CareerDetails{
			ExperienceItems: []models.ExperienceItem{
				{CompanyName: proto.String("Example"), PositionTitle: proto.String("Engineer"), StartedOn: proto.String("2020-01-01"),
					EndedOn: proto.String("2024-01-01"), Achievements: proto.String("Results")},
				{IsCurrent: true},
			},
			EducationItems: []models.EducationItem{
				{RecordType: proto.String("degree"), Institution: proto.String("University"), CompletionYear: &year, Specialization: proto.String("CS")},
				{},
			},
			Skills:   []models.NamedItem{{Name: "Go"}, {Name: "SQL"}},
			Projects: []models.ProjectItem{{Title: proto.String("Tool"), URL: proto.String("https://example.org"), Description: proto.String("Project")}},
		},
	}
	h := NewHandler(profileStub{profile: func(_ context.Context, id int) (models.SeekerProfile, error) {
		if id != 42 {
			t.Fatalf("use case received id %d", id)
		}
		return model, nil
	}})
	want := &seekerv1.SeekerProfile{
		Profile: &seekerv1.SeekerProfile_Profile{
			UserId: 42, Description: "Developer", CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(updated),
		},
		PersonalInfo: &seekerv1.SeekerProfile_PersonalInfo{
			FirstName: "Anna", LastName: "Ivanova", City: proto.String(""), AvatarUrl: proto.String("/api/seekers/42/avatar"),
		},
		Contacts: &seekerv1.SeekerProfile_Contacts{
			Phone: "1234", Messenger: proto.String("@anna"), ContactEmail: proto.String("anna@example.org"),
		},
		Preferences: &seekerv1.SeekerProfile_Preferences{
			PreferredPosition: proto.String("Backend"), PreferredCities: proto.String("Moscow"), WorkFormat: proto.String("remote"),
		},
		Career: &seekerv1.CareerDetails{
			ExperienceItems: []*seekerv1.ExperienceItem{
				{CompanyName: proto.String("Example"), PositionTitle: proto.String("Engineer"), StartedOn: proto.String("2020-01-01"),
					EndedOn: proto.String("2024-01-01"), Achievements: proto.String("Results")},
				{IsCurrent: true},
			},
			EducationItems: []*seekerv1.EducationItem{
				{RecordType: proto.String("degree"), Institution: proto.String("University"), CompletionYear: proto.Int32(2020), Specialization: proto.String("CS")},
				{},
			},
			Skills:   []*seekerv1.Skill{{Name: "Go"}, {Name: "SQL"}},
			Projects: []*seekerv1.ProjectItem{{Title: proto.String("Tool"), Url: proto.String("https://example.org"), Description: proto.String("Project")}},
		},
	}
	req := &seekerv1.GetSeekerProfileRequest{SeekerId: 42}
	got, err := h.GetSeekerProfile(t.Context(), req)
	if err != nil || !proto.Equal(got.GetProfile(), want) {
		t.Fatalf("GetSeekerProfile() = %v, %v, want %v", got, err, want)
	}

	for _, key := range []*string{nil, proto.String("")} {
		model.PersonalInfo.AvatarKey = key
		got, err = h.GetSeekerProfile(t.Context(), req)
		if err != nil || got.GetProfile().GetPersonalInfo().AvatarUrl != nil {
			t.Fatalf("missing avatar key: got %v, %v", got, err)
		}
	}
}

func TestCareerMapperKeepsSlicePresence(t *testing.T) {
	absent := toProtoCareer(models.CareerDetails{})
	if absent.ExperienceItems != nil || absent.EducationItems != nil || absent.Skills != nil || absent.Projects != nil {
		t.Fatalf("mapper populated absent slices: %+v", absent)
	}
	empty := toProtoCareer(models.CareerDetails{
		ExperienceItems: []models.ExperienceItem{},
		EducationItems:  []models.EducationItem{},
		Skills:          []models.NamedItem{},
		Projects:        []models.ProjectItem{},
	})
	if empty.ExperienceItems == nil || empty.EducationItems == nil || empty.Skills == nil || empty.Projects == nil {
		t.Fatalf("mapper dropped explicit empty slices: %+v", empty)
	}
}

func TestSeekerRPCInvalidSeekerID(t *testing.T) {
	h := NewHandler(profileStub{})
	for _, id := range []int64{0, -1} {
		_, profileErr := h.GetSeekerProfile(t.Context(), &seekerv1.GetSeekerProfileRequest{SeekerId: id})
		_, avatarErr := h.GetSeekerAvatar(t.Context(), &seekerv1.GetSeekerAvatarRequest{SeekerId: id})
		if status.Code(profileErr) != codes.InvalidArgument || status.Code(avatarErr) != codes.InvalidArgument {
			t.Fatalf("id %d: got %v and %v", id, profileErr, avatarErr)
		}
	}
}

func TestSeekerRPCDependencyErrors(t *testing.T) {
	for _, tc := range []struct {
		code codes.Code
		err  error
	}{
		{codes.NotFound, apperrors.ErrNotFound},
		{codes.Canceled, context.Canceled},
		{codes.DeadlineExceeded, context.DeadlineExceeded},
		{codes.Internal, errors.New("private storage error")},
	} {
		t.Run(tc.code.String(), func(t *testing.T) {
			wrapped := fmt.Errorf("dependency: %w", tc.err)
			h := NewHandler(profileStub{
				profile: func(context.Context, int) (models.SeekerProfile, error) {
					return models.SeekerProfile{}, wrapped
				},
				avatar: func(context.Context, int) (io.ReadCloser, string, error) { return nil, "", wrapped },
			})
			_, profileErr := h.GetSeekerProfile(t.Context(), &seekerv1.GetSeekerProfileRequest{SeekerId: 42})
			_, avatarErr := h.GetSeekerAvatar(t.Context(), &seekerv1.GetSeekerAvatarRequest{SeekerId: 42})
			for _, err := range []error{profileErr, avatarErr} {
				if status.Code(err) != tc.code || status.Convert(err).Message() == wrapped.Error() {
					t.Fatalf("got %v, want sanitized %v", err, tc.code)
				}
			}
		})
	}
}

type trackedBody struct {
	io.Reader
	closeErr error
	closes   int
}

func (b *trackedBody) Close() error {
	b.closes++
	return b.closeErr
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

func TestGetSeekerAvatar(t *testing.T) {
	fullLogo := bytes.Repeat([]byte{7}, maxAvatarSize+1)
	for _, tc := range []struct {
		name     string
		content  []byte
		readErr  error
		closeErr error
		code     codes.Code
	}{
		{name: "success", content: []byte("avatar"), code: codes.OK},
		{name: "at size limit", content: fullLogo[:maxAvatarSize], code: codes.OK},
		{name: "too large", content: fullLogo, code: codes.ResourceExhausted},
		{name: "read error", readErr: io.ErrUnexpectedEOF, code: codes.Internal},
		{name: "read canceled", readErr: context.Canceled, code: codes.Canceled},
		{name: "read deadline", readErr: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "close error", content: []byte("avatar"), closeErr: io.ErrClosedPipe, code: codes.Internal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: bytes.NewReader(tc.content), closeErr: tc.closeErr}
			if tc.readErr != nil {
				body.Reader = failedReader{tc.readErr}
			}
			h := NewHandler(profileStub{avatar: func(_ context.Context, id int) (io.ReadCloser, string, error) {
				if id != 42 {
					t.Fatalf("use case received id %d", id)
				}
				return body, "image/png", nil
			}})
			got, err := h.GetSeekerAvatar(t.Context(), &seekerv1.GetSeekerAvatarRequest{SeekerId: 42})
			if status.Code(err) != tc.code || body.closes != 1 {
				t.Fatalf("error = %v, close calls = %d, want %v and 1", err, body.closes, tc.code)
			}
			if tc.code == codes.OK && (!bytes.Equal(got.GetContent(), tc.content) || got.GetContentType() != "image/png") {
				t.Fatal("response content or type differs from storage")
			}
			if tc.code != codes.OK && got != nil {
				t.Fatal("RPC returned a response on error")
			}
		})
	}
}

func TestGetSeekerAvatarGRPCReceiveLimit(t *testing.T) {
	avatar := bytes.Repeat([]byte{7}, maxAvatarSize)
	h := NewHandler(profileStub{avatar: func(context.Context, int) (io.ReadCloser, string, error) {
		return io.NopCloser(bytes.NewReader(avatar)), "image/png", nil
	}})
	listener := bufconn.Listen(1 << 16)
	server := grpc.NewServer()
	seekerv1.RegisterSeekerServiceServer(server, h)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient("passthrough:///seeker-test",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	client := seekerv1.NewSeekerServiceClient(conn)
	req := &seekerv1.GetSeekerAvatarRequest{SeekerId: 42}
	if _, err := client.GetSeekerAvatar(ctx, req); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("default receive limit: got %v, want ResourceExhausted", err)
	}

	got, err := client.GetSeekerAvatar(ctx, req, grpc.MaxCallRecvMsgSize(11<<20))
	if err != nil || !bytes.Equal(got.GetContent(), avatar) || got.GetContentType() != "image/png" {
		t.Fatalf("receive limit with overhead: got %d bytes, content type %q, error %v",
			len(got.GetContent()), got.GetContentType(), err)
	}
}
