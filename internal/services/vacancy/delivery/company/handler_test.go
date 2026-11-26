package company

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	vacancyv1 "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/gen/go/vacancy/v1"
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

type companyStub struct {
	profile func(context.Context, int) (models.EmployerProfile, error)
	logo    func(context.Context, int) (io.ReadCloser, string, error)
}

func (s companyStub) GetEmployerByID(ctx context.Context, id int) (models.EmployerProfile, error) {
	return s.profile(ctx, id)
}

func (s companyStub) GetEmployerLogo(ctx context.Context, id int) (io.ReadCloser, string, error) {
	return s.logo(ctx, id)
}

func TestGetCompanyProfile(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 400, time.UTC)
	updated := created.Add(time.Hour)
	model := models.EmployerProfile{
		Profile: models.Profile{UserID: 42, Description: "Company description", CreatedAt: created, UpdatedAt: updated},
		Company: models.EmployerCompanyInfo{CompanyName: "Example", Industry: proto.String("IT"),
			TeamSize: proto.String("50"), Benefits: proto.String(""), LogoKey: proto.String("private/s3/key")},
		Contacts: models.EmployerContacts{Website: "https://example.org", HiringEmail: proto.String("jobs@example.org")},
		Location: models.EmployerLocation{MainCity: proto.String("Moscow"), OfficeAddress: proto.String("Office")},
		Details: models.CompanyDetails{
			Technologies: []models.NamedItem{{Name: "Go"}, {Name: "PostgreSQL"}},
			Culture: []models.CultureItem{
				{Title: proto.String("Team"), ShortDescription: proto.String("Short"), Description: proto.String("Long")},
				{Description: proto.String("")},
			},
		},
	}

	h := NewHandler(companyStub{profile: func(_ context.Context, id int) (models.EmployerProfile, error) {
		if id != 42 {
			t.Fatalf("use case received id %d", id)
		}
		return model, nil
	}})
	want := &vacancyv1.CompanyProfile{
		Profile: &vacancyv1.CompanyProfile_Profile{
			UserId: 42, Description: "Company description", CreatedAt: timestamppb.New(created), UpdatedAt: timestamppb.New(updated),
		},
		Company: &vacancyv1.CompanyProfile_Company{
			CompanyName: "Example", Industry: proto.String("IT"), TeamSize: proto.String("50"),
			Benefits: proto.String(""), LogoUrl: proto.String("/api/employers/42/logo"),
		},
		Contacts: &vacancyv1.CompanyProfile_Contacts{Website: "https://example.org", HiringEmail: proto.String("jobs@example.org")},
		Location: &vacancyv1.CompanyProfile_Location{MainCity: proto.String("Moscow"), OfficeAddress: proto.String("Office")},
		Details: &vacancyv1.CompanyProfile_Details{
			Technologies: []*vacancyv1.Technology{{Name: "Go"}, {Name: "PostgreSQL"}},
			Culture: []*vacancyv1.CultureItem{
				{Title: proto.String("Team"), ShortDescription: proto.String("Short"), Description: proto.String("Long")},
				{Description: proto.String("")},
			},
		},
	}
	req := &vacancyv1.GetCompanyProfileRequest{EmployerId: 42}
	got, err := h.GetCompanyProfile(t.Context(), req)
	if err != nil || !proto.Equal(got.GetProfile(), want) {
		t.Fatalf("GetCompanyProfile() = %v, %v, want %v", got, err, want)
	}

	model.Company.LogoKey = proto.String("")
	got, err = h.GetCompanyProfile(t.Context(), req)
	if err != nil || got.GetProfile().GetCompany().LogoUrl != nil {
		t.Fatalf("empty logo key: got %v, %v", got, err)
	}
}

func TestCompanyRPCInvalidEmployerID(t *testing.T) {
	h := NewHandler(companyStub{})
	for _, id := range []int64{0, -1} {
		_, profileErr := h.GetCompanyProfile(t.Context(), &vacancyv1.GetCompanyProfileRequest{EmployerId: id})
		_, logoErr := h.GetCompanyLogo(t.Context(), &vacancyv1.GetCompanyLogoRequest{EmployerId: id})
		if status.Code(profileErr) != codes.InvalidArgument || status.Code(logoErr) != codes.InvalidArgument {
			t.Fatalf("id %d: got %v and %v", id, profileErr, logoErr)
		}
	}
}

func TestCompanyRPCDependencyErrors(t *testing.T) {
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
			h := NewHandler(companyStub{
				profile: func(context.Context, int) (models.EmployerProfile, error) {
					return models.EmployerProfile{}, wrapped
				},
				logo: func(context.Context, int) (io.ReadCloser, string, error) { return nil, "", wrapped },
			})
			_, profileErr := h.GetCompanyProfile(t.Context(), &vacancyv1.GetCompanyProfileRequest{EmployerId: 42})
			_, logoErr := h.GetCompanyLogo(t.Context(), &vacancyv1.GetCompanyLogoRequest{EmployerId: 42})
			for _, err := range []error{profileErr, logoErr} {
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

func TestGetCompanyLogo(t *testing.T) {
	fullLogo := bytes.Repeat([]byte{7}, maxLogoSize+1)
	for _, tc := range []struct {
		name     string
		content  []byte
		readErr  error
		closeErr error
		code     codes.Code
	}{
		{name: "success", content: []byte("logo"), code: codes.OK},
		{name: "at size limit", content: fullLogo[:maxLogoSize], code: codes.OK},
		{name: "too large", content: fullLogo, code: codes.ResourceExhausted},
		{name: "read error", readErr: io.ErrUnexpectedEOF, code: codes.Internal},
		{name: "read canceled", readErr: context.Canceled, code: codes.Canceled},
		{name: "read deadline", readErr: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "close error", content: []byte("logo"), closeErr: io.ErrClosedPipe, code: codes.Internal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: bytes.NewReader(tc.content), closeErr: tc.closeErr}
			if tc.readErr != nil {
				body.Reader = failedReader{tc.readErr}
			}
			h := NewHandler(companyStub{logo: func(_ context.Context, id int) (io.ReadCloser, string, error) {
				if id != 42 {
					t.Fatalf("use case received id %d", id)
				}
				return body, "image/png", nil
			}})
			got, err := h.GetCompanyLogo(t.Context(), &vacancyv1.GetCompanyLogoRequest{EmployerId: 42})
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

func TestGetCompanyLogoGRPCReceiveLimit(t *testing.T) {
	logo := bytes.Repeat([]byte{7}, maxLogoSize)
	h := NewHandler(companyStub{logo: func(context.Context, int) (io.ReadCloser, string, error) {
		return io.NopCloser(bytes.NewReader(logo)), "image/png", nil
	}})
	listener := bufconn.Listen(1 << 16)
	server := grpc.NewServer()
	vacancyv1.RegisterCompanyServiceServer(server, h)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient("passthrough:///company-test",
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
	client := vacancyv1.NewCompanyServiceClient(conn)
	req := &vacancyv1.GetCompanyLogoRequest{EmployerId: 42}
	if _, err := client.GetCompanyLogo(ctx, req); status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("default receive limit: got %v, want ResourceExhausted", err)
	}

	got, err := client.GetCompanyLogo(ctx, req, grpc.MaxCallRecvMsgSize(11<<20))
	if err != nil || !bytes.Equal(got.GetContent(), logo) || got.GetContentType() != "image/png" {
		t.Fatalf("receive limit with overhead: got %d bytes, content type %q, error %v",
			len(got.GetContent()), got.GetContentType(), err)
	}
}
