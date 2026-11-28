package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var rollbackFixture = errors.New("rollback repository test fixture")

func openStore(t *testing.T) (context.Context, *repository.Store) {
	t.Helper()
	dsn := os.Getenv("ONLINECONF_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set ONLINECONF_TEST_DATABASE_URL to an isolated, migrated OnlineConf database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	var valid bool
	err = pool.QueryRow(ctx, "SELECT current_database() = 'onlineconf' AND current_user = 'onlineconf_runtime'").Scan(&valid)
	if err != nil || !valid {
		t.Fatalf("test connection must use onlineconf_runtime on a test onlineconf database: %v", err)
	}
	return ctx, repository.NewStore(pool)
}

func withFixture(t *testing.T, fn func(context.Context, *repository.Store)) {
	t.Helper()
	ctx, store := openStore(t)
	err := store.WithinTx(ctx, func(tx *repository.Store) error {
		fn(ctx, tx)
		return rollbackFixture
	})
	requireError(t, err, rollbackFixture)
}

func requireError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
}

func scopeInput(name string) models.Scope {
	return models.Scope{Name: name}
}

func keyInput(scope models.Scope, name string) models.ConfigKey {
	return models.ConfigKey{ScopeID: scope.ID, Name: name, OwnerUserID: 1}
}

func addScope(t *testing.T, ctx context.Context, store *repository.Store, name string) models.Scope {
	t.Helper()
	scope, err := store.Scopes.Create(ctx, scopeInput(name))
	requireError(t, err, nil)
	return scope
}

func addConfig(t *testing.T, ctx context.Context, store *repository.Store, scope models.Scope, name, value string) models.Config {
	t.Helper()
	result, err := store.Configs.Create(ctx, models.Config{ConfigKey: keyInput(scope, name), Value: json.RawMessage(value)})
	requireError(t, err, nil)
	return result
}

func addFlag(t *testing.T, ctx context.Context, store *repository.Store, scope models.Scope, name string) models.FeatureFlag {
	t.Helper()
	flag, err := store.Flags.Create(ctx, models.FeatureFlag{ConfigKey: keyInput(scope, name)})
	requireError(t, err, nil)
	return flag
}

func assertJSON(t *testing.T, actual json.RawMessage, expected string) {
	t.Helper()
	var got, want any
	requireError(t, json.Unmarshal(actual, &got), nil)
	requireError(t, json.Unmarshal([]byte(expected), &want), nil)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON = %s, want %s", actual, expected)
	}
}

func TestPostgresScopeAndConfigCRUD(t *testing.T) {
	withFixture(t, func(ctx context.Context, store *repository.Store) {
		scope := addScope(t, ctx, store, "repository_config")
		desc := "updated scope"
		scope.Description = &desc
		updatedScope, err := store.Scopes.Update(ctx, scope)
		requireError(t, err, nil)
		if updatedScope.Description == nil || *updatedScope.Description != desc || updatedScope.UpdatedAt.Before(scope.UpdatedAt) {
			t.Fatalf("updated scope = %+v", updatedScope)
		}
		gotScope, err := store.Scopes.Get(ctx, scope.ID)
		requireError(t, err, nil)
		scopeByName, err := store.Scopes.GetByName(ctx, scope.Name)
		requireError(t, err, nil)
		if gotScope.ID != scope.ID || scopeByName.ID != scope.ID {
			t.Fatal("scope lookup returned the wrong ID")
		}
		page, err := store.Scopes.List(ctx, repository.PageRequest{Limit: 2})
		requireError(t, err, nil)
		if len(page.Items) != 2 || page.Items[0].ID >= page.Items[1].ID || page.NextCursor == nil {
			t.Fatalf("scope page = %+v", page)
		}

		config := addConfig(t, ctx, store, scope, "json_value", "null")
		if config.Type != models.KeyTypeConfig || config.ScopeName != scope.Name || config.Value == nil {
			t.Fatalf("created config = %+v", config)
		}
		for _, value := range []string{`"text"`, `1.25`, `true`, `{"item":1}`, `[1,false]`, `null`} {
			config.Value = json.RawMessage(value)
			updated, err := store.Configs.Update(ctx, config)
			requireError(t, err, nil)
			assertJSON(t, updated.Value, value)
			got, err := store.Configs.Get(ctx, config.ID)
			requireError(t, err, nil)
			assertJSON(t, got.Value, value)
		}
		config.Value = nil
		_, err = store.Configs.Update(ctx, config)
		requireError(t, err, apperrors.ErrInvalidInput)

		requireError(t, store.Configs.Delete(ctx, config.ID), nil)
		_, err = store.Configs.Get(ctx, config.ID)
		requireError(t, err, apperrors.ErrNotFound)
		requireError(t, store.Scopes.Delete(ctx, scope.ID), nil)
		requireError(t, store.Scopes.Delete(ctx, scope.ID), apperrors.ErrNotFound)
	})
}

