SELECT pg_advisory_lock(498842692103894483);

DO $$
DECLARE
    role_name TEXT;
BEGIN
    IF current_database() <> 'postgres' THEN
        RAISE EXCEPTION 'initialize OnlineConf from the postgres maintenance database';
    END IF;

    FOREACH role_name IN ARRAY ARRAY['onlineconf_migrator', 'onlineconf_runtime']
    LOOP
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = role_name) THEN
            EXECUTE format(
                'CREATE ROLE %I LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS',
                role_name
            );
        END IF;
        IF EXISTS (
            SELECT FROM pg_roles r
            WHERE r.rolname = role_name
              AND (NOT r.rolcanlogin OR r.rolsuper OR r.rolcreatedb OR r.rolcreaterole
                   OR r.rolreplication OR r.rolbypassrls
                   OR EXISTS (SELECT FROM pg_auth_members m WHERE m.member = r.oid))
        ) THEN
            RAISE EXCEPTION 'existing role % has incompatible privileges', role_name;
        END IF;
    END LOOP;

    IF EXISTS (
        SELECT FROM pg_database d JOIN pg_roles r ON r.oid = d.datdba
        WHERE d.datname = 'onlineconf' AND r.rolname <> current_user
    ) THEN
        RAISE EXCEPTION 'initialize OnlineConf as the existing onlineconf database owner';
    END IF;
END;
$$;

SELECT 'CREATE DATABASE onlineconf TEMPLATE template0'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'onlineconf')
\gexec

\connect -reuse-previous=on onlineconf

BEGIN;

SELECT pg_advisory_xact_lock(498842692103894483);

REVOKE ALL ON DATABASE onlineconf FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT CONNECT ON DATABASE onlineconf TO onlineconf_migrator, onlineconf_runtime;
GRANT USAGE, CREATE ON SCHEMA public TO onlineconf_migrator;
GRANT USAGE ON SCHEMA public TO onlineconf_runtime;

COMMIT;
