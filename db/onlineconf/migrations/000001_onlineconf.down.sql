BEGIN;

LOCK TABLE public.scope, public.config_key, public.config, public.feature_flag,
    public.feature_flag_override, public.admin, public.change_history IN ACCESS EXCLUSIVE MODE;

DO $$
BEGIN
    IF EXISTS (SELECT FROM public.config_key)
       OR EXISTS (SELECT FROM public.admin)
       OR EXISTS (SELECT FROM public.change_history)
       OR (SELECT count(*) FROM public.scope) <> 5
       OR EXISTS (
           SELECT FROM public.scope
           WHERE name NOT IN ('frontend', 'auth', 'vacancy', 'seeker', 'application')
              OR description IS NOT NULL OR updated_at IS DISTINCT FROM created_at
       ) THEN
        RAISE EXCEPTION 'refusing to roll back OnlineConf with non-seed data';
    END IF;
END;
$$;

DROP TABLE public.change_history;
DROP TABLE public.admin;
DROP TABLE public.feature_flag_override;
DROP TABLE public.feature_flag;
DROP TABLE public.config;
DROP TABLE public.config_key;
DROP TABLE public.scope;

DROP FUNCTION public.onlineconf_forbid_subtype_truncate();
DROP FUNCTION public.onlineconf_enforce_key_subtype();
DROP FUNCTION public.onlineconf_check_key_subtype(BIGINT);
DROP FUNCTION public.onlineconf_keep_key_identity();
DROP FUNCTION public.onlineconf_keep_admin_owner();
DROP FUNCTION public.onlineconf_touch_parent_key();
DROP FUNCTION public.onlineconf_touch_updated_at();

COMMIT;
