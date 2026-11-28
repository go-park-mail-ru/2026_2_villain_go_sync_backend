package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery"
	deliveryapplication "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/application"
	deliveryruntime "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/runtimeconfig"
	deliveryvacancy "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/delivery/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/application"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/runtimeconfig"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/usecase/vacancy"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
	"go.uber.org/mock/gomock"
)

const runtimePath = "/api/runtime-config"
const fixtureSecret = "test-only-snapshot-secret"

func productRequest(handler http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	}
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf"})
	req.Header.Set("X-CSRF-Token", "csrf")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func requireStatus(t *testing.T, res *httptest.ResponseRecorder, status int) {
	t.Helper()
	if res.Code != status {
		t.Fatalf("HTTP status %d, want %d; response = %s", res.Code, status, res.Body.String())
	}
}

func accessToken(t *testing.T, tokens *auth.TokenManager, id int64, role models.Role) string {
	t.Helper()
	token, err := tokens.Generate(id, role, auth.TokenTypeAccess)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

type frontendConfig struct {
	Configs map[string]int  `json:"configs"`
	Flags   map[string]bool `json:"flags"`
}

func readRuntimeConfig(t *testing.T, res *httptest.ResponseRecorder) frontendConfig {
	t.Helper()
	requireStatus(t, res, http.StatusOK)
	var raw struct {
		Data frontendConfig `json:"data"`
	}
	decoder := json.NewDecoder(res.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		t.Fatal(err)
	}
	return raw.Data
}

func TestRuntimeConfigUsesDefaultsWithoutOnlineConf(t *testing.T) {
	tokens := auth.NewTokenManager([]byte("test-jwt-key"), 0, 0)
	handler := routes(&delivery.Config{RuntimeConfig: deliveryruntime.NewHandler(runtimeconfig.Settings{})}, tokens)
	unauthorized := productRequest(handler, "GET", runtimePath, "", "")
	requireStatus(t, unauthorized, http.StatusUnauthorized)
	if unauthorized.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("unauthorized runtime response allowed caching")
	}

	token := accessToken(t, tokens, 27, models.RoleSeeker)
	res := productRequest(handler, "GET", runtimePath, "", token)
	if res.Header().Get("Cache-Control") != "private, no-store" || !strings.Contains(res.Header().Get("Vary"), "Cookie") {
		t.Fatal("authenticated runtime response allowed caching")
	}
	assertPublicConfig(t, readRuntimeConfig(t, res), 5, true, true)
}

func assertPublicConfig(t *testing.T, current frontendConfig, categories int, applications, vacancies bool) {
	t.Helper()
	expected := frontendConfig{
		Configs: map[string]int{"vacancy.max_categories": categories},
		Flags:   map[string]bool{"application.applications_enabled": applications, "vacancy.creation_enabled": vacancies},
	}
	if !reflect.DeepEqual(current, expected) {
		t.Fatalf("public runtime config = %+v, want %+v", current, expected)
	}
}