func TestPostgresKeySearchFiltersAndCursor(t *testing.T) {
	withFixture(t, func(ctx context.Context, store *repository.Store) {
		first := addScope(t, ctx, store, "repository_search")
		second := addScope(t, ctx, store, "repository_search_two")
		config := addConfig(t, ctx, store, first, "alpha_beta", "1")
		flag := addFlag(t, ctx, store, first, "alpha_flag")
		other := addConfig(t, ctx, store, second, "alpha_beta", "2")
		addConfig(t, ctx, store, first, "alphaxbeta", "3")

		_, err := store.Flags.Create(ctx, models.FeatureFlag{ConfigKey: keyInput(first, "alpha_beta")})
		requireError(t, err, apperrors.ErrConflict)
		got, err := store.Keys.GetByName(ctx, second.Name, other.Name)
		requireError(t, err, nil)
		if got.ID != other.ID {
			t.Fatalf("qualified key = %+v", got)
		}
		got, err = store.Keys.Get(ctx, config.ID)
		requireError(t, err, nil)
		if got.ID != config.ID || got.ScopeName != first.Name {
			t.Fatalf("key by ID = %+v", got)
		}

		filter := repository.KeyFilter{ScopeID: &first.ID, Search: "ALPHA_", Page: repository.PageRequest{Limit: 1}}
		page, err := store.Keys.List(ctx, filter)
		requireError(t, err, nil)
		if len(page.Items) != 1 || page.Items[0].ID != config.ID || page.NextCursor == nil {
			t.Fatalf("first key page = %+v", page)
		}
		filter.Page.Cursor = page.NextCursor
		page, err = store.Keys.List(ctx, filter)
		requireError(t, err, nil)
		if len(page.Items) != 1 || page.Items[0].ID != flag.ID || page.NextCursor != nil {
			t.Fatalf("second key page = %+v", page)
		}
		flagType := models.KeyTypeFeatureFlag
		filter.Type = &flagType
		filter.Page.Cursor = nil
		page, err = store.Keys.List(ctx, filter)
		requireError(t, err, nil)
		if len(page.Items) != 1 || page.Items[0].ID != flag.ID {
			t.Fatalf("type filter = %+v", page)
		}

		filter.Page.Limit = 101
		_, err = store.Keys.List(ctx, filter)
		requireError(t, err, apperrors.ErrInvalidInput)
	})
}

func TestPostgresFlagOverrideAndAdminCRUD(t *testing.T) {
	withFixture(t, func(ctx context.Context, store *repository.Store) {
		scope := addScope(t, ctx, store, "repository_flags")
		flag := addFlag(t, ctx, store, scope, "rollout")
		if flag.Enabled || flag.RolloutPercentage != 0 || flag.Type != models.KeyTypeFeatureFlag {
			t.Fatalf("flag defaults = %+v", flag)
		}
		flag.Enabled, flag.RolloutPercentage = true, 100
		flag, err := store.Flags.Update(ctx, flag)
		requireError(t, err, nil)
		if !flag.Enabled || flag.RolloutPercentage != 100 {
			t.Fatalf("updated flag = %+v", flag)
		}
		_, err = store.Configs.Get(ctx, flag.ID)
		requireError(t, err, apperrors.ErrNotFound)

		for _, userID := range []int64{6, 5} {
			_, err := store.Overrides.Upsert(ctx, models.FeatureFlagOverride{KeyID: flag.ID, UserID: userID, Enabled: true})
			requireError(t, err, nil)
		}
		override, err := store.Overrides.Upsert(ctx, models.FeatureFlagOverride{KeyID: flag.ID, UserID: 5})
		requireError(t, err, nil)
		if override.Enabled || override.UserID != 5 {
			t.Fatalf("updated override = %+v", override)
		}
		byFlag, err := store.Overrides.ListByFlag(ctx, flag.ID, repository.PageRequest{Limit: 1})
		requireError(t, err, nil)
		if len(byFlag.Items) != 1 || byFlag.Items[0].UserID != 5 || byFlag.NextCursor == nil {
			t.Fatalf("flag overrides = %+v", byFlag)
		}
		next, err := store.Overrides.ListByFlag(ctx, flag.ID, repository.PageRequest{Cursor: byFlag.NextCursor})
		requireError(t, err, nil)
		if len(next.Items) != 1 || next.Items[0].UserID != 6 {
			t.Fatalf("next override page = %+v", next)
		}
		byUser, err := store.Overrides.ListByUser(ctx, 5, repository.PageRequest{})
		requireError(t, err, nil)
		if len(byUser.Items) != 1 || byUser.Items[0].KeyID != flag.ID {
			t.Fatalf("user overrides = %+v", byUser)
		}
		requireError(t, store.Overrides.Delete(ctx, flag.ID, 6), nil)
		requireError(t, store.Flags.Delete(ctx, flag.ID), nil)
		_, err = store.Overrides.Get(ctx, flag.ID, 5)
		requireError(t, err, apperrors.ErrNotFound)

		owner, err := store.Admins.Create(ctx, models.Admin{UserID: 1, Role: models.AdminRoleOwner})
		requireError(t, err, nil)
		child, err := store.Admins.Create(ctx, models.Admin{UserID: 2, Role: models.AdminRoleAdmin, GrantedBy: &owner.UserID})
		requireError(t, err, nil)
		child, err = store.Admins.UpdateRole(ctx, child.UserID, models.AdminRoleOwner, nil)
		requireError(t, err, nil)
		if child.Role != models.AdminRoleOwner || child.GrantedBy == nil || *child.GrantedBy != owner.UserID {
			t.Fatalf("updated admin = %+v", child)
		}
		list, err := store.Admins.List(ctx, repository.PageRequest{})
		requireError(t, err, nil)
		if len(list.Items) != 2 || list.Items[0].UserID != owner.UserID {
			t.Fatalf("admin list = %+v", list)
		}
		requireError(t, store.Admins.Delete(ctx, owner.UserID), nil)
		got, err := store.Admins.Get(ctx, child.UserID)
		requireError(t, err, nil)
		if got.Role != models.AdminRoleOwner {
			t.Fatalf("admin lookup = %+v", got)
		}
	})
}

