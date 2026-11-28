package delivery_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	productmodels "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
)

const admin = "/api/onlineconf/admin"
const snapshotPath = "/internal/onlineconf/snapshot"
const testSecret = "test-only-snapshot-secret"
const testJWT = "test-only-jwt-key"
const csrfToken = "csrf-for-test"

func tokenFor(t *testing.T, tokens *auth.TokenManager, id int64, role productmodels.Role, typ auth.TokenType) string {
	t.Helper()
	token, err := tokens.Generate(id, role, typ)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func request(t *testing.T, handler http.Handler, method, path, body, token, csrfHeader, bearer string, want int) json.RawMessage {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken})
	}
	if csrfHeader != "" {
		req.Header.Set("X-CSRF-Token", csrfHeader)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s = %d, want %d; %s", method, path, rec.Code, want, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") == "" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("%s %s missing response headers", method, path)
	}
	if want == http.StatusNoContent {
		if rec.Body.Len() != 0 {
			t.Fatalf("%s %s: 204 returned a body", method, path)
		}
		return nil
	}
	var envelope struct {
		Data  json.RawMessage `json:"data"`
		Error string          `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if want >= 400 {
		if envelope.Error == "" || envelope.Data != nil {
			t.Fatalf("expected error envelope, got %s", rec.Body.String())
		}
		return json.RawMessage(envelope.Error)
	}
	if envelope.Data == nil || envelope.Error != "" {
		t.Fatalf("expected success envelope, got %s", rec.Body.String())
	}
	return envelope.Data
}

func TestJWTAndCSRF(t *testing.T) {
	tokens := auth.NewTokenManager([]byte(testJWT), 0, 0)
	h := newHTTPHandler(usecase.Usecases{}, tokens, testSecret, func(context.Context) error { return nil })
	access := tokenFor(t, tokens, 1, productmodels.RoleSeeker, auth.TokenTypeAccess)
	refresh := tokenFor(t, tokens, 1, productmodels.RoleSeeker, auth.TokenTypeRefresh)
	expired := tokenFor(t, auth.NewTokenManager([]byte(testJWT), time.Nanosecond, 0), 1, productmodels.RoleSeeker, auth.TokenTypeAccess)
	for _, token := range []string{"", "invalid.jwt", refresh, expired} {
		request(t, h, "GET", admin+"/me", "", token, "", "", 401)
	}
	request(t, h, "GET", admin+"/me", "", "", "", access, 401)
	request(t, h, "POST", admin+"/scopes", `{"name":"scope"}`, "", "", "", 401)
	request(t, h, "POST", admin+"/scopes", `{"name":"scope"}`, access, "", "", 403)
	request(t, h, "POST", admin+"/scopes", `{"name":"scope"}`, access, "wrong", "", 403)
	request(t, h, "GET", snapshotPath, "", access, "", "", 401)
	request(t, h, "GET", snapshotPath, "", "", "", access, 401)
	request(t, h, "GET", snapshotPath, "", "", "", "incorrect", 401)
}

func TestRequestValidationAndHealth(t *testing.T) {
	tokens := auth.NewTokenManager([]byte(testJWT), 0, 0)
	ok := false
	h := newHTTPHandler(usecase.Usecases{}, tokens, testSecret, func(context.Context) error {
		if !ok {
			return errors.New("private connection details")
		}
		return nil
	})
	access := tokenFor(t, tokens, 1, productmodels.RoleEmployer, auth.TokenTypeAccess)
	err := request(t, h, "GET", "/health", "", "", "", "", 503)
	if strings.Contains(string(err), "private") {
		t.Fatal("dependency error leaked")
	}
	ok = true
	request(t, h, "GET", "/health", "", "", "", "", 200)
	for _, body := range []string{`null`, `{"name":"scope","id":5}`, `{"name":"scope","actor_user_id":4}`, `{} {}`, `{"name":`} {
		request(t, h, "POST", admin+"/scopes", body, access, csrfToken, "", 400)
	}
	request(t, h, "GET", admin+"/scopes/0", "", access, "", "", 400)
	request(t, h, "GET", admin+"/keys?cursor=bad", "", access, "", "", 400)
	request(t, h, "GET", admin+"/keys?limit=101", "", access, "", "", 400)
	request(t, h, "GET", admin+"/keys?type=wrong", "", access, "", "", 400)
	request(t, h, "GET", admin+"/history?resource_type=wrong", "", access, "", "", 400)
	request(t, h, "PUT", admin+"/scopes/1", `{"name":"scope"}`, access, csrfToken, "", 422)
	request(t, h, "POST", admin+"/configs", `{"scope_id":1,"name":"x","owner_user_id":1}`, access, csrfToken, "", 422)
	request(t, h, "PUT", admin+"/flags/1", `{"scope_id":1,"name":"x","owner_user_id":1,"description":null,"enabled":false}`, access, csrfToken, "", 422)
	request(t, h, "PUT", admin+"/flags/1/enabled", `{"enabled":null}`, access, csrfToken, "", 422)
	request(t, h, "PUT", admin+"/flags/1/rollout", `{"rollout_percentage":101}`, access, csrfToken, "", 422)
}

func TestReplacementValidationContract(t *testing.T) {
	tokens := auth.NewTokenManager([]byte(testJWT), 0, 0)
	h := newHTTPHandler(usecase.Usecases{}, tokens, testSecret, func(context.Context) error { return nil })
	access := tokenFor(t, tokens, 1, productmodels.RoleEmployer, auth.TokenTypeAccess)
	for _, tc := range []struct {
		path string
		body string
		code int
	}{
		{"/configs/1", `{"scope_id":1,"name":"x","owner_user_id":1,"description":null}`, 422},
		{"/flags/1", `{"scope_id":1,"name":"x","owner_user_id":1,"description":null,"enabled":false}`, 422},
		{"/flags/1/rollout", `{"rollout_percentage":0,"actor_user_id":1}`, 400},
		{"/flags/1/enabled", `{"enabled":null}`, 422},
	} {
		request(t, h, "PUT", admin+tc.path, tc.body, access, csrfToken, "", tc.code)
	}
}

func TestMalformedAndRepeatedQueryFilters(t *testing.T) {
	tokens := auth.NewTokenManager([]byte(testJWT), 0, 0)
	h := newHTTPHandler(usecase.Usecases{}, tokens, testSecret, func(context.Context) error { return nil })
	access := tokenFor(t, tokens, 1, productmodels.RoleEmployer, auth.TokenTypeAccess)
	for _, path := range []string{
		"/keys?limit=1&limit=2", "/keys?search=a&search=b", "/keys?search=%zz",
		"/history?resource_id=1&resource_id=2", "/flags/1/evaluate?user_id=1&user_id=2",
	} {
		request(t, h, "GET", admin+path, "", access, "", "", 400)
	}
}
