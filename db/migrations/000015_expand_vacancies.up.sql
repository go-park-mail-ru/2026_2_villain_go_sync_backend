BEGIN;

ALTER TABLE vacancy
    ADD COLUMN apply_until DATE;

ALTER TABLE vacancy_history
    ADD COLUMN responsibilities TEXT,
    ADD COLUMN requirements TEXT,
    ADD COLUMN offer TEXT,
    ADD COLUMN experience_level TEXT,
    ADD COLUMN employment_type TEXT,
    ADD COLUMN work_format TEXT,
    ADD COLUMN work_schedule TEXT,
    ADD COLUMN work_hours TEXT,
    ADD COLUMN city TEXT,
    ADD COLUMN office_address TEXT,
    ADD CONSTRAINT vacancy_history_title_check
        CHECK (length(btrim(title)) > 0) NOT VALID,
    ADD CONSTRAINT vacancy_history_experience_level_check
        CHECK (experience_level IS NULL OR experience_level IN ('no_experience', '1_3', '3_6', '6_plus')),
    ADD CONSTRAINT vacancy_history_employment_type_check
        CHECK (employment_type IS NULL OR employment_type IN ('full', 'part')),
    ADD CONSTRAINT vacancy_history_work_format_check
        CHECK (work_format IS NULL OR work_format IN ('office', 'hybrid', 'remote')),
    ADD CONSTRAINT vacancy_history_required_fields_check
        CHECK (
            responsibilities IS NOT NULL
            AND responsibilities ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
            AND requirements IS NOT NULL
            AND requirements ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
        ) NOT VALID,
    ADD CONSTRAINT vacancy_history_work_schedule_check
        CHECK (work_schedule IS NULL OR work_schedule IN ('5_2', '2_2', 'flexible', 'shift'));

CREATE TABLE vacancy_version_skill (
    skill_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    vacancy_id BIGINT NOT NULL,
    version INT NOT NULL,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vacancy_version_skill_history_fk
        FOREIGN KEY (vacancy_id, version)
        REFERENCES vacancy_history (vacancy_id, version) ON DELETE CASCADE,
    CONSTRAINT vacancy_version_skill_position_unique
        UNIQUE (vacancy_id, version, position)
);

CREATE UNIQUE INDEX vacancy_version_skill_name_unique_idx
    ON vacancy_version_skill (vacancy_id, version, lower(btrim(name)));


COMMIT;