func TestPostgresAtomicChangesAndHistory(t *testing.T) {
	ctx, store := openStore(t)
	scope := addScope(t, ctx, store, fmt.Sprintf("repository_atomic_%x", time.Now().UnixNano()))
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		requireError(t, store.Scopes.Delete(cleanupCtx, scope.ID), nil)
	})

	_, err := store.Flags.Create(ctx, models.FeatureFlag{ConfigKey: keyInput(scope, "invalid_flag"), RolloutPercentage: 101})
	requireError(t, err, apperrors.ErrInvalidInput)
	_, err = store.Keys.GetByName(ctx, scope.Name, "invalid_flag")
	requireError(t, err, apperrors.ErrNotFound)

	err = store.WithinTx(ctx, func(tx *repository.Store) error {
		config := addConfig(t, ctx, tx, scope, "rollback_config", "null")
		_, err := tx.History.Append(ctx, models.ChangeHistory{
			ActorUserID: 1, ResourceType: models.ResourceTypeConfig, ResourceID: config.ID,
			ResourceName: config.Name, Action: "invalid", NewValue: config.Value,
		})
		return err
	})
	requireError(t, err, apperrors.ErrInvalidInput)
	_, err = store.Keys.GetByName(ctx, scope.Name, "rollback_config")
	requireError(t, err, apperrors.ErrNotFound)

	err = store.WithinTx(ctx, func(tx *repository.Store) error {
		config := addConfig(t, ctx, tx, scope, "audit_config", "null")
		first, err := tx.History.Append(ctx, models.ChangeHistory{
			ActorUserID: 1, ResourceType: models.ResourceTypeConfig, ResourceID: config.ID,
			ScopeName: &scope.Name, ResourceName: config.Name, Action: models.ActionCreate, NewValue: config.Value,
		})
		requireError(t, err, nil)
		if first.OldValue != nil || string(first.NewValue) != "null" {
			t.Fatalf("SQL NULL and JSON null were not distinct: %+v", first)
		}
		requireError(t, tx.Configs.Delete(ctx, config.ID), nil)
		last, err := tx.History.Append(ctx, models.ChangeHistory{
			ActorUserID: 1, ResourceType: models.ResourceTypeConfig, ResourceID: config.ID,
			ScopeName: &scope.Name, ResourceName: config.Name, Action: models.ActionDelete, OldValue: config.Value,
		})
		requireError(t, err, nil)
		get, err := tx.History.Get(ctx, first.ID)
		requireError(t, err, nil)
		if get.ID != first.ID || string(get.NewValue) != "null" || get.OldValue != nil {
			t.Fatalf("history lookup = %+v", get)
		}
		resourceType := models.ResourceTypeConfig
		filter := repository.HistoryFilter{
			ResourceType: &resourceType, ResourceID: &config.ID, ScopeName: &scope.Name, Page: repository.PageRequest{Limit: 1},
		}
		page, err := tx.History.List(ctx, filter)
		requireError(t, err, nil)
		if len(page.Items) != 1 || page.Items[0].ID != last.ID || page.NextCursor == nil {
			t.Fatalf("latest history = %+v", page)
		}
		filter.Page.Cursor = page.NextCursor
		page, err = tx.History.List(ctx, filter)
		requireError(t, err, nil)
		if len(page.Items) != 1 || page.Items[0].ID != first.ID || page.NextCursor != nil {
			t.Fatalf("older history = %+v", page)
		}
		return rollbackFixture
	})
	requireError(t, err, rollbackFixture)
	history, err := store.History.List(ctx, repository.HistoryFilter{ScopeName: &scope.Name})
	requireError(t, err, nil)
	if len(history.Items) != 0 {
		t.Fatalf("history escaped rollback: %+v", history)
	}

	var committedConfig models.Config
	var committedChange models.ChangeHistory
	err = store.WithinTx(ctx, func(tx *repository.Store) error {
		committedConfig = addConfig(t, ctx, tx, scope, "committed_config", "null")
		var err error
		committedChange, err = tx.History.Append(ctx, models.ChangeHistory{
			ActorUserID: 1, ResourceType: models.ResourceTypeConfig, ResourceID: committedConfig.ID,
			ScopeName: &scope.Name, ResourceName: committedConfig.Name, Action: models.ActionCreate,
			NewValue: committedConfig.Value,
		})
		return err
	})
	requireError(t, err, nil)
	storedConfig, err := store.Configs.Get(ctx, committedConfig.ID)
	requireError(t, err, nil)
	assertJSON(t, storedConfig.Value, "null")

	requireError(t, store.Configs.Delete(ctx, committedConfig.ID), nil)
	storedChange, err := store.History.Get(ctx, committedChange.ID)
	requireError(t, err, nil)
	if storedChange.ResourceID != committedConfig.ID || string(storedChange.NewValue) != "null" {
		t.Fatalf("change did not survive the resource deletion: %+v", storedChange)
	}
}

