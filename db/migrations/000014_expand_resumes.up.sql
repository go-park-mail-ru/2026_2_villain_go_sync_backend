BEGIN;

ALTER TABLE resume
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN target_position TEXT,
    ADD COLUMN city TEXT,
    ADD COLUMN employment_type TEXT,
    ADD COLUMN work_format TEXT,
    ADD COLUMN work_schedule TEXT,
    ADD COLUMN salary_from NUMERIC(12, 2),
    ADD COLUMN portfolio_url TEXT,
    ADD COLUMN languages TEXT,
    ADD CONSTRAINT resume_required_fields_check
        CHECK (
            length(btrim(title)) > 0
            AND target_position IS NOT NULL
            AND target_position ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'
        ) NOT VALID,
    ADD CONSTRAINT resume_employment_type_check
        CHECK (employment_type IS NULL OR employment_type IN ('full', 'part')),
    ADD CONSTRAINT resume_work_format_check
        CHECK (work_format IS NULL OR work_format IN ('office', 'hybrid', 'remote')),
    ADD CONSTRAINT resume_work_schedule_check
        CHECK (work_schedule IS NULL OR work_schedule IN ('5_2', '2_2', 'flexible', 'shift')),
    ADD CONSTRAINT resume_salary_from_check
        CHECK (salary_from IS NULL OR salary_from >= 0);


CREATE TABLE resume_legacy_content (
    resume_id BIGINT PRIMARY KEY REFERENCES resume (resume_id) ON DELETE CASCADE,
    experience_text TEXT,
    education_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO resume_legacy_content (resume_id, experience_text, education_text)
SELECT resume_id, experience, education
FROM resume
WHERE experience IS NOT NULL OR education IS NOT NULL;

ALTER TABLE resume
    DROP COLUMN experience,
    DROP COLUMN education;

CREATE TABLE resume_experience (
    experience_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id BIGINT NOT NULL
        REFERENCES resume (resume_id) ON DELETE CASCADE,
    company_name TEXT,
    position_title TEXT,
    started_on DATE,
    ended_on DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    achievements TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT resume_experience_position_unique UNIQUE (resume_id, position),
    CONSTRAINT resume_experience_dates_check CHECK (
        (started_on IS NULL OR ended_on IS NULL OR ended_on >= started_on)
        AND (NOT is_current OR ended_on IS NULL)
    )
);

CREATE TABLE resume_education (
    education_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id BIGINT NOT NULL
        REFERENCES resume (resume_id) ON DELETE CASCADE,
    record_type TEXT,
    institution TEXT,
    completion_year SMALLINT,
    specialization TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT resume_education_position_unique UNIQUE (resume_id, position)
);

CREATE TABLE resume_skill (
    skill_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id BIGINT NOT NULL
        REFERENCES resume (resume_id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT resume_skill_position_unique UNIQUE (resume_id, position)
);
CREATE UNIQUE INDEX resume_skill_name_unique_idx
    ON resume_skill (resume_id, lower(btrim(name)));

CREATE TABLE resume_project (
    project_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id BIGINT NOT NULL
        REFERENCES resume (resume_id) ON DELETE CASCADE,
    title TEXT,
    url TEXT,
    description TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT resume_project_position_unique UNIQUE (resume_id, position),
    CONSTRAINT resume_project_not_empty_check CHECK (
        length(btrim(coalesce(title, ''))) > 0 OR
        length(btrim(coalesce(url, ''))) > 0 OR
        length(btrim(coalesce(description, ''))) > 0
    )
);


COMMIT;
