package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/repository"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/usecase"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ownerID int64 = 1

func openTestStore(t *testing.T) (context.Context, *repository.Store, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("ONLINECONF_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set ONLINECONF_TEST_DATABASE_URL to an isolated, bootstrapped OnlineConf database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	t.Cleanup(cancel)
	config, err := pgxpool.ParseConfig(dsn)
	requireError(t, err, nil)
	config.ConnConfig.RuntimeParams["application_name"] = "onlineconf_usecase_test"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	requireError(t, err, nil)
	t.Cleanup(pool.Close)

	var valid bool
	err = pool.QueryRow(ctx, "SELECT current_database() = 'onlineconf' AND current_user = 'onlineconf_runtime'").Scan(&valid)
	requireError(t, err, nil)
	if !valid {
		t.Fatal("use onlineconf_runtime on an isolated onlineconf database")
	}
	store := repository.NewStore(pool)
	owner, err := store.Admins.Get(ctx, ownerID)
	requireError(t, err, nil)
	if owner.Role != models.AdminRoleOwner {
		t.Fatalf("expected bootstrap owner %d", ownerID)
	}
	return ctx, store, pool
}

func withFixture(t *testing.T, fn func(context.Context, *repository.Store, usecase.Usecases)) {
	t.Helper()
	ctx, store, _ := openTestStore(t)
	rollback := fmt.Errorf("rollback fixture")
	err := store.WithinTx(ctx, func(tx *repository.Store) error {
		fn(ctx, tx, usecase.New(tx, fixtureUsers{}))
		return rollback
	})
	requireError(t, err, rollback)
}

func readHistory(t *testing.T, ctx context.Context, u usecase.HistoryUsecase, actor int64,
	kind models.ResourceType, id int64,
) []models.ChangeHistory {
	t.Helper()
	page, err := u.List(ctx, actor, models.HistoryFilter{ResourceType: &kind, ResourceID: &id})
	requireError(t, err, nil)
	return page.Items
}

func TestScopeAndConfigHistoryIgnoresNoOps(t *testing.T) {
	withFixture(t, func(ctx context.Context, _ *repository.Store, u usecase.Usecases) {
		scope, err := u.Scopes.Create(ctx, ownerID, models.Scope{Name: "usecase_config"})
		requireError(t, err, nil)
		unchanged, err := u.Scopes.Update(ctx, ownerID, scope)
		requireError(t, err, nil)
		if !unchanged.UpdatedAt.Equal(scope.UpdatedAt) {
			t.Fatal("no-op scope update advanced updated_at")
		}
		scope.Name = "usecase_config_renamed"
		scope, err = u.Scopes.Update(ctx, ownerID, scope)
		requireError(t, err, nil)
		gotScope, err := u.Scopes.Get(ctx, ownerID, scope.ID)
		requireError(t, err, nil)
		if gotScope.Name != scope.Name {
			t.Fatalf("scope lookup = %+v", gotScope)
		}
		scopes, err := u.Scopes.List(ctx, ownerID, models.PageRequest{})
		requireError(t, err, nil)
		if len(scopes.Items) == 0 {
			t.Fatal("scope list is empty")
		}

		config, err := u.Configs.Create(ctx, ownerID, models.Config{
			ConfigKey: models.ConfigKey{ScopeID: scope.ID, OwnerUserID: ownerID, Name: "search_item"},
			Value:     json.RawMessage(`{"z":1,"x":true}`),
		})
		requireError(t, err, nil)
		originalConfig := config
		config.Value = json.RawMessage(`{ "x": true, "z": 1.00 }`)
		config, err = u.Configs.Update(ctx, ownerID, config)
		requireError(t, err, nil)
		if !config.UpdatedAt.Equal(originalConfig.UpdatedAt) || string(config.Value) != string(originalConfig.Value) {
			t.Fatal("semantically identical JSON changed the stored resource")
		}
		config.Value = json.RawMessage("null")
		config, err = u.Configs.Update(ctx, ownerID, config)
		requireError(t, err, nil)
		readConfig, err := u.Configs.Get(ctx, ownerID, config.ID)
		requireError(t, err, nil)
		if string(readConfig.Value) != "null" {
			t.Fatalf("config value = %s", readConfig.Value)
		}
		keys, err := u.Keys.List(ctx, ownerID, models.KeyFilter{
			ScopeID: &scope.ID, Search: "SEARCH_", Page: models.PageRequest{Limit: 1},
		})
		requireError(t, err, nil)
		key, err := u.Keys.Get(ctx, ownerID, config.ID)
		requireError(t, err, nil)
		if len(keys.Items) != 1 || keys.Items[0].ID != key.ID || key.ScopeName != scope.Name {
			t.Fatalf("key lookup = %+v, page = %+v", key, keys)
		}

		requireError(t, u.Configs.Delete(ctx, ownerID, config.ID), nil)
		changes := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeConfig, config.ID)
		if len(changes) != 3 || changes[0].Action != models.ActionDelete ||
			changes[1].Action != models.ActionUpdate || changes[2].Action != models.ActionCreate ||
			changes[0].NewValue != nil || changes[2].OldValue != nil {
			t.Fatalf("config history = %+v", changes)
		}
		entry, err := u.History.Get(ctx, ownerID, changes[1].ID)
		requireError(t, err, nil)
		var before, after models.Config
		requireError(t, json.Unmarshal(entry.OldValue, &before), nil)
		requireError(t, json.Unmarshal(entry.NewValue, &after), nil)
		if string(before.Value) != string(originalConfig.Value) || string(after.Value) != "null" ||
			entry.ActorUserID != ownerID || entry.ScopeName == nil || *entry.ScopeName != scope.Name {
			t.Fatalf("config update snapshots = %s, %s", entry.OldValue, entry.NewValue)
		}

		requireError(t, u.Scopes.Delete(ctx, ownerID, scope.ID), nil)
		scopeChanges := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeScope, scope.ID)
		if len(scopeChanges) != 3 || scopeChanges[0].Action != models.ActionDelete ||
			scopeChanges[1].Action != models.ActionUpdate || scopeChanges[2].Action != models.ActionCreate {
			t.Fatalf("scope history = %+v", scopeChanges)
		}
	})
}

