package productusers

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Reader struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, dsn string) (*Reader, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("invalid product reader connection URL")
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0
	cfg.ConnConfig.RuntimeParams["default_transaction_read_only"] = "on"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("cannot configure product reader pool")
	}

	var allowed bool
	err = pool.QueryRow(ctx, `
		SELECT current_database() <> 'onlineconf'
		   AND session_user = current_user
		   AND current_user NOT IN ('onlineconf_runtime', 'onlineconf_migrator')
		   AND current_setting('transaction_read_only') = 'on'
		   AND NOT (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)
		   AND NOT EXISTS (
       SELECT FROM pg_roles AS other WHERE other.rolname <> current_user
         AND pg_has_role(current_user, other.oid, 'MEMBER')
   )
   AND NOT has_database_privilege(current_user, current_database(), 'CREATE')
   AND NOT EXISTS (
       SELECT FROM pg_namespace n WHERE n.nspname !~ '^pg_'
         AND n.nspname <> 'information_schema'
         AND has_schema_privilege(current_user, n.oid, 'CREATE')
   )
   AND NOT EXISTS (
       SELECT FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
       WHERE n.nspname !~ '^pg_' AND n.nspname <> 'information_schema'
         AND c.oid <> 'public.app_user'::regclass
         AND CASE WHEN c.relkind = 'S' THEN has_sequence_privilege(current_user, c.oid, 'USAGE, SELECT, UPDATE')
                  WHEN c.relkind IN ('r', 'p', 'v', 'm', 'f') THEN
                      has_table_privilege(current_user, c.oid, 'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN')
                      OR has_any_column_privilege(current_user, c.oid, 'SELECT, INSERT, UPDATE, REFERENCES')
                  ELSE false END
   )
   AND has_column_privilege(current_user, 'public.app_user', 'user_id', 'SELECT')
		   AND NOT has_table_privilege(current_user, 'public.app_user',
		       'SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, MAINTAIN')
		   AND NOT has_any_column_privilege(current_user, 'public.app_user', 'INSERT, UPDATE, REFERENCES')
		   AND NOT EXISTS (
		       SELECT FROM pg_attribute WHERE attrelid = 'public.app_user'::regclass AND attnum > 0
		         AND NOT attisdropped AND attname <> 'user_id'
		         AND has_column_privilege(current_user, attrelid, attnum, 'SELECT')
		   )
		FROM pg_roles WHERE rolname = current_user`).Scan(&allowed)
	if err != nil || !allowed {
		pool.Close()
		return nil, errors.New("product reader must be a separate unprivileged role with SELECT on public.app_user.user_id only")
	}
	return &Reader{pool: pool}, nil
}

func (r *Reader) Exists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM public.app_user WHERE user_id = $1)", userID).Scan(&exists)
	return exists, err
}

func (r *Reader) Ping(ctx context.Context) error {
	_, err := r.Exists(ctx, 0)
	return err
}

func (r *Reader) Close() {
	r.pool.Close()
}
