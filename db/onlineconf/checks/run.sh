#!/usr/bin/env bash
set -euo pipefail

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../.." && pwd)
cd "$root"
scratch=$(mktemp -d)
container=
writer=
cleanup() {
    if [[ -n $writer ]]; then
        wait "$writer" || true
    fi
    if [[ -n $container ]]; then
        docker stop -t 1 "$container" >/dev/null
    fi
    rm -r -- "$scratch"
}
trap cleanup EXIT

export POSTGRES_PASSWORD
POSTGRES_PASSWORD=$(openssl rand -hex 24)
migrator_password=$(openssl rand -hex 24)
runtime_password=$(openssl rand -hex 24)
container=$(docker run --rm -d -p 127.0.0.1::5432 -e POSTGRES_PASSWORD -e POSTGRES_DB=product_test postgres:17-alpine)
address=$(docker port "$container" 5432/tcp)
port=${address##*:}

export PGPASSFILE="$scratch/pgpass"
unset PGPASSWORD
(
    umask 077
    {
        printf '127.0.0.1:%s:*:postgres:%s\n' "$port" "$POSTGRES_PASSWORD"
        printf '127.0.0.1:%s:*:onlineconf_migrator:%s\n' "$port" "$migrator_password"
        printf '127.0.0.1:%s:*:onlineconf_runtime:%s\n' "$port" "$runtime_password"
    } > "$PGPASSFILE"
)
export ONLINECONF_ADMIN_DATABASE_URL="postgresql://postgres@$address/postgres?sslmode=disable"
export ONLINECONF_MIGRATE_DATABASE_URL="postgresql://onlineconf_migrator@$address/onlineconf?sslmode=disable"
export ONLINECONF_PRODUCT_DATABASE_URL="postgresql://postgres@$address/product_test?sslmode=disable"
owner_url="postgresql://postgres@$address/onlineconf?sslmode=disable"
runtime_url="postgresql://onlineconf_runtime@$address/onlineconf?sslmode=disable"

psql_checked() {
    psql -X --no-password -v ON_ERROR_STOP=1 "$@"
}

expect_failure() {
    local message=$1
    shift
    if "$@" >"$scratch/expected-error" 2>&1; then
        echo "Expected failure: $*" >&2
        exit 1
    fi
    if ! grep -Fq -- "$message" "$scratch/expected-error"; then
        cat "$scratch/expected-error" >&2
        exit 1
    fi
}

ready=false
for ((attempt=0; attempt<60; attempt++)); do
    if psql_checked "$ONLINECONF_ADMIN_DATABASE_URL" -Atc 'SELECT 1' >/dev/null 2>&1; then
        ready=true
        break
    fi
    sleep 0.5
done
if [[ $ready != true ]]; then
    psql_checked "$ONLINECONF_ADMIN_DATABASE_URL" -Atc 'SELECT 1' || true
    docker logs "$container" >&2
    exit 1
fi
psql_checked "$ONLINECONF_ADMIN_DATABASE_URL" -c 'SHOW server_version'

make onlineconf-db-init
psql_checked "$ONLINECONF_PRODUCT_DATABASE_URL" -f db/migrations/000001_create_app_user.up.sql
psql_checked "$ONLINECONF_PRODUCT_DATABASE_URL" -c \
    "INSERT INTO public.app_user (email, password_hash, role) VALUES ('owner@example.test', 'unused', 'seeker')"

export migrator_password runtime_password
psql_checked "$ONLINECONF_ADMIN_DATABASE_URL" <<'SQL'
\getenv migrator_password migrator_password
\getenv runtime_password runtime_password
ALTER ROLE onlineconf_migrator PASSWORD :'migrator_password';
ALTER ROLE onlineconf_runtime PASSWORD :'runtime_password';
SQL
unset POSTGRES_PASSWORD migrator_password runtime_password

for role in onlineconf_migrator onlineconf_runtime; do
    product_role_url="postgresql://$role@$address/product_test?sslmode=disable"
    [[ $(psql_checked "$product_role_url" -Atc "SELECT
        has_database_privilege(current_user, current_database(), 'CONNECT')
        AND has_database_privilege(current_user, current_database(), 'TEMPORARY')
        AND NOT has_database_privilege(current_user, current_database(), 'CREATE')
        AND has_schema_privilege(current_user, 'public', 'USAGE')
        AND NOT has_schema_privilege(current_user, 'public', 'CREATE')") == t ]]
    psql_checked "$product_role_url" -c 'CREATE TEMP TABLE public_privilege_check (id INTEGER)'
    expect_failure 'permission denied for table app_user' psql_checked "$product_role_url" -c 'SELECT user_id FROM public.app_user'
    expect_failure 'permission denied for schema public' psql_checked "$product_role_url" -c 'CREATE TABLE public.unexpected_table (id INTEGER)'
done
psql_checked "$ONLINECONF_PRODUCT_DATABASE_URL" -f db/onlineconf/checks/product_privileges.sql
echo 'Effective product permissions verified, including PUBLIC grants'

make onlineconf-db-init
make onlineconf-migrate-up
make onlineconf-migrate-version
make onlineconf-migrate-down
[[ $(psql_checked "$owner_url" -Atc "SELECT to_regclass('public.config_key') IS NULL") == t ]]
make onlineconf-migrate-up
make onlineconf-migrate-version
psql_checked "$runtime_url" -f db/onlineconf/checks/schema.sql
echo 'Schema and runtime permissions verified'

expect_failure 'delete config_key rows' psql_checked "$ONLINECONF_MIGRATE_DATABASE_URL" -c 'TRUNCATE public.config'
expect_failure 'must use onlineconf_migrator' env ONLINECONF_MIGRATE_DATABASE_URL="$ONLINECONF_PRODUCT_DATABASE_URL" make onlineconf-migrate-up

psql_checked "$runtime_url" <<'SQL'
BEGIN;
INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
SELECT scope_id, 'concurrent_check', 'config', 1 FROM public.scope WHERE name = 'auth';
INSERT INTO public.config (key_id, value)
SELECT key_id, '1' FROM public.config_key WHERE name = 'concurrent_check';
COMMIT;
SQL

PGAPPNAME=onlineconf_check_writer psql_checked "$runtime_url" >"$scratch/writer.log" 2>&1 <<'SQL' &
BEGIN;
SELECT key_id FROM public.config_key WHERE name = 'concurrent_check' FOR NO KEY UPDATE;
DELETE FROM public.config USING public.config_key k
WHERE config.key_id = k.key_id AND k.name = 'concurrent_check';
INSERT INTO public.config (key_id, value)
SELECT key_id, '2' FROM public.config_key WHERE name = 'concurrent_check';
SELECT pg_sleep(3);
COMMIT;
SQL
writer=$!
wait_for_writer_sleep() {
    for ((attempt=0; attempt<40; attempt++)); do
        if [[ $(psql_checked "$owner_url" -Atc \
            "SELECT EXISTS (SELECT FROM pg_stat_activity WHERE application_name = 'onlineconf_check_writer' AND wait_event = 'PgSleep')") == t ]]; then
            return
        fi
        sleep 0.1
    done
    cat "$scratch/writer.log" >&2
    return 1
}
wait_for_writer_sleep
expect_failure 'must have exactly one subtype' psql_checked "$runtime_url" -c \
    "INSERT INTO public.feature_flag (key_id) SELECT key_id FROM public.config_key WHERE name = 'concurrent_check'"
wait "$writer"
writer=
[[ $(psql_checked "$runtime_url" -Atc "SELECT value FROM public.config JOIN public.config_key USING (key_id) WHERE name = 'concurrent_check'") == 2 ]]
psql_checked "$runtime_url" -c "DELETE FROM public.config_key WHERE name = 'concurrent_check'"
echo 'Concurrent subtype writes verified'

expect_failure 'product user does not exist' make onlineconf-admin-bootstrap ONLINECONF_OWNER_USER_ID=9999
make onlineconf-admin-bootstrap ONLINECONF_OWNER_USER_ID=1
expect_failure 'already has or previously had administrators' make onlineconf-admin-bootstrap ONLINECONF_OWNER_USER_ID=1
make onlineconf-db-init
[[ $(psql_checked "$runtime_url" -Atc "SELECT count(*) FROM public.admin WHERE user_id = 1 AND role = 'owner' AND granted_by IS NULL") == 1 ]]
[[ $(psql_checked "$runtime_url" -Atc "SELECT count(*) FROM public.change_history WHERE resource_type = 'admin' AND resource_id = 1 AND actor_user_id = 1 AND action = 'create'") == 1 ]]
expect_failure 'cannot remove the last OnlineConf owner' psql_checked "$runtime_url" -c 'DELETE FROM public.admin WHERE user_id = 1'
expect_failure 'cannot remove the last OnlineConf owner' psql_checked "$runtime_url" -c "UPDATE public.admin SET role = 'admin', granted_by = 2 WHERE user_id = 1"
expect_failure 'cannot remove the last OnlineConf owner' psql_checked "$ONLINECONF_MIGRATE_DATABASE_URL" -c 'TRUNCATE public.admin'

psql_checked "$runtime_url" -c "INSERT INTO public.admin (user_id, role, granted_by) VALUES (2, 'owner', 1)"
expect_failure 'cannot remove the last OnlineConf owner' psql_checked "$runtime_url" -c "DELETE FROM public.admin WHERE role = 'owner'"
PGAPPNAME=onlineconf_check_writer psql_checked "$runtime_url" >"$scratch/writer.log" 2>&1 <<'SQL' &
BEGIN;
DELETE FROM public.admin WHERE user_id = 2;
SELECT pg_sleep(3);
COMMIT;
SQL
writer=$!
wait_for_writer_sleep
expect_failure 'cannot remove the last OnlineConf owner' psql_checked "$runtime_url" -c "UPDATE public.admin SET role = 'admin', granted_by = 2 WHERE user_id = 1"
wait "$writer"
writer=
[[ $(psql_checked "$runtime_url" -Atc "SELECT count(*) FROM public.admin WHERE role = 'owner'") == 1 ]]
echo 'Last-owner protection verified with concurrent deletion and demotion'
expect_failure 'refusing to roll back OnlineConf with non-seed data' make onlineconf-migrate-down
[[ $(psql_checked "$runtime_url" -Atc 'SELECT count(*) FROM public.change_history') == 1 ]]
[[ $(psql_checked "$ONLINECONF_PRODUCT_DATABASE_URL" -Atc 'SELECT count(*) FROM public.app_user') == 1 ]]
echo 'Owner bootstrap, reinitialization and rollback protection verified'
