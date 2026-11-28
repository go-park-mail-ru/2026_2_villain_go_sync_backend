BEGIN;

DO $$
DECLARE
    scope_frontend BIGINT;
    scope_auth BIGINT;
    config_id BIGINT;
    other_config_id BIGINT;
    flag_id BIGINT;
    incomplete_id BIGINT;
    json_value JSONB;
    previous_update TIMESTAMPTZ;
    allowed_json_types TEXT[] := ARRAY['string', 'number', 'boolean', 'object', 'array', 'null'];
    position INTEGER := 0;
BEGIN
    IF (SELECT array_agg(name ORDER BY name) FROM public.scope)
        IS DISTINCT FROM ARRAY['application', 'auth', 'frontend', 'seeker', 'vacancy'] THEN
        RAISE EXCEPTION 'unexpected seed scopes';
    END IF;
    IF EXISTS (SELECT FROM public.admin) OR EXISTS (SELECT FROM public.config_key) THEN
        RAISE EXCEPTION 'unexpected seeded admins or keys';
    END IF;

    SELECT scope_id INTO STRICT scope_frontend FROM public.scope WHERE name = 'frontend';
    SELECT scope_id INTO STRICT scope_auth FROM public.scope WHERE name = 'auth';

    BEGIN
        INSERT INTO public.scope (name) VALUES ('bad.scope');
        RAISE EXCEPTION 'dotted scope accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.scope (name) VALUES ('');
        RAISE EXCEPTION 'empty scope accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.scope (name) VALUES ('frontend');
        RAISE EXCEPTION 'duplicate scope accepted';
    EXCEPTION WHEN unique_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, 'bad.key', 'config', 42);
        RAISE EXCEPTION 'dotted key name accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, '', 'config', 42);
        RAISE EXCEPTION 'empty key name accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;

    INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
    VALUES (scope_frontend, 'same_name', 'config', 42) RETURNING key_id INTO config_id;
    INSERT INTO public.config (key_id, value) VALUES (config_id, 'null');

    INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
    VALUES (scope_auth, 'same_name', 'feature_flag', 42) RETURNING key_id INTO flag_id;
    INSERT INTO public.feature_flag (key_id) VALUES (flag_id);

    SET CONSTRAINTS ALL IMMEDIATE;
    SET CONSTRAINTS ALL DEFERRED;

    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, 'same_name', 'feature_flag', 42);
        RAISE EXCEPTION 'duplicate key within scope accepted';
    EXCEPTION WHEN unique_violation THEN NULL;
    END;
    BEGIN
        UPDATE public.config_key SET type = 'feature_flag' WHERE key_id = config_id;
        RAISE EXCEPTION 'key type mutation accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, 'missing_subtype', 'config', 42);
        SET CONSTRAINTS ALL IMMEDIATE;
        RAISE EXCEPTION 'key without subtype accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, 'wrong_subtype', 'feature_flag', 42) RETURNING key_id INTO incomplete_id;
        INSERT INTO public.config (key_id, value) VALUES (incomplete_id, 'null');
        SET CONSTRAINTS ALL IMMEDIATE;
        RAISE EXCEPTION 'key with wrong subtype accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.feature_flag (key_id) VALUES (config_id);
        SET CONSTRAINTS ALL IMMEDIATE;
        RAISE EXCEPTION 'key with both subtypes accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        DELETE FROM public.config WHERE key_id = config_id;
        SET CONSTRAINTS ALL IMMEDIATE;
        RAISE EXCEPTION 'standalone subtype deletion accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config_key (scope_id, name, type, owner_user_id)
        VALUES (scope_frontend, 'moved_subtype', 'config', 42) RETURNING key_id INTO other_config_id;
        UPDATE public.config SET key_id = other_config_id WHERE key_id = config_id;
        SET CONSTRAINTS ALL IMMEDIATE;
        RAISE EXCEPTION 'subtype move left an orphan key';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        INSERT INTO public.config (key_id, value) VALUES (0, 'null');
        RAISE EXCEPTION 'subtype without a parent accepted';
    EXCEPTION WHEN foreign_key_violation THEN NULL;
    END;

    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = config_id;
    PERFORM pg_sleep(0.002);
    UPDATE public.config SET value = '42' WHERE key_id = config_id;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = config_id) <= previous_update THEN
        RAISE EXCEPTION 'config value did not update the key timestamp';
    END IF;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = config_id;
    BEGIN
        UPDATE public.config SET value = '43' WHERE key_id = config_id;
        RAISE no_data_found;
    EXCEPTION WHEN no_data_found THEN NULL;
    END;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = config_id) IS DISTINCT FROM previous_update THEN
        RAISE EXCEPTION 'rolled back config change updated the key timestamp';
    END IF;

    FOREACH json_value IN ARRAY ARRAY['"text"', '12.5', 'true', '{"a":1}', '[1,false]', 'null']::JSONB[]
    LOOP
        position := position + 1;
        UPDATE public.config SET value = json_value WHERE key_id = config_id;
        IF NOT EXISTS (SELECT FROM public.config WHERE key_id = config_id
                      AND value = json_value AND jsonb_typeof(value) = allowed_json_types[position]) THEN
            RAISE EXCEPTION 'JSON round-trip failed for type %', allowed_json_types[position];
        END IF;
    END LOOP;
    BEGIN
        UPDATE public.config SET value = NULL WHERE key_id = config_id;
        RAISE EXCEPTION 'SQL NULL accepted as a config value';
    EXCEPTION WHEN not_null_violation THEN NULL;
    END;

    IF NOT EXISTS (SELECT FROM public.feature_flag WHERE key_id = flag_id
                   AND NOT enabled AND rollout_percentage = 0) THEN
        RAISE EXCEPTION 'incorrect flag defaults';
    END IF;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = flag_id;
    PERFORM pg_sleep(0.002);
    UPDATE public.feature_flag SET rollout_percentage = 100 WHERE key_id = flag_id;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = flag_id) <= previous_update THEN
        RAISE EXCEPTION 'rollout did not update the key timestamp';
    END IF;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = flag_id;
    PERFORM pg_sleep(0.002);
    UPDATE public.feature_flag SET enabled = TRUE WHERE key_id = flag_id;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = flag_id) <= previous_update THEN
        RAISE EXCEPTION 'toggle did not update the key timestamp';
    END IF;
    BEGIN
        UPDATE public.feature_flag SET rollout_percentage = -1 WHERE key_id = flag_id;
        RAISE EXCEPTION 'negative rollout accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    BEGIN
        UPDATE public.feature_flag SET rollout_percentage = 101 WHERE key_id = flag_id;
        RAISE EXCEPTION 'rollout greater than 100 accepted';
    EXCEPTION WHEN check_violation THEN NULL;
    END;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = flag_id;
    PERFORM pg_sleep(0.002);
    INSERT INTO public.feature_flag_override (key_id, user_id, enabled)
    VALUES (flag_id, 100, TRUE), (flag_id, 101, FALSE);
    IF (SELECT updated_at FROM public.config_key WHERE key_id = flag_id) <= previous_update THEN
        RAISE EXCEPTION 'override insert did not update the key timestamp';
    END IF;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = flag_id;
    PERFORM pg_sleep(0.002);
    UPDATE public.feature_flag_override SET enabled = TRUE WHERE key_id = flag_id AND user_id = 101;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = flag_id) <= previous_update THEN
        RAISE EXCEPTION 'override update did not update the key timestamp';
    END IF;
    SELECT updated_at INTO previous_update FROM public.config_key WHERE key_id = flag_id;
    PERFORM pg_sleep(0.002);
    DELETE FROM public.feature_flag_override WHERE key_id = flag_id AND user_id = 101;
    IF (SELECT updated_at FROM public.config_key WHERE key_id = flag_id) <= previous_update THEN
        RAISE EXCEPTION 'override delete did not update the key timestamp';
    END IF;
    BEGIN
        INSERT INTO public.feature_flag_override (key_id, user_id, enabled) VALUES (flag_id, 100, FALSE);
        RAISE EXCEPTION 'duplicate override accepted';
    EXCEPTION WHEN unique_violation THEN NULL;
    END;
    BEGIN
        DELETE FROM public.scope WHERE scope_id = scope_auth;
        RAISE EXCEPTION 'nonempty scope deleted';
    EXCEPTION WHEN foreign_key_violation THEN NULL;
    END;

    INSERT INTO public.change_history
        (actor_user_id, resource_type, resource_id, scope_name, resource_name, action, old_value)
    VALUES (42, 'scope', scope_auth, 'auth', 'auth', 'delete', '{}'),
           (42, 'config', config_id, 'frontend', 'same_name', 'delete', 'null'),
           (42, 'feature_flag', flag_id, 'auth', 'same_name', 'delete', '{"enabled":false}'),
           (42, 'feature_flag_override', flag_id, 'auth', 'same_name', 'delete', '{"user_id":100,"enabled":true}'),
           (42, 'admin', 100, NULL, '100', 'delete', '{"role":"admin"}');

    DELETE FROM public.config_key WHERE key_id IN (config_id, flag_id);
    DELETE FROM public.scope WHERE scope_id = scope_auth;
    SET CONSTRAINTS ALL IMMEDIATE;

    IF EXISTS (SELECT FROM public.config WHERE key_id = config_id)
        OR EXISTS (SELECT FROM public.feature_flag WHERE key_id = flag_id)
        OR EXISTS (SELECT FROM public.feature_flag_override WHERE key_id = flag_id) THEN
        RAISE EXCEPTION 'cascade delete failed';
    END IF;
    IF (SELECT count(*) FROM public.change_history) <> 5 THEN
        RAISE EXCEPTION 'history was lost';
    END IF;

    BEGIN
        UPDATE public.change_history SET action = 'update';
        RAISE EXCEPTION 'runtime can UPDATE history';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        DELETE FROM public.change_history;
        RAISE EXCEPTION 'runtime can DELETE history';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        TRUNCATE public.change_history;
        RAISE EXCEPTION 'runtime can TRUNCATE history';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        PERFORM 1 FROM public.schema_migrations;
        RAISE EXCEPTION 'runtime can read migration state';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        CREATE TABLE public.unexpected_table (value INTEGER);
        RAISE EXCEPTION 'runtime can create schema objects';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    IF pg_has_role(current_user, 'onlineconf_migrator', 'MEMBER') THEN
        RAISE EXCEPTION 'runtime can assume migrator privileges';
    END IF;
END;
$$;

ROLLBACK;