func TestFlagEvaluationOverrideHistoryAndCascade(t *testing.T) {
	withFixture(t, func(ctx context.Context, store *repository.Store, u usecase.Usecases) {
		scope, err := u.Scopes.Create(ctx, ownerID, models.Scope{Name: "usecase_flags"})
		requireError(t, err, nil)
		flag, err := u.Flags.Create(ctx, ownerID, models.FeatureFlag{
			ConfigKey: models.ConfigKey{ScopeID: scope.ID, OwnerUserID: ownerID, Name: "feature"},
		})
		requireError(t, err, nil)
		gotFlag, err := u.Flags.Get(ctx, ownerID, flag.ID)
		requireError(t, err, nil)
		if gotFlag.Enabled {
			t.Fatal("new feature flag is enabled")
		}

		override := models.FeatureFlagOverride{KeyID: flag.ID, UserID: 200, Enabled: true}
		_, err = u.Overrides.Upsert(ctx, ownerID, override)
		requireError(t, err, nil)
		check := func(userID int64, enabled bool, reason onlineconf.Reason, hasBucket bool) onlineconf.Evaluation {
			t.Helper()
			result, err := u.Flags.Evaluate(ctx, ownerID, flag.ID, userID)
			requireError(t, err, nil)
			if result.Enabled != enabled || result.Reason != reason || (result.Bucket != nil) != hasBucket {
				t.Fatalf("flag evaluation for %d = %+v", userID, result)
			}
			return result
		}
		check(200, false, onlineconf.ReasonGloballyDisabled, false)
		flag, err = u.Flags.SetEnabled(ctx, ownerID, flag.ID, true)
		requireError(t, err, nil)
		_, err = u.Flags.SetEnabled(ctx, ownerID, flag.ID, true)
		requireError(t, err, nil)
		check(200, true, onlineconf.ReasonForcedInclude, false)
		first := check(300, false, onlineconf.ReasonRolloutExcluded, true)

		_, err = u.Overrides.Upsert(ctx, ownerID, override)
		requireError(t, err, nil)
		override.Enabled = false
		_, err = u.Overrides.Upsert(ctx, ownerID, override)
		requireError(t, err, nil)
		_, err = u.Flags.SetRolloutPercentage(ctx, ownerID, flag.ID, 100)
		requireError(t, err, apperrors.ErrInvalidInput)
		check(200, false, onlineconf.ReasonForcedExclude, false)
		_, err = u.Flags.SetRolloutPercentage(ctx, ownerID, flag.ID, 101)
		requireError(t, err, apperrors.ErrInvalidInput)
		flag.Name = "feature_renamed"
		flag, err = u.Flags.Update(ctx, ownerID, flag)
		requireError(t, err, nil)
		last := check(300, false, onlineconf.ReasonRolloutExcluded, true)
		if *first.Bucket != *last.Bucket || last.RolloutPercentage != 0 {
			t.Fatal("renaming feature flag changed its bucket")
		}
		item, err := u.Overrides.Get(ctx, ownerID, flag.ID, 200)
		requireError(t, err, nil)
		flagPage, err := u.Overrides.ListByFlag(ctx, ownerID, flag.ID, models.PageRequest{})
		requireError(t, err, nil)
		userPage, err := u.Overrides.ListByUser(ctx, ownerID, 200, models.PageRequest{})
		requireError(t, err, nil)
		if item.Enabled || len(flagPage.Items) != 1 || len(userPage.Items) != 1 {
			t.Fatal("override lookups do not match the saved override")
		}
		requireError(t, u.Overrides.Delete(ctx, ownerID, flag.ID, 200), nil)
		overrideChanges := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeFeatureFlagOverride, flag.ID)
		if len(overrideChanges) != 3 || overrideChanges[0].Action != models.ActionDelete ||
			overrideChanges[1].Action != models.ActionUpdate || overrideChanges[2].Action != models.ActionCreate {
			t.Fatalf("override history = %+v", overrideChanges)
		}

		for id := int64(1); id <= 105; id++ {
			_, err = store.Overrides.Upsert(ctx, models.FeatureFlagOverride{KeyID: flag.ID, UserID: id, Enabled: id%2 == 0})
			requireError(t, err, nil)
		}
		snapshot, err := u.Snapshots.Get(ctx)
		requireError(t, err, nil)
		if len(snapshot.Overrides) != 105 {
			t.Fatalf("snapshot has %d overrides", len(snapshot.Overrides))
		}

		requireError(t, u.Flags.Delete(ctx, ownerID, flag.ID), nil)
		changes := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeFeatureFlag, flag.ID)
		if len(changes) != 4 || changes[0].Action != models.ActionDelete || changes[3].Action != models.ActionCreate {
			t.Fatalf("flag history has %d entries", len(changes))
		}
		var deletion struct {
			models.FeatureFlag
			Overrides []models.FeatureFlagOverride `json:"overrides"`
		}
		requireError(t, json.Unmarshal(changes[0].OldValue, &deletion), nil)
		if len(deletion.Overrides) != 105 || deletion.Overrides[104].UserID != 105 ||
			deletion.ID != flag.ID || deletion.RolloutPercentage != 0 {
			t.Fatalf("flag deletion snapshot has %d overrides", len(deletion.Overrides))
		}
	})
}

