BEGIN;

CREATE TABLE public.user_group (
    group_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (char_length(name) BETWEEN 1 AND 100)
);

CREATE TABLE public.group_member (
    group_id BIGINT NOT NULL REFERENCES public.user_group (group_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL CHECK (user_id > 0),
    PRIMARY KEY (group_id, user_id)
);
CREATE INDEX group_member_by_user ON public.group_member (user_id, group_id);

CREATE TABLE public.access_grant (
    grant_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT CHECK (user_id > 0),
    group_id BIGINT REFERENCES public.user_group (group_id) ON DELETE CASCADE,
    scope_id BIGINT REFERENCES public.scope (scope_id) ON DELETE CASCADE,
    key_type TEXT CHECK (key_type IN ('config', 'feature_flag')),
    key_id BIGINT REFERENCES public.config_key (key_id) ON DELETE CASCADE,
    action TEXT NOT NULL CHECK (action IN ('read', 'create', 'update', 'delete', 'approve')),
    CHECK (num_nonnulls(user_id, group_id) = 1),
    CHECK (
        (scope_id IS NOT NULL AND key_type IS NOT NULL AND key_id IS NULL) OR
        (scope_id IS NULL AND key_type IS NULL AND key_id IS NOT NULL AND action <> 'create')
    ),
    CHECK (action <> 'approve' OR key_type IS DISTINCT FROM 'config'),
    UNIQUE NULLS NOT DISTINCT (user_id, group_id, scope_id, key_type, key_id, action)
);
CREATE INDEX access_grant_by_user ON public.access_grant (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX access_grant_by_group ON public.access_grant (group_id) WHERE group_id IS NOT NULL;

CREATE TABLE public.rollout_policy (
    policy_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (char_length(name) BETWEEN 1 AND 100),
    steps JSONB NOT NULL CHECK (jsonb_typeof(steps) = 'array')
);

CREATE TABLE public.flag_rollout (
    key_id BIGINT PRIMARY KEY REFERENCES public.feature_flag (key_id) ON DELETE CASCADE,
    run_id BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,
    policy_name TEXT NOT NULL,
    steps JSONB NOT NULL CHECK (jsonb_typeof(steps) = 'array'),
    status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'canceled')),
    next_step INTEGER NOT NULL CHECK (next_step >= 0),
    due_at TIMESTAMPTZ,
    approved_by BIGINT CHECK (approved_by > 0),
    CHECK ((status = 'running') = (due_at IS NOT NULL)),
    CHECK (status = 'running' OR approved_by IS NULL)
);
CREATE INDEX flag_rollout_ready ON public.flag_rollout (due_at, key_id)
    WHERE status = 'running' AND approved_by IS NOT NULL;

GRANT SELECT, INSERT, UPDATE, DELETE ON public.user_group, public.group_member,
    public.access_grant, public.rollout_policy, public.flag_rollout TO onlineconf_runtime;
GRANT USAGE ON SEQUENCE public.user_group_group_id_seq, public.access_grant_grant_id_seq,
    public.rollout_policy_policy_id_seq, public.flag_rollout_run_id_seq TO onlineconf_runtime;

COMMIT;
