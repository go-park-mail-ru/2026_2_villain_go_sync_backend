DO $$
DECLARE
    missing_count INT;
BEGIN
    SELECT count(*) INTO missing_count
    FROM (VALUES
      ('vacancy', 'apply_until'),
      ('vacancy_history', 'responsibilities'),
      ('vacancy_history', 'requirements'),
      ('vacancy_history', 'offer'),
      ('vacancy_history', 'experience_level'),
      ('vacancy_history', 'employment_type'),
      ('vacancy_history', 'work_format'),
      ('vacancy_history', 'work_schedule'),
      ('vacancy_history', 'work_hours'),
      ('vacancy_history', 'city'),
      ('vacancy_history', 'office_address'),
      ('vacancy_version_skill', 'vacancy_id'),
      ('vacancy_version_skill', 'version'),
      ('vacancy_version_skill', 'name'),
      ('vacancy_version_skill', 'position')
    ) AS expected(table_name, column_name)
    WHERE NOT EXISTS (
        SELECT 1 FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.table_name = expected.table_name
          AND c.column_name = expected.column_name
    );
    IF missing_count > 0 THEN
        RAISE EXCEPTION '000015: % required columns are missing', missing_count;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'vacancy_version_skill_history_fk'
    ) THEN
        RAISE EXCEPTION '000015: missing vacancy_version_skill FK';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE schemaname = 'public'
          AND indexname = 'vacancy_version_skill_name_unique_idx'
    ) THEN
        RAISE EXCEPTION '000015: missing unique index for skill names';
    END IF;
    RAISE NOTICE '000015 structure: OK';
END $$;

SELECT count(*) AS vacancy_versions FROM vacancy_history;
SELECT count(*) AS vacancy_skills FROM vacancy_version_skill;

SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name IN ('vacancy', 'vacancy_history', 'vacancy_version_skill')
ORDER BY table_name, ordinal_position;

SELECT conname, contype, convalidated
FROM pg_constraint
WHERE conname IN (
    'vacancy_history_required_fields_check',
    'vacancy_history_title_check',
    'vacancy_version_skill_history_fk',
    'vacancy_version_skill_position_unique'
)
ORDER BY conname;