func TestAdminPermissionAndLastOwner(t *testing.T) {
	withFixture(t, func(ctx context.Context, _ *repository.Store, u usecase.Usecases) {
		const adminID int64 = 300
		const otherOwner int64 = 301
		_, err := u.Admins.CheckAccess(ctx, adminID)
		requireError(t, err, apperrors.ErrForbidden)
		_, err = u.Scopes.Create(ctx, adminID, models.Scope{Name: "unauthorized"})
		requireError(t, err, apperrors.ErrForbidden)
		_, err = u.Admins.Grant(ctx, ownerID, adminID, models.AdminRoleAdmin)
		requireError(t, err, nil)
		_, err = u.Admins.CheckAccess(ctx, adminID)
		requireError(t, err, nil)

		_, err = u.Admins.Grant(ctx, adminID, otherOwner, models.AdminRoleOwner)
		requireError(t, err, apperrors.ErrForbidden)
		_, err = u.Admins.ChangeRole(ctx, adminID, adminID, models.AdminRoleOwner)
		requireError(t, err, apperrors.ErrForbidden)
		requireError(t, u.Admins.Revoke(ctx, adminID, ownerID), apperrors.ErrForbidden)
		_, err = u.Admins.List(ctx, adminID, models.PageRequest{})
		requireError(t, err, apperrors.ErrForbidden)
		_, err = u.Scopes.Create(ctx, adminID, models.Scope{Name: "admin_scope"})
		requireError(t, err, nil)
		_, err = u.History.List(ctx, adminID, models.HistoryFilter{})
		requireError(t, err, nil)

		requireError(t, u.Admins.Revoke(ctx, ownerID, ownerID), apperrors.ErrInvalidInput)
		_, err = u.Admins.ChangeRole(ctx, ownerID, ownerID, models.AdminRoleAdmin)
		requireError(t, err, apperrors.ErrInvalidInput)

		_, err = u.Admins.ChangeRole(ctx, ownerID, adminID, models.AdminRoleAdmin)
		requireError(t, err, nil)
		_, err = u.Admins.ChangeRole(ctx, ownerID, adminID, models.AdminRoleOwner)
		requireError(t, err, nil)
		_, err = u.Admins.List(ctx, ownerID, models.PageRequest{})
		requireError(t, err, nil)
		requireError(t, u.Admins.Revoke(ctx, ownerID, adminID), nil)
		_, err = u.Admins.CheckAccess(ctx, adminID)
		requireError(t, err, apperrors.ErrForbidden)
		changes := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeAdmin, adminID)
		if len(changes) != 3 || changes[0].Action != models.ActionDelete ||
			changes[1].Action != models.ActionUpdate || changes[2].Action != models.ActionCreate {
			t.Fatalf("admin history = %+v", changes)
		}
	})
}

