BEGIN;
LOCK TABLE public.flag_rollout, public.rollout_policy, public.access_grant, public.group_member, public.user_group
    IN ACCESS EXCLUSIVE MODE;
DO $$
BEGIN
    IF EXISTS (SELECT FROM public.flag_rollout) OR EXISTS (SELECT FROM public.rollout_policy)
       OR EXISTS (SELECT FROM public.access_grant) OR EXISTS (SELECT FROM public.user_group) THEN
        RAISE EXCEPTION 'refusing to roll back advanced OnlineConf with stored rollouts, policies or access rules';
    END IF;
END;
$$;
DROP TABLE public.flag_rollout, public.rollout_policy, public.access_grant, public.group_member, public.user_group;
COMMIT;
