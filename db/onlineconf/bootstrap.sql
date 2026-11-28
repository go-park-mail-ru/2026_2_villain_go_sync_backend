BEGIN;

LOCK TABLE public.admin IN SHARE ROW EXCLUSIVE MODE;

DO $$
BEGIN
    IF EXISTS (SELECT FROM public.admin)
       OR EXISTS (SELECT FROM public.change_history WHERE resource_type = 'admin') THEN
        RAISE EXCEPTION 'OnlineConf already has or previously had administrators';
    END IF;
END;
$$;

WITH first_owner AS (
    INSERT INTO public.admin (user_id, role)
    VALUES (:'owner_id'::BIGINT, 'owner')
    RETURNING user_id, role, granted_by, created_at
)
INSERT INTO public.change_history
    (actor_user_id, resource_type, resource_id, resource_name, action, new_value)
SELECT owner.user_id, 'admin', owner.user_id, owner.user_id::TEXT, 'create', to_jsonb(owner)
FROM first_owner owner;

COMMIT;