func TestConcurrentScopeAuditUsesActualPreviousVersion(t *testing.T) {
	ctx, store, pool := openTestStore(t)
	u := usecase.New(store, fixtureUsers{})
	scope, err := u.Scopes.Create(ctx, ownerID, models.Scope{Name: "concurrent_usecase"})
	requireError(t, err, nil)
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		requireError(t, u.Scopes.Delete(cleanupCtx, ownerID, scope.ID), nil)
	})
	first, second := scope, scope
	first.Name, second.Name = "concurrent_first", "concurrent_second"

	ready := make(chan struct{})
	release := make(chan struct{})
	releaseOnce := sync.OnceFunc(func() { close(release) })
	defer releaseOnce()
	writerDone := make(chan error, 1)
	go func() {
		writerDone <- store.WithinTx(ctx, func(tx *repository.Store) error {
			if _, err := usecase.New(tx, fixtureUsers{}).Scopes.Update(ctx, ownerID, first); err != nil {
				return err
			}
			close(ready)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-ready:
	case err := <-writerDone:
		t.Fatalf("first writer: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	secondDone := make(chan error, 1)
	go func() {
		_, err := u.Scopes.Update(ctx, ownerID, second)
		secondDone <- err
	}()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT FROM pg_stat_activity
				WHERE application_name = 'onlineconf_usecase_test' AND wait_event_type = 'Lock'
				  AND query LIKE '%UPDATE public.scope SET%'
			)`).Scan(&waiting)
		requireError(t, err, nil)
		if waiting {
			break
		}
		select {
		case err := <-secondDone:
			t.Fatalf("second writer returned before the lock was released: %v", err)
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	releaseOnce()
	requireError(t, <-writerDone, nil)
	requireError(t, <-secondDone, apperrors.ErrRetryable)

	_, err = u.Scopes.Update(ctx, ownerID, second)
	requireError(t, err, nil)
	changes := readHistory(t, ctx, u.History, ownerID, models.ResourceTypeScope, scope.ID)
	if len(changes) != 3 {
		t.Fatalf("unexpected scope history length: %d", len(changes))
	}
	var previous models.Scope
	requireError(t, json.Unmarshal(changes[0].OldValue, &previous), nil)
	if previous.Name != first.Name {
		t.Fatalf("audited previous scope name = %q, want %q", previous.Name, first.Name)
	}
}

func TestFlagDeleteAndOverrideUseParentLockFirst(t *testing.T) {
	ctx, store, pool := openTestStore(t)
	u := usecase.New(store, fixtureUsers{})
	scope, err := store.Scopes.GetByName(ctx, "auth")
	requireError(t, err, nil)
	flag, err := u.Flags.Create(ctx, ownerID, models.FeatureFlag{
		ConfigKey: models.ConfigKey{ScopeID: scope.ID, Name: "lock_order_test", OwnerUserID: ownerID},
	})
	requireError(t, err, nil)
	override := models.FeatureFlagOverride{KeyID: flag.ID, UserID: ownerID, Enabled: true}
	_, err = u.Overrides.Upsert(ctx, ownerID, override)
	requireError(t, err, nil)

	competing := make(chan error, 1)
	err = store.WithinTx(ctx, func(tx *repository.Store) error {
		if _, err := tx.Flags.GetForUpdate(ctx, flag.ID); err != nil {
			return err
		}
		override.Enabled = false
		go func() {
			_, err := u.Overrides.Upsert(ctx, ownerID, override)
			competing <- err
		}()
		if err := waitForBlockedStatement(ctx, pool, "%FOR NO KEY UPDATE OF k", competing); err != nil {
			return err
		}
		return usecase.New(tx, fixtureUsers{}).Flags.Delete(ctx, ownerID, flag.ID)
	})
	requireError(t, err, nil)
	requireError(t, <-competing, apperrors.ErrRetryable)
}

func TestMutualAdminDemotionChecksActorAfterLock(t *testing.T) {
	ctx, store, pool := openTestStore(t)
	const otherOwner int64 = 42
	actor := ownerID
	_, err := store.Admins.Create(ctx, models.Admin{UserID: otherOwner, Role: models.AdminRoleOwner, GrantedBy: &actor})
	requireError(t, err, nil)
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		requireError(t, store.Admins.Delete(cleanup, otherOwner), nil)
	})
	competing := make(chan error, 1)
	err = store.WithinTx(ctx, func(tx *repository.Store) error {
		if err := tx.Admins.LockChanges(ctx); err != nil {
			return err
		}
		go func() {
			_, err := usecase.New(store, fixtureUsers{}).Admins.ChangeRole(ctx, otherOwner, ownerID, models.AdminRoleAdmin)
			competing <- err
		}()
		if err := waitForBlockedStatement(ctx, pool, "LOCK TABLE public.admin%", competing); err != nil {
			return err
		}
		_, err := tx.Admins.UpdateRole(ctx, otherOwner, models.AdminRoleAdmin, &actor)
		return err
	})
	requireError(t, err, nil)
	requireError(t, <-competing, apperrors.ErrForbidden)
}

func waitForBlockedStatement(ctx context.Context, pool *pgxpool.Pool, statement string, result <-chan error) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		err := pool.QueryRow(ctx, `SELECT EXISTS (
   SELECT FROM pg_stat_activity WHERE application_name = 'onlineconf_usecase_test'
     AND pid <> pg_backend_pid() AND wait_event_type = 'Lock' AND query LIKE $1
  )`, statement).Scan(&blocked)
		if err != nil || blocked {
			return err
		}
		select {
		case err := <-result:
			return fmt.Errorf("competing operation finished before blocking: %v", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
