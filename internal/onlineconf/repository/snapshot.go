package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/apperrors"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/internal/onlineconf/models"
	"github.com/go-park-mail-ru/2026_2_villain_go_sync_backend/pkg/onlineconf"
	"github.com/jackc/pgx/v5"
)

type SnapshotRepository struct {
	conn connection
}

// The estimate includes the wire fields, escaping, and overhead per row.
// The HTTP layer separately checks the exact envelope size.
const snapshotBudgetQuery = `
 SELECT
   COALESCE((SELECT sum(256 + 6 * (octet_length(k.name)::bigint + octet_length(s.name)
     + COALESCE(octet_length(k.description), 0)) + 6 * octet_length(c.value::text)::bigint)
     FROM public.config c JOIN public.config_key k USING (key_id) JOIN public.scope s USING (scope_id)), 0)
 + COALESCE((SELECT sum(256 + 6 * (octet_length(k.name)::bigint + octet_length(s.name)
     + COALESCE(octet_length(k.description), 0)))
     FROM public.feature_flag f JOIN public.config_key k USING (key_id) JOIN public.scope s USING (scope_id)), 0)
 + (SELECT count(*) * 256 FROM public.feature_flag_override)`

func (r SnapshotRepository) Get(ctx context.Context) (models.Snapshot, error) {
	var snapshot models.Snapshot
	err := r.conn.withinTx(ctx, func(tx pgx.Tx) error {
		var size int64
		if err := tx.QueryRow(ctx, snapshotBudgetQuery).Scan(&size); err != nil {
			return err
		}
		if size > onlineconf.MaxSnapshotBytes {
			return apperrors.ErrDependencyUnavailable
		}
		var err error
		snapshot.Configs, err = snapshotRows(ctx, tx, "SELECT "+keyColumns+", c.value"+keyFrom+
			" JOIN public.config c ON c.key_id = k.key_id ORDER BY k.key_id", func(row rowScanner) (models.Config, error) {
			var c models.Config
			err := row.Scan(append(keyFields(&c.ConfigKey), &c.Value)...)
			return c, err
		})
		if err != nil {
			return err
		}
		snapshot.Flags, err = snapshotRows(ctx, tx, "SELECT "+keyColumns+", f.enabled, f.rollout_percentage"+keyFrom+
			" JOIN public.feature_flag f ON f.key_id = k.key_id ORDER BY k.key_id", func(row rowScanner) (models.FeatureFlag, error) {
			var f models.FeatureFlag
			err := row.Scan(append(keyFields(&f.ConfigKey), &f.Enabled, &f.RolloutPercentage)...)
			return f, err
		})
		if err != nil {
			return err
		}
		snapshot.Overrides, err = snapshotRows(ctx, tx, "SELECT "+overrideColumns+
			" FROM public.feature_flag_override ORDER BY key_id, user_id", scanOverride)
		return err
	})
	if err != nil {
		return models.Snapshot{}, err
	}
	return snapshot, nil
}

func snapshotRows[T any](ctx context.Context, db database, query string, scan func(rowScanner) (T, error)) ([]T, error) {
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.AppendRows([]T{}, rows, func(row pgx.CollectableRow) (T, error) { return scan(row) })
}
