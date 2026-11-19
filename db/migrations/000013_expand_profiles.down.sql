BEGIN;

DROP TABLE IF EXISTS employer_culture;
DROP TABLE IF EXISTS employer_technology;

ALTER TABLE employer_profile
    DROP CONSTRAINT IF EXISTS employer_profile_company_name_check,
    DROP COLUMN IF EXISTS office_address,
    DROP COLUMN IF EXISTS hiring_email,
    DROP COLUMN IF EXISTS benefits,
    DROP COLUMN IF EXISTS main_city,
    DROP COLUMN IF EXISTS team_size,
    DROP COLUMN IF EXISTS industry;

DROP TABLE IF EXISTS seeker_project;
DROP TABLE IF EXISTS seeker_skill;
DROP TABLE IF EXISTS seeker_education;
DROP TABLE IF EXISTS seeker_experience;

ALTER TABLE seeker_profile
    DROP CONSTRAINT IF EXISTS seeker_profile_work_format_check,
    DROP CONSTRAINT IF EXISTS seeker_profile_last_name_required_check,
    DROP COLUMN IF EXISTS work_format,
    DROP COLUMN IF EXISTS preferred_cities,
    DROP COLUMN IF EXISTS preferred_position,
    DROP COLUMN IF EXISTS contact_email,
    DROP COLUMN IF EXISTS messenger,
    DROP COLUMN IF EXISTS city;

COMMIT;
