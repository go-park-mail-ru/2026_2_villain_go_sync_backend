SELECT COUNT(*) AS table_count
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
  AND table_name <> 'schema_migrations';

SELECT conname, contype, convalidated
FROM pg_constraint
WHERE conname IN (
    'application_vacancy_seeker_unique',
    'application_vacancy_version_fk',
    'application_seeker_fk',
    'application_rejection_required_check',
    'application_rejection_code_check',
    'application_rejection_status_check',
    'resume_required_fields_check',
    'vacancy_history_required_fields_check'
)
ORDER BY conname;

SELECT COUNT(*) AS missing_or_unsealed_snapshot_count
FROM application a
LEFT JOIN application_resume_snapshot s ON s.application_id = a.application_id
WHERE s.application_id IS NULL OR NOT s.is_sealed;

SELECT COUNT(*) AS duplicate_application_count
FROM (
    SELECT vacancy_id, seeker_id
    FROM application
    GROUP BY vacancy_id, seeker_id
    HAVING COUNT(*) > 1
) d;

SELECT table_name, column_name
FROM information_schema.columns
WHERE table_schema = 'public'
  AND udt_name IN ('json', 'jsonb', '_text', 'varchar')
ORDER BY table_name, column_name;
