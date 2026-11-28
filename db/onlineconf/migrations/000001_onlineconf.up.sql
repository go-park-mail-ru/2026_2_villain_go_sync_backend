BEGIN;

CREATE TABLE public.scope (
    scope_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT scope_name_format CHECK (name COLLATE "C" ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE public.config_key (
    key_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    scope_id BIGINT NOT NULL REFERENCES public.scope (scope_id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    owner_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT config_key_scope_name_unique UNIQUE (scope_id, name),
    CONSTRAINT config_key_name_format CHECK (name COLLATE "C" ~ '^[a-z][a-z0-9_]*$'),
    CONSTRAINT config_key_type_valid CHECK (type IN ('config', 'feature_flag')),
    CONSTRAINT config_key_owner_positive CHECK (owner_user_id > 0)
);

CREATE TABLE public.config (
    key_id BIGINT PRIMARY KEY REFERENCES public.config_key (key_id) ON DELETE CASCADE,
    value JSONB NOT NULL
);

CREATE TABLE public.feature_flag (
    key_id BIGINT PRIMARY KEY REFERENCES public.config_key (key_id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    rollout_percentage SMALLINT NOT NULL DEFAULT 0,
    CONSTRAINT feature_flag_rollout_range CHECK (rollout_percentage BETWEEN 0 AND 100)
);

CREATE TABLE public.feature_flag_override (
    key_id BIGINT REFERENCES public.feature_flag (key_id) ON DELETE CASCADE,
    user_id BIGINT,
    enabled BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (key_id, user_id),
    CONSTRAINT feature_flag_override_user_positive CHECK (user_id > 0)
);

CREATE INDEX feature_flag_override_user_key_idx
    ON public.feature_flag_override (user_id, key_id);

CREATE TABLE public.admin (
    user_id BIGINT PRIMARY KEY,
    role TEXT NOT NULL,
    granted_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT admin_user_positive CHECK (user_id > 0),
    CONSTRAINT admin_role_valid CHECK (role IN ('owner', 'admin')),
    CONSTRAINT admin_grant_valid CHECK (
        (granted_by IS NULL AND role = 'owner') OR (granted_by IS NOT NULL AND granted_by > 0)
    )
);

CREATE UNIQUE INDEX admin_initial_owner_idx ON public.admin ((granted_by IS NULL))
    WHERE granted_by IS NULL;

CREATE TABLE public.change_history (
    change_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_user_id BIGINT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id BIGINT NOT NULL,
    scope_name TEXT,
    resource_name TEXT NOT NULL,
    action TEXT NOT NULL,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT change_history_actor_positive CHECK (actor_user_id > 0),
    CONSTRAINT change_history_resource_positive CHECK (resource_id > 0),
    CONSTRAINT change_history_resource_name_nonempty CHECK (btrim(resource_name) <> ''),
    CONSTRAINT change_history_resource_type_valid CHECK (
        resource_type IN ('scope', 'config', 'feature_flag', 'feature_flag_override', 'admin')
    ),
    CONSTRAINT change_history_action_valid CHECK (action IN ('create', 'update', 'delete'))
);

CREATE INDEX change_history_resource_idx
    ON public.change_history (resource_type, resource_id, change_id DESC);

CREATE INDEX change_history_scope_idx
    ON public.change_history (scope_name, change_id DESC)
    WHERE scope_name IS NOT NULL;

CREATE FUNCTION public.onlineconf_touch_updated_at() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    NEW.updated_at := GREATEST(OLD.updated_at, clock_timestamp());
    RETURN NEW;
END;
$$;

CREATE TRIGGER scope_touch_updated_at
BEFORE UPDATE ON public.scope FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_updated_at();

CREATE TRIGGER config_key_touch_updated_at
BEFORE UPDATE ON public.config_key FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_updated_at();

CREATE TRIGGER feature_flag_override_touch_updated_at
BEFORE UPDATE ON public.feature_flag_override FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_updated_at();

CREATE FUNCTION public.onlineconf_touch_parent_key() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
    old_key_id BIGINT;
    new_key_id BIGINT;
BEGIN
    IF TG_OP = 'UPDATE' AND OLD IS NOT DISTINCT FROM NEW THEN
        RETURN NULL;
    END IF;
    IF TG_OP <> 'INSERT' THEN
        old_key_id := OLD.key_id;
    END IF;
    IF TG_OP <> 'DELETE' THEN
        new_key_id := NEW.key_id;
    END IF;

    UPDATE public.config_key SET updated_at = config_key.updated_at
    WHERE key_id IN (old_key_id, new_key_id);
    RETURN NULL;
END;
$$;

CREATE TRIGGER config_touch_parent_key
AFTER INSERT OR UPDATE OR DELETE ON public.config
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_parent_key();

CREATE TRIGGER feature_flag_touch_parent_key
AFTER INSERT OR UPDATE OR DELETE ON public.feature_flag
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_parent_key();

CREATE TRIGGER feature_flag_override_touch_parent_key
AFTER INSERT OR UPDATE OR DELETE ON public.feature_flag_override
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_touch_parent_key();

CREATE FUNCTION public.onlineconf_keep_admin_owner() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'TRUNCATE' THEN
        IF EXISTS (SELECT FROM public.admin WHERE role = 'owner') THEN
            RAISE EXCEPTION 'cannot remove the last OnlineConf owner' USING ERRCODE = 'check_violation';
        END IF;
    ELSE
        PERFORM 1 FROM public.admin WHERE role = 'owner' LIMIT 1 FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'cannot remove the last OnlineConf owner' USING ERRCODE = 'check_violation';
        END IF;
    END IF;
    RETURN NULL;
END;
$$;

CREATE TRIGGER admin_keep_owner_on_delete
AFTER DELETE ON public.admin
FOR EACH ROW WHEN (OLD.role = 'owner') EXECUTE FUNCTION public.onlineconf_keep_admin_owner();

CREATE TRIGGER admin_keep_owner_on_demotion
AFTER UPDATE OF role ON public.admin
FOR EACH ROW WHEN (OLD.role = 'owner' AND NEW.role <> 'owner')
EXECUTE FUNCTION public.onlineconf_keep_admin_owner();

CREATE TRIGGER admin_keep_owner_on_truncate
BEFORE TRUNCATE ON public.admin FOR EACH STATEMENT EXECUTE FUNCTION public.onlineconf_keep_admin_owner();

CREATE FUNCTION public.onlineconf_keep_key_identity() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    IF NEW.type IS DISTINCT FROM OLD.type OR NEW.key_id IS DISTINCT FROM OLD.key_id THEN
        RAISE EXCEPTION 'config_key id and type are immutable' USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER config_key_keep_identity
BEFORE UPDATE ON public.config_key FOR EACH ROW EXECUTE FUNCTION public.onlineconf_keep_key_identity();

CREATE FUNCTION public.onlineconf_check_key_subtype(candidate_id BIGINT) RETURNS void
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
DECLARE
    key_type TEXT;
    has_config BOOLEAN;
    has_flag BOOLEAN;
BEGIN
    SELECT type INTO key_type
    FROM public.config_key WHERE key_id = candidate_id
    FOR NO KEY UPDATE;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    SELECT EXISTS (SELECT FROM public.config WHERE key_id = candidate_id),
           EXISTS (SELECT FROM public.feature_flag WHERE key_id = candidate_id)
    INTO has_config, has_flag;

    IF has_config IS DISTINCT FROM (key_type = 'config')
       OR has_flag IS DISTINCT FROM (key_type = 'feature_flag') THEN
        RAISE EXCEPTION 'config_key % must have exactly one subtype matching type %', candidate_id, key_type
            USING ERRCODE = 'check_violation', CONSTRAINT = 'config_key_exact_subtype';
    END IF;
END;
$$;

CREATE FUNCTION public.onlineconf_enforce_key_subtype() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        PERFORM public.onlineconf_check_key_subtype(OLD.key_id);
    ELSE
        PERFORM public.onlineconf_check_key_subtype(NEW.key_id);
        IF TG_OP = 'UPDATE' AND OLD.key_id IS DISTINCT FROM NEW.key_id THEN
            PERFORM public.onlineconf_check_key_subtype(OLD.key_id);
        END IF;
    END IF;
    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER config_key_exact_subtype
AFTER INSERT ON public.config_key
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_enforce_key_subtype();

CREATE CONSTRAINT TRIGGER config_exact_subtype
AFTER INSERT OR DELETE OR UPDATE OF key_id ON public.config
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_enforce_key_subtype();

CREATE CONSTRAINT TRIGGER feature_flag_exact_subtype
AFTER INSERT OR DELETE OR UPDATE OF key_id ON public.feature_flag
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION public.onlineconf_enforce_key_subtype();

CREATE FUNCTION public.onlineconf_forbid_subtype_truncate() RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    RAISE EXCEPTION 'delete config_key rows to remove their subtypes'
        USING ERRCODE = 'feature_not_supported';
END;
$$;

CREATE TRIGGER config_forbid_truncate
BEFORE TRUNCATE ON public.config EXECUTE FUNCTION public.onlineconf_forbid_subtype_truncate();

CREATE TRIGGER feature_flag_forbid_truncate
BEFORE TRUNCATE ON public.feature_flag EXECUTE FUNCTION public.onlineconf_forbid_subtype_truncate();

GRANT SELECT, INSERT, UPDATE, DELETE ON
    public.scope, public.config_key, public.config, public.feature_flag,
    public.feature_flag_override, public.admin
    TO onlineconf_runtime;

GRANT SELECT, INSERT ON public.change_history TO onlineconf_runtime;
GRANT USAGE ON SEQUENCE public.scope_scope_id_seq, public.config_key_key_id_seq, public.change_history_change_id_seq
    TO onlineconf_runtime;

INSERT INTO public.scope (name) VALUES
    ('frontend'), ('auth'), ('vacancy'), ('seeker'), ('application');

COMMIT;
