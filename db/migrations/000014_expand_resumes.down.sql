BEGIN;

DROP TABLE IF EXISTS resume_project;
DROP TABLE IF EXISTS resume_skill;
DROP TABLE IF EXISTS resume_education;
DROP TABLE IF EXISTS resume_experience;


ALTER TABLE resume
    ADD COLUMN experience TEXT,
    ADD COLUMN education TEXT;

UPDATE resume r
SET experience = l.experience_text,
    education = l.education_text
FROM resume_legacy_content l
WHERE l.resume_id = r.resume_id;

DROP TABLE resume_legacy_content;

ALTER TABLE resume
    DROP CONSTRAINT IF EXISTS resume_required_fields_check,
    DROP CONSTRAINT IF EXISTS resume_salary_from_check,
    DROP CONSTRAINT IF EXISTS resume_work_schedule_check,
    DROP CONSTRAINT IF EXISTS resume_work_format_check,
    DROP CONSTRAINT IF EXISTS resume_employment_type_check,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS languages,
    DROP COLUMN IF EXISTS portfolio_url,
    DROP COLUMN IF EXISTS salary_from,
    DROP COLUMN IF EXISTS work_schedule,
    DROP COLUMN IF EXISTS work_format,
    DROP COLUMN IF EXISTS employment_type,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS target_position;

COMMIT;
