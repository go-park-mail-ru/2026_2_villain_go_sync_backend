ALTER TABLE public.user_group ADD COLUMN revision BIGINT NOT NULL DEFAULT 1 CHECK (revision > 0);
