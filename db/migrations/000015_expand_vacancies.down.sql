BEGIN;

DROP TABLE IF EXISTS vacancy_version_skill;

ALTER TABLE vacancy_history
    DROP CONSTRAINT IF EXISTS vacancy_history_title_check,
    DROP CONSTRAINT IF EXISTS vacancy_history_required_fields_check,
    DROP CONSTRAINT IF EXISTS vacancy_history_work_schedule_check,
    DROP CONSTRAINT IF EXISTS vacancy_history_work_format_check,
    DROP CONSTRAINT IF EXISTS vacancy_history_employment_type_check,
    DROP CONSTRAINT IF EXISTS vacancy_history_experience_level_check,
    DROP COLUMN IF EXISTS office_address,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS work_hours,
    DROP COLUMN IF EXISTS work_schedule,
    DROP COLUMN IF EXISTS work_format,
    DROP COLUMN IF EXISTS employment_type,
    DROP COLUMN IF EXISTS experience_level,
    DROP COLUMN IF EXISTS offer,
    DROP COLUMN IF EXISTS requirements,
    DROP COLUMN IF EXISTS responsibilities;

ALTER TABLE vacancy
    DROP COLUMN IF EXISTS apply_until;

COMMIT;