func TestPostgresSnapshotIsCompleteAndConsistent(t *testing.T) {
	ctx, store := openStore(t)
	scope := addScope(t, ctx, store, "repository_snapshot")
	config := addConfig(t, ctx, store, scope, "setting", "null")
	flag := addFlag(t, ctx, store, scope, "switch")
	emptyFlag := addFlag(t, ctx, store, scope, "empty_switch")
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		requireError(t, store.Configs.Delete(cleanupCtx, config.ID), nil)
		requireError(t, store.Flags.Delete(cleanupCtx, flag.ID), nil)
		requireError(t, store.Flags.Delete(cleanupCtx, emptyFlag.ID), nil)
		requireError(t, store.Scopes.Delete(cleanupCtx, scope.ID), nil)
	})
	check := func(snapshot repository.Snapshot, enabled bool, value string, count int) {
		t.Helper()
		var gotConfig *models.Config
		var flags []models.FeatureFlag
		var overrides []models.FeatureFlagOverride
		for i, item := range snapshot.Configs {
			if item.ID == config.ID {
				gotConfig = &snapshot.Configs[i]
			}
		}
		for _, item := range snapshot.Flags {
			if item.ScopeID == scope.ID {
				flags = append(flags, item)
			}
		}
		for _, item := range snapshot.Overrides {
			if item.KeyID == flag.ID {
				overrides = append(overrides, item)
			}
		}
		if gotConfig == nil || len(flags) != 2 || len(overrides) != count {
			t.Fatalf("snapshot counts: config %v, flags %d, overrides %d", gotConfig, len(flags), len(overrides))
		}
		assertJSON(t, gotConfig.Value, value)
		if flags[0].ID != flag.ID || flags[0].Enabled != enabled || flags[1].ID != emptyFlag.ID {
			t.Fatalf("snapshot flags = %+v", flags)
		}
		for i, override := range overrides {
			if override.UserID != int64(i+1) || override.Enabled != (i%2 == 0) {
				t.Fatalf("snapshot override %d = %+v", i, override)
			}
		}
	}
	err := store.WithinTx(ctx, func(tx *repository.Store) error {
		config.Value = json.RawMessage("true")
		if _, err := tx.Configs.Update(ctx, config); err != nil {
			return err
		}
		flag.Enabled = true
		if _, err := tx.Flags.Update(ctx, flag); err != nil {
			return err
		}
		for id := int64(1); id <= 105; id++ {
			_, err := tx.Overrides.Upsert(ctx, models.FeatureFlagOverride{KeyID: flag.ID, UserID: id, Enabled: id%2 == 1})
			if err != nil {
				return fmt.Errorf("set override %d: %w", id, err)
			}
		}
		outside, err := store.Snapshots.Get(ctx)
		requireError(t, err, nil)
		check(outside, false, "null", 0)
		inside, err := tx.Snapshots.Get(ctx)
		requireError(t, err, nil)
		check(inside, true, "true", 105)
		return nil
	})
	requireError(t, err, nil)
	committed, err := store.Snapshots.Get(ctx)
	requireError(t, err, nil)
	check(committed, true, "true", 105)
}
