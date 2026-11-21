BEGIN;

ALTER TABLE seeker_profile
    ADD COLUMN city TEXT,
    ADD COLUMN messenger TEXT,
    ADD COLUMN contact_email TEXT,
    ADD COLUMN preferred_position TEXT,
    ADD COLUMN preferred_cities TEXT,
    ADD COLUMN work_format TEXT,
    ADD CONSTRAINT seeker_profile_work_format_check
        CHECK (work_format IS NULL OR work_format IN ('office', 'hybrid', 'remote')),
    ADD CONSTRAINT seeker_profile_last_name_required_check
        CHECK (last_name IS NOT NULL AND last_name ~ U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]') NOT VALID;

CREATE TABLE seeker_experience (
    experience_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seeker_id BIGINT NOT NULL REFERENCES seeker_profile(user_id) ON DELETE CASCADE,
    company_name TEXT NOT NULL CHECK (length(btrim(company_name)) > 0),
    position_title TEXT NOT NULL CHECK (length(btrim(position_title)) > 0),
    started_on DATE,
    ended_on DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    achievements TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT seeker_experience_position_unique UNIQUE (seeker_id, position),
    CONSTRAINT seeker_experience_dates_check CHECK (
        (started_on IS NULL OR ended_on IS NULL OR ended_on >= started_on)
        AND (NOT is_current OR ended_on IS NULL)
    )
);

CREATE TABLE seeker_education (
    education_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seeker_id BIGINT NOT NULL REFERENCES seeker_profile(user_id) ON DELETE CASCADE,
    record_type TEXT,
    institution TEXT,
    completion_year SMALLINT,
    specialization TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT seeker_education_position_unique UNIQUE (seeker_id, position)
);

CREATE TABLE seeker_skill (
    skill_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seeker_id BIGINT NOT NULL REFERENCES seeker_profile(user_id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT seeker_skill_position_unique UNIQUE (seeker_id, position)
);
CREATE UNIQUE INDEX seeker_skill_name_unique_idx
    ON seeker_skill (seeker_id, (lower(btrim(name))));

CREATE TABLE seeker_project (
    project_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seeker_id BIGINT NOT NULL REFERENCES seeker_profile(user_id) ON DELETE CASCADE,
    title TEXT,
    url TEXT,
    description TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 5),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT seeker_project_position_unique UNIQUE (seeker_id, position),
    CONSTRAINT seeker_project_not_empty_check CHECK (
        length(btrim(coalesce(title, ''))) > 0 OR
        length(btrim(coalesce(url, ''))) > 0 OR
        length(btrim(coalesce(description, ''))) > 0
    )
);

ALTER TABLE employer_profile
    ADD COLUMN industry TEXT,
    ADD COLUMN team_size TEXT,
    ADD COLUMN main_city TEXT,
    ADD COLUMN benefits TEXT,
    ADD COLUMN hiring_email TEXT,
    ADD COLUMN office_address TEXT;

ALTER TABLE employer_profile
    ADD CONSTRAINT employer_profile_company_name_check
        CHECK (length(btrim(company_name)) > 0) NOT VALID;

CREATE TABLE employer_technology (
    technology_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employer_id BIGINT NOT NULL REFERENCES employer_profile(user_id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    position INTEGER NOT NULL CHECK (position > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employer_technology_position_unique UNIQUE (employer_id, position)
);
CREATE UNIQUE INDEX employer_technology_name_unique_idx
    ON employer_technology (employer_id, (lower(btrim(name))));

CREATE TABLE employer_culture (
    culture_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employer_id BIGINT NOT NULL REFERENCES employer_profile(user_id) ON DELETE CASCADE,
    title TEXT,
    short_description TEXT,
    description TEXT,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT employer_culture_position_unique UNIQUE (employer_id, position)
);

COMMIT;
