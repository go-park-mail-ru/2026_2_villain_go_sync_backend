package delivery_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/auth"
	productmodels "github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/delivery"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/productusers"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
	"github.com/jackc/pgx/v5/pgxpool"
)

func decodeData[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var data T
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode data: %v; %s", err, raw)
	}
	return data
}

type failureReader struct{}

func (failureReader) Exists(context.Context, int64) (bool, error) {
	return false, errors.New("private product PG connection failure")
}

func TestPostgresHTTP(t *testing.T) {
	dsn, productDSN := os.Getenv("ONLINECONF_TEST_DATABASE_URL"), os.Getenv("ONLINECONF_TEST_PRODUCT_READER_DATABASE_URL")
	if dsn == "" || productDSN == "" {
		t.Skip("set both OnlineConf HTTP test URLs for isolated, bootstrapped databases")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var valid bool
	err = pool.QueryRow(ctx, "SELECT current_database() = 'onlineconf' AND current_user = 'onlineconf_runtime'").Scan(&valid)
	if err != nil || !valid {
		t.Fatalf("expected an isolated runtime connection: %v", err)
	}
	users, err := productusers.Open(ctx, productDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer users.Close()
	if reader, err := productusers.Open(ctx, dsn); err == nil {
		reader.Close()
		t.Fatal("product reader accepted the OnlineConf runtime role")
	}

	tokens := auth.NewTokenManager([]byte(testJWT), 0, 0)
	owner := tokenFor(t, tokens, 1, productmodels.RoleSeeker, auth.TokenTypeAccess)
	adminUser := tokenFor(t, tokens, 2, productmodels.RoleEmployer, auth.TokenTypeAccess)
	ordinary := tokenFor(t, tokens, 3, productmodels.RoleEmployer, auth.TokenTypeAccess)
	rollback := errors.New("rollback HTTP fixtures")
	err = repository.NewStore(pool).WithinTx(ctx, func(tx *repository.Store) error {
		u := usecase.New(tx, users)
		h := newHTTPHandler(u, tokens, testSecret, users.Ping)
		snapshotServer := httptest.NewServer(h)
		defer snapshotServer.Close()
		snapshotClient, err := onlineconf.NewClient(snapshotServer.URL, testSecret)
		if err != nil {
			t.Fatal(err)
		}
		do := func(method, path, body, token string, status int) json.RawMessage {
			t.Helper()
			return request(t, h, method, admin+path, body, token, csrfToken, "", status)
		}
		asAdmin := func(method, path, body string, status int) json.RawMessage {
			t.Helper()
			return do(method, path, body, adminUser, status)
		}

		data := do("GET", "/me", "", owner, 200)
		if !strings.Contains(string(data), `"role":"owner"`) {
			t.Fatalf("owner /me = %s", data)
		}
		do("GET", "/me", "", ordinary, 403)
		do("GET", "/keys", "", ordinary, 403)
		do("POST", "/scopes", `{"name":"forbidden"}`, ordinary, 403)
		do("POST", "/admins", `{"user_id":99999,"role":"admin"}`, owner, 422)
		do("POST", "/admins", `{"user_id":2,"role":"admin"}`, owner, 201)
		do("GET", "/admins?limit=1", "", owner, 200)
		asAdmin("GET", "/me", "", 200)
		asAdmin("GET", "/admins", "", 403)
		asAdmin("POST", "/admins", `{"user_id":3,"role":"owner"}`, 403)
		asAdmin("PUT", "/admins/2", `{"role":"owner"}`, 403)
		asAdmin("DELETE", "/admins/1", "", 403)
		do("DELETE", "/admins/1", "", owner, 422)

		scope := decodeData[struct {
			ID int64 `json:"id"`
		}](t, asAdmin("POST", "/scopes", `{"name":"http_scope","description":"initial"}`, 201))
		scopePath := fmt.Sprintf("/scopes/%d", scope.ID)
		asAdmin("GET", "/scopes?limit=1", "", 200)
		asAdmin("GET", scopePath, "", 200)
		data = asAdmin("PUT", scopePath, `{"name":"http_renamed","description":null}`, 200)
		if !strings.Contains(string(data), `"description":null`) {
			t.Fatalf("scope description was not cleared: %s", data)
		}
		errBody := asAdmin("POST", "/scopes", `{"name":"http_renamed"}`, 409)
		if strings.Contains(string(errBody), "SQLSTATE") || strings.Contains(string(errBody), "scope_name") {
			t.Fatalf("SQL details leaked: %s", errBody)
		}
		asAdmin("POST", "/scopes", `{"name":"invalid.name"}`, 422)
		asAdmin("GET", "/scopes/99999", "", 404)
		asAdmin("GET", "/flags/99999/overrides", "", 404)

		configBody := fmt.Sprintf(`{"name":"sample_config","scope_id":%d,"owner_user_id":3,"value":null}`, scope.ID)
		asAdmin("POST", "/configs", strings.Replace(configBody, `"owner_user_id":3`, `"owner_user_id":99999`, 1), 422)
		failureHandler := newHTTPHandler(usecase.New(tx, failureReader{}), tokens, testSecret, users.Ping)
		errBody = request(t, failureHandler, "POST", admin+"/configs", configBody, adminUser, csrfToken, "", 503)
		if strings.Contains(string(errBody), "private") {
			t.Fatalf("product error leaked: %s", errBody)
		}

		config := decodeData[struct {
			ID    int64           `json:"id"`
			Value json.RawMessage `json:"value"`
		}](t, asAdmin("POST", "/configs", configBody, 201))
		if string(config.Value) != "null" {
			t.Fatalf("config null = %s", config.Value)
		}
		configPath := fmt.Sprintf("/configs/%d", config.ID)
		asAdmin("GET", configPath, "", 200)
		asAdmin("PUT", configPath, fmt.Sprintf(
			`{"name":"sample_config","scope_id":%d,"owner_user_id":2,"description":null,"value":[false,0]}`, scope.ID), 200)

		flag := decodeData[struct {
			ID int64 `json:"id"`
		}](t, asAdmin("POST", "/flags", fmt.Sprintf(
			`{"name":"sample_flag","scope_id":%d,"owner_user_id":3}`, scope.ID), 201))
		flagPath := fmt.Sprintf("/flags/%d", flag.ID)
		asAdmin("GET", flagPath, "", 200)
		assertPage := func(path, field, value string) *int64 {
			t.Helper()
			page := decodeData[struct {
				Items      []map[string]json.RawMessage `json:"items"`
				NextCursor *int64                       `json:"next_cursor"`
			}](t, asAdmin("GET", path, "", 200))
			if len(page.Items) != 1 || string(page.Items[0][field]) != value {
				t.Fatalf("page %s = %+v", path, page.Items)
			}
			return page.NextCursor
		}
		cursor := assertPage("/keys?search=HTTP_RENAMED.SAMPLE&limit=1", "value", "[false,0]")
		if cursor == nil || *cursor != config.ID {
			t.Fatalf("unexpected key cursor: %v", cursor)
		}
		assertPage(fmt.Sprintf("/keys?search=sample&scope_id=%d&type=feature_flag&cursor=%d", scope.ID, *cursor), "enabled", "false")
		assertPage("/keys?search=http_renamed.sample_flag", "rollout_percentage", "0")
		keyData := asAdmin("GET", fmt.Sprintf("/keys/%d", flag.ID), "", 200)
		if !strings.Contains(string(keyData), `"full_name":"http_renamed.sample_flag"`) {
			t.Fatalf("key detail = %s", keyData)
		}

		evaluate := func(wantReason string, wantEnabled, bucket bool) {
			t.Helper()
			result := decodeData[onlineconf.Evaluation](t, asAdmin("GET", flagPath+"/evaluate?user_id=3", "", 200))
			if string(result.Reason) != wantReason || result.Enabled != wantEnabled || (result.Bucket != nil) != bucket {
				t.Fatalf("evaluate result = %+v", result)
			}
			snapshot, err := snapshotClient.FetchSnapshot(ctx)
			if err != nil {
				t.Fatal(err)
			}
			productResult, ok := snapshot.Evaluate("http_renamed", "sample_flag", 3)
			if !ok || !reflect.DeepEqual(productResult, result) {
				t.Fatalf("product result = %+v, admin result = %+v", productResult, result)
			}
		}
		asAdmin("GET", flagPath+"/evaluate?user_id=99999", "", 422)
		asAdmin("PUT", flagPath+"/overrides/99999", `{"enabled":true}`, 422)
		asAdmin("PUT", flagPath+"/overrides/3", `{"enabled":true}`, 200)
		evaluate("globally_disabled", false, false)
		asAdmin("PUT", flagPath+"/enabled", `{"enabled":true}`, 200)
		evaluate("forced_include", true, false)
		asAdmin("PUT", flagPath+"/overrides/3", `{"enabled":false}`, 200)
		asAdmin("PUT", flagPath+"/rollout", `{"rollout_percentage":100}`, 422)
		evaluate("forced_exclude", false, false)
		assertPage(flagPath+"/overrides", "enabled", "false")
		assertPage("/users/3/overrides", "user_id", "3")
		asAdmin("DELETE", flagPath+"/overrides/3", "", 204)
		evaluate("rollout_excluded", false, true)
		verifyAdvancedHTTP(t, u, do, owner, adminUser, ordinary, scope.ID, flag.ID, config.ID)
		asAdmin("PUT", flagPath, fmt.Sprintf(
			`{"name":"sample_flag","scope_id":%d,"owner_user_id":3,"description":null,"enabled":false,"rollout_percentage":0}`, scope.ID), 200)
		evaluate("globally_disabled", false, false)

		snap := decodeData[map[string]json.RawMessage](t, request(t, h, "GET", snapshotPath, "", "", "", testSecret, 200))
		if len(snap) != 3 || snap["flags"] == nil || snap["configs"] == nil || string(snap["overrides"]) != "[]" {
			t.Fatalf("snapshot = %+v", snap)
		}
		request(t, h, "GET", "/health", "", "", "", "", 200)

		history := decodeData[struct {
			Items []struct {
				ID       int64           `json:"id"`
				Actor    int64           `json:"actor_user_id"`
				OldValue json.RawMessage `json:"old_value"`
				Resource string          `json:"resource_name"`
			} `json:"items"`
			NextCursor *int64 `json:"next_cursor"`
		}](t, asAdmin("GET", fmt.Sprintf("/history?resource_type=config&resource_id=%d&scope_name=http_renamed&limit=1", config.ID), "", 200))
		if len(history.Items) != 1 || history.Items[0].Actor != 2 || history.Items[0].Resource != "sample_config" ||
			history.NextCursor == nil || !strings.Contains(string(history.Items[0].OldValue), `"value":null`) {
			t.Fatalf("history = %+v", history)
		}
		asAdmin("GET", fmt.Sprintf("/history/%d", history.Items[0].ID), "", 200)

		asAdmin("DELETE", configPath, "", 204)
		asAdmin("GET", configPath, "", 404)
		asAdmin("DELETE", flagPath, "", 204)
		asAdmin("DELETE", scopePath, "", 204)
		do("PUT", "/admins/2", `{"role":"owner"}`, owner, 200)
		do("DELETE", "/admins/2", "", owner, 204)
		do("GET", "/me", "", adminUser, 403)
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("test transaction: %v", err)
	}
}

func verifyAdvancedHTTP(t *testing.T, u usecase.Usecases,
	do func(string, string, string, string, int) json.RawMessage, owner, admin, member string,
	scopeID, flagID, configID int64,
) {
	t.Helper()
	flagPath := fmt.Sprintf("/flags/%d", flagID)
	rolloutPath := flagPath + "/rollout"
	configPath := fmt.Sprintf("/configs/%d", configID)
	group := decodeData[models.Group](t, do("POST", "/groups", `{"name":"Reviewers","member_ids":[3]}`, owner, 201))
	groupPath := fmt.Sprintf("/groups/%d", group.ID)
	do("GET", "/grants", "", admin, 403)
	do("GET", "/me", "", member, 403)

	do("POST", "/grants", fmt.Sprintf(`{"group_id":%d,"scope_id":%d,"key_type":"feature_flag","action":"read"}`, group.ID, scopeID), owner, 201)
	me := decodeData[map[string]any](t, do("GET", "/me", "", member, 200))
	if me["role"] != "member" {
		t.Fatalf("delegated actor = %+v", me)
	}
	keys := decodeData[struct {
		Items []struct{ ID int64 }
	}](t, do("GET", "/keys?limit=1", "", member, 200))
	if len(keys.Items) != 1 || keys.Items[0].ID != flagID {
		t.Fatalf("delegated catalog = %+v", keys)
	}
	scopes := decodeData[struct {
		Items []struct{ ID int64 }
	}](t, do("GET", "/scopes?limit=1", "", member, 200))
	if len(scopes.Items) != 1 || scopes.Items[0].ID != scopeID {
		t.Fatalf("delegated scope catalog = %+v", scopes)
	}
	do("GET", "/scopes/1", "", member, 403)
	do("GET", fmt.Sprintf("/scopes/%d", scopeID), "", member, 200)
	do("GET", configPath, "", member, 403)
	do("GET", fmt.Sprintf("/keys/%d", configID), "", member, 403)
	do("GET", "/history", "", member, 403)
	do("PUT", flagPath+"/enabled", `{"enabled":false}`, member, 403)
	do("POST", "/grants", fmt.Sprintf(`{"user_id":3,"key_id":%d,"action":"read"}`, configID), owner, 201)
	do("GET", configPath, "", member, 200)
	do("DELETE", configPath, "", member, 403)
	do("POST", "/grants", fmt.Sprintf(`{"group_id":%d,"scope_id":%d,"key_type":"config","action":"create"}`, group.ID, scopeID), owner, 201)
	created := decodeData[struct{ ID int64 }](t, do("POST", "/configs",
		fmt.Sprintf(`{"name":"delegated","scope_id":%d,"owner_user_id":3,"value":5}`, scopeID), member, 201))
	createdPath := fmt.Sprintf("/configs/%d", created.ID)
	do("GET", createdPath, "", member, 403)
	do("POST", "/grants", fmt.Sprintf(`{"group_id":%d,"key_id":%d,"action":"update"}`, group.ID, created.ID), owner, 201)
	configBody := fmt.Sprintf(`{"name":"delegated","scope_id":%d,"owner_user_id":3,"description":null,"value":10}`, scopeID)
	do("PUT", createdPath, configBody, member, 200)
	do("PUT", createdPath, `{"name":"delegated","scope_id":1,"owner_user_id":3,"description":null,"value":10}`, member, 403)
	do("POST", "/grants", fmt.Sprintf(`{"user_id":3,"key_id":%d,"action":"delete"}`, created.ID), owner, 201)
	do("DELETE", createdPath, "", member, 204)

	do("POST", "/rollout-policies", `{"name":"Invalid","steps":[{"percentage":100,"wait_seconds":0}]}`, admin, 422)
	do("POST", "/rollout-policies", `{"name":"Invalid","steps":[{"percentage":25,"wait_seconds":0},{"percentage":100,"wait_seconds":0}]}`, admin, 422)
	policy := decodeData[models.RolloutPolicy](t, do("POST", "/rollout-policies",
		`{"name":"Validated policy","steps":[{"percentage":25,"wait_seconds":0},{"percentage":100,"wait_seconds":1}]}`, admin, 201))
	startBody := fmt.Sprintf(`{"policy_id":%d}`, policy.ID)
	do("POST", "/flags", fmt.Sprintf(`{"name":"skip","scope_id":%d,"owner_user_id":3,"enabled":true,"rollout_percentage":100}`, scopeID), admin, 422)
	do("PUT", flagPath, fmt.Sprintf(`{"name":"sample_flag","scope_id":%d,"owner_user_id":3,"description":null,"enabled":true,"rollout_percentage":100}`, scopeID), admin, 422)
	run := decodeData[models.Rollout](t, do("POST", rolloutPath, startBody, admin, 201))
	do("POST", rolloutPath, startBody, admin, 409)
	do("GET", rolloutPath, "", member, 200)
	confirmation := func(step int) string { return fmt.Sprintf(`{"run_id":%d,"step":%d}`, run.RunID, step) }
	do("POST", rolloutPath+"/approve", confirmation(0), member, 403)
	do("POST", "/grants", fmt.Sprintf(`{"group_id":%d,"key_id":%d,"action":"approve"}`, group.ID, flagID), owner, 201)

	process := func(want int16) {
		t.Helper()
		if err := u.Rollouts.ProcessReady(t.Context()); err != nil {
			t.Fatal(err)
		}
		flag := decodeData[struct {
			Percentage int16 `json:"rollout_percentage"`
		}](t, do("GET", flagPath, "", admin, 200))
		if flag.Percentage != want {
			t.Fatalf("flag rollout percentage = %d, want %d", flag.Percentage, want)
		}
	}
	process(0)
	do("POST", rolloutPath+"/approve", confirmation(0), member, 200)
	process(25)
	do("PATCH", flagPath, fmt.Sprintf(`{"name":"sample_flag","scope_id":%d,"owner_user_id":3,"description":null,"enabled":true}`, scopeID), admin, 200)
	do("POST", rolloutPath+"/approve", confirmation(0), member, 409)
	second := decodeData[models.Rollout](t, do("POST", rolloutPath+"/approve", confirmation(1), member, 200))
	process(25)
	do("PUT", groupPath, `{"name":"Reviewers","member_ids":[]}`, owner, 422)
	savedGroup := decodeData[models.Group](t, do("PUT", groupPath,
		fmt.Sprintf(`{"name":"Reviewers","member_ids":[],"revision":%d}`, group.Revision), owner, 200))
	if savedGroup.Revision != group.Revision+1 {
		t.Fatalf("group update revision = %d, previous = %d", savedGroup.Revision, group.Revision)
	}
	do("PUT", groupPath, fmt.Sprintf(`{"name":"Reviewers","member_ids":[3],"revision":%d}`, group.Revision), owner, 409)
	do("GET", flagPath, "", member, 403)
	do("GET", configPath, "", member, 200)
	do("GET", fmt.Sprintf("/scopes/%d", scopeID), "", member, 200)
	time.Sleep(time.Until(*second.DueAt) + 20*time.Millisecond)
	process(25)
	revoked := decodeData[models.Rollout](t, do("GET", rolloutPath, "", admin, 200))
	if revoked.ApprovedBy != nil {
		t.Fatal("scheduler retained revoked approval")
	}
	do("POST", rolloutPath+"/approve", confirmation(1), admin, 200)
	process(100)
	complete := decodeData[models.Rollout](t, do("GET", rolloutPath, "", admin, 200))
	if complete.Status != models.RolloutCompleted || complete.NextStep != 2 {
		t.Fatalf("completed rollout = %+v", complete)
	}
	do("PUT", flagPath+"/enabled", `{"enabled":false}`, admin, 200)
	do("PUT", flagPath+"/enabled", `{"enabled":true}`, admin, 200)
	process(0)
	restarted := decodeData[models.Rollout](t, do("POST", rolloutPath, startBody, admin, 201))
	do("POST", rolloutPath+"/approve", confirmation(0), admin, 409)
	do("POST", rolloutPath+"/cancel", fmt.Sprintf(`{"run_id":%d}`, restarted.RunID), admin, 200)
	process(0)
	do("DELETE", groupPath, "", owner, 204)
}

func newHTTPHandler(u usecase.Usecases, tokens *auth.TokenManager, secret string, health func(context.Context) error) http.Handler {
	return delivery.New(delivery.Services{
		Scopes: u.Scopes, Configs: u.Configs, Flags: u.Flags, Overrides: u.Overrides,
		Admins: u.Admins, History: u.History, Keys: u.Keys, Snapshots: u.Snapshots,
		Access: u.Access, Rollouts: u.Rollouts,
	}, tokens, secret, health)
}
