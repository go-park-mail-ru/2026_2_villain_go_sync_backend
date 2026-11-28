#!/usr/bin/env bash
set -euo pipefail

directory=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)

psql_onlineconf() {
    psql -X --no-password -v ON_ERROR_STOP=1 "$@"
}

require_migrator() {
    : "${ONLINECONF_MIGRATE_DATABASE_URL:?set the OnlineConf migrator connection URL}"
    local connection
    connection=$(psql_onlineconf --dbname="$ONLINECONF_MIGRATE_DATABASE_URL" -Atc \
        "SELECT current_database() = 'onlineconf' AND current_user = 'onlineconf_migrator' AND current_schema() = 'public'")
    if [[ $connection != t ]]; then
        echo 'The migration URL must use onlineconf_migrator on the onlineconf database with public as the current schema.' >&2
        exit 1
    fi
}

case "${1:-}" in
    init)
        : "${ONLINECONF_ADMIN_DATABASE_URL:?set the PostgreSQL admin URL for the postgres maintenance database}"
        psql_onlineconf --dbname="$ONLINECONF_ADMIN_DATABASE_URL" --file="$directory/init.sql"
        ;;
    up|down|version)
        require_migrator
        operation=("$1")
        if [[ $1 == down ]]; then
            operation+=(1)
        fi
        "${ONLINECONF_MIGRATE_BIN:-migrate}" -path "$directory/migrations" \
            -database "$ONLINECONF_MIGRATE_DATABASE_URL" "${operation[@]}"
        ;;
    bootstrap)
        require_migrator
        : "${ONLINECONF_PRODUCT_DATABASE_URL:?set the product database URL with read access to public.app_user}"
        if [[ ! ${ONLINECONF_OWNER_USER_ID:-} =~ ^[1-9][0-9]{0,18}$ ]]; then
            echo 'Set ONLINECONF_OWNER_USER_ID to the positive ID of an existing product user.' >&2
            exit 1
        fi
        exists=$(psql_onlineconf --dbname="$ONLINECONF_PRODUCT_DATABASE_URL" -At \
            -v owner_id="$ONLINECONF_OWNER_USER_ID" <<'SQL'
SELECT EXISTS (SELECT FROM public.app_user WHERE user_id = :'owner_id'::BIGINT);
SQL
        )
        if [[ $exists != t ]]; then
            echo 'The specified product user does not exist.' >&2
            exit 1
        fi
        psql_onlineconf --dbname="$ONLINECONF_MIGRATE_DATABASE_URL" \
            -v owner_id="$ONLINECONF_OWNER_USER_ID" --file="$directory/bootstrap.sql"
        ;;
    *)
        echo 'Usage: manage.sh init|up|down|version|bootstrap' >&2
        exit 2
        ;;
esac