func TestOnlineConfUpdatesProductRoutesWithoutRestart(t *testing.T) {
	var allow atomic.Bool
	snapshotServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/onlineconf/snapshot" || r.Header.Get("Authorization") != "Bearer "+fixtureSecret {
			http.Error(w, "unauthorized", 401)
			return
		}
		limit := 2
		overrides := `[{"key_id":21,"user_id":27,"enabled":false},{"key_id":21,"user_id":43,"enabled":true}]`
		allowCreation := allow.Load()
		if allowCreation {
			limit = 4
			overrides = "[]"
		}
		fmt.Fprintf(w, `{"data":{
			"configs":[
				{"scope_name":"vacancy","name":"max_categories","value":%d},
				{"scope_name":"auth","name":"internal_secret","value":"sensitive-value"}
			],
			"flags":[
				{"id":20,"scope_name":"vacancy","name":"creation_enabled","enabled":%t,"rollout_percentage":100},
				{"id":21,"scope_name":"application","name":"applications_enabled","enabled":true,"rollout_percentage":100},
				{"id":22,"scope_name":"auth","name":"internal","enabled":true,"rollout_percentage":100}
			],
			"overrides":%s
		}}`, limit, allowCreation, overrides)
	}))
	defer snapshotServer.Close()
	client, err := onlineconf.NewClient(snapshotServer.URL, fixtureSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cache := onlineconf.StartCache(ctx, client)
	defer func() {
		cancel()
		<-cache.Done()
	}()
	settings := runtimeconfig.New(cache)

	ctrl := gomock.NewController(t)
	vacancies := vacancy.NewMockVacancyRepository(ctrl)
	employers := vacancy.NewMockUserRepository(ctrl)
	seekers := application.NewMockUserRepository(ctrl)
	applications := application.NewMockApplicationRepository(ctrl)
	resumes := application.NewMockResumeRepository(ctrl)
	applicationVacancies := application.NewMockVacancyRepository(ctrl)
	tokens := auth.NewTokenManager([]byte("test-jwt-key"), 0, 0)
	handler := routes(&delivery.Config{
		RuntimeConfig: deliveryruntime.NewHandler(settings),
		Vacancy:       deliveryvacancy.NewHandler(vacancy.NewVacancyUsecase(vacancies, employers, settings)),
		Application: deliveryapplication.NewHandler(application.NewApplicationUsecase(applications, resumes,
			applicationVacancies, seekers, settings)),
	}, tokens)

	const employerID int64 = 42
	const seekerID int64 = 27
	const vacancyID int64 = 7
	const resumeID = 11
	employerToken := accessToken(t, tokens, employerID, models.RoleEmployer)
	seekerToken := accessToken(t, tokens, seekerID, models.RoleSeeker)
	otherToken := accessToken(t, tokens, 43, models.RoleSeeker)

	assertPublicConfig(t, readRuntimeConfig(t, productRequest(handler, "GET", runtimePath+"?user_id=43", "", seekerToken)), 2, false, false)
	assertPublicConfig(t, readRuntimeConfig(t, productRequest(handler, "GET", runtimePath+"?user_id=27", "", otherToken)), 2, true, false)

	employers.EXPECT().GetRoleByID(gomock.Any(), employerID).Return(models.RoleEmployer, nil).Times(4)
	employers.EXPECT().HasEmployerProfile(gomock.Any(), employerID).Return(true, nil).Times(4)
	seekers.EXPECT().GetRoleByID(gomock.Any(), seekerID).Return(models.RoleSeeker, nil).Times(2)
	const vacancyBody = `{"title":"New vacancy","description":"Job","responsibilities":"Work","requirements":"Go","category_ids":[1,2,3]}`
	const applicationBody = `{"resume_id":11}`
	for _, res := range []*httptest.ResponseRecorder{
		productRequest(handler, "POST", "/api/vacancies", vacancyBody, employerToken),
		productRequest(handler, "POST", "/api/vacancies/7/applications?user_id=43", applicationBody, seekerToken),
	} {
		requireStatus(t, res, http.StatusForbidden)
		if !strings.Contains(res.Body.String(), `"error":"feature disabled"`) {
			t.Fatalf("disabled feature error = %s", res.Body.String())
		}
	}

	existing := models.Vacancy{
		ID: vacancyID, Version: 3, Title: "Existing", Company: models.VacancyCompany{EmployerID: employerID},
		Content: models.VacancyContent{Responsibilities: "Work"}, Requirements: models.VacancyRequirements{Description: "Go"},
		Salary: models.VacancySalary{Currency: models.CurrencyRUB}, Categories: []string{"a", "b", "c", "d", "e"},
	}
	vacancies.EXPECT().GetByID(gomock.Any(), vacancyID).Return(existing, nil).Times(3)
	vacancies.EXPECT().Update(gomock.Any(), gomock.Any(), (*[]int64)(nil)).Return(existing, nil)
	requireStatus(t, productRequest(handler, "GET", "/api/vacancies/7", "", seekerToken), http.StatusOK)
	requireStatus(t, productRequest(handler, "PATCH", "/api/vacancies/7", `{"title":"Renamed"}`, employerToken), http.StatusOK)
	requireStatus(t, productRequest(handler, "PATCH", "/api/vacancies/7", `{"category_ids":[1,2,3]}`, employerToken), http.StatusBadRequest)
	applications.EXPECT().ListBySeeker(gomock.Any(), seekerID, gomock.Any()).
		Return([]models.SeekerApplication{{ID: 9, SeekerID: seekerID}}, int64(1), nil)
	requireStatus(t, productRequest(handler, "GET", "/api/seeker/applications", "", seekerToken), http.StatusOK)

	allow.Store(true)
	deadline := time.NewTimer(onlineconf.RefreshInterval + 5*time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for settings.VacancyMaxCategories() != 4 {
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatal("product did not refresh OnlineConf")
		}
	}

	assertPublicConfig(t, readRuntimeConfig(t, productRequest(handler, "GET", runtimePath, "", seekerToken)), 4, true, true)
	vacancies.EXPECT().Create(gomock.Any(), gomock.Any(), []int64{1, 2, 3}).Return(existing, nil)
	requireStatus(t, productRequest(handler, "POST", "/api/vacancies", vacancyBody, employerToken), http.StatusCreated)
	resumes.EXPECT().GetByID(gomock.Any(), resumeID).Return(models.Resume{ID: resumeID, SeekerID: int(seekerID)}, nil)
	applicationVacancies.EXPECT().GetByID(gomock.Any(), vacancyID).Return(existing, nil)
	applications.EXPECT().Create(gomock.Any(), models.Application{
		SeekerID: seekerID, VacancyID: vacancyID, VacancyVersion: existing.Version, ResumeID: resumeID,
	}).Return(models.Application{ID: 10}, nil)
	requireStatus(t, productRequest(handler, "POST", "/api/vacancies/7/applications", applicationBody, seekerToken), http.StatusCreated)
}
