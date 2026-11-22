BEGIN;

DO $$
DECLARE
    duplicate_vacancy BIGINT;
    duplicate_seeker BIGINT;
BEGIN
    SELECT a.vacancy_id, r.seeker_id
    INTO duplicate_vacancy, duplicate_seeker
    FROM application a
    JOIN resume r ON r.resume_id = a.resume_id
    GROUP BY a.vacancy_id, r.seeker_id
    HAVING COUNT(*) > 1
    LIMIT 1;

    IF FOUND THEN
        RAISE EXCEPTION 'Duplicate applications: vacancy_id=%, seeker_id=%',
            duplicate_vacancy, duplicate_seeker;
    END IF;
END;
$$;

ALTER TABLE application_status_history
    DROP CONSTRAINT application_status_history_status_check;

ALTER TABLE application_status_history
    ADD CONSTRAINT application_status_history_status_check
        CHECK (status IN (
            'created',
            'reviewing',
            'interview',
            'offer',
            'rejected'
        )),
    ADD COLUMN reason_code TEXT,
    ADD COLUMN changed_by BIGINT
        REFERENCES app_user (user_id) ON DELETE SET NULL,
    ADD CONSTRAINT application_rejection_code_check
        CHECK (reason_code IS NULL OR reason_code IN (
            'insufficient_experience',
            'skills_mismatch',
            'salary_mismatch',
            'work_format_mismatch',
            'location_mismatch',
            'candidate_declined',
            'other'
        )),
    ADD CONSTRAINT application_rejection_status_check
        CHECK (reason_code IS NULL OR status = 'rejected');

ALTER TABLE application_status_history
    ADD CONSTRAINT application_rejection_required_check
        CHECK (status <> 'rejected' OR reason_code IS NOT NULL) NOT VALID;

ALTER TABLE application
    ADD COLUMN seeker_id BIGINT;

UPDATE application a
SET seeker_id = r.seeker_id
FROM resume r
WHERE a.resume_id = r.resume_id;

ALTER TABLE application
    ALTER COLUMN seeker_id SET NOT NULL,
    ADD CONSTRAINT application_seeker_fk
        FOREIGN KEY (seeker_id)
        REFERENCES seeker_profile (user_id) ON DELETE RESTRICT;

ALTER TABLE application
    DROP CONSTRAINT application_vacancy_version_fk;

ALTER TABLE application
    ADD CONSTRAINT application_vacancy_version_fk
        FOREIGN KEY (vacancy_id, vacancy_version)
        REFERENCES vacancy_history (vacancy_id, version) ON DELETE CASCADE;

CREATE TABLE application_resume_snapshot (
    application_id BIGINT PRIMARY KEY
        REFERENCES application (application_id) ON DELETE CASCADE,
    source_resume_id BIGINT
        REFERENCES resume (resume_id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    legacy_experience_text TEXT,
    legacy_education_text TEXT,
    target_position TEXT,
    city TEXT,
    employment_type TEXT,
    work_format TEXT,
    work_schedule TEXT,
    salary_from NUMERIC(12, 2),
    portfolio_url TEXT,
    languages TEXT,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_reconstructed BOOLEAN NOT NULL DEFAULT FALSE,
    is_sealed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE application_snapshot_experience (
    application_id BIGINT NOT NULL
        REFERENCES application_resume_snapshot (application_id) ON DELETE CASCADE,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    company_name TEXT,
    position_title TEXT,
    started_on DATE,
    ended_on DATE,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    achievements TEXT,
    PRIMARY KEY (application_id, position),
    CONSTRAINT application_snapshot_experience_dates_check
        CHECK (
            (started_on IS NULL OR ended_on IS NULL OR ended_on >= started_on)
            AND (NOT is_current OR ended_on IS NULL)
        )
);

CREATE TABLE application_snapshot_education (
    application_id BIGINT NOT NULL
        REFERENCES application_resume_snapshot (application_id) ON DELETE CASCADE,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 10),
    record_type TEXT,
    institution TEXT,
    completion_year SMALLINT,
    specialization TEXT,
    PRIMARY KEY (application_id, position)
);

CREATE TABLE application_snapshot_skill (
    application_id BIGINT NOT NULL
        REFERENCES application_resume_snapshot (application_id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position > 0),
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    PRIMARY KEY (application_id, position)
);

CREATE TABLE application_snapshot_project (
    application_id BIGINT NOT NULL
        REFERENCES application_resume_snapshot (application_id) ON DELETE CASCADE,
    position SMALLINT NOT NULL CHECK (position BETWEEN 1 AND 5),
    title TEXT,
    url TEXT,
    description TEXT,
    PRIMARY KEY (application_id, position)
);

CREATE TABLE application_snapshot_category (
    application_id BIGINT NOT NULL
        REFERENCES application_resume_snapshot (application_id) ON DELETE CASCADE,
    category_name TEXT NOT NULL CHECK (length(btrim(category_name)) > 0),
    PRIMARY KEY (application_id, category_name)
);

INSERT INTO application_resume_snapshot (
    application_id, source_resume_id, title, description,
    legacy_experience_text, legacy_education_text,
    target_position, city, employment_type, work_format,
    work_schedule, salary_from, portfolio_url, languages,
    is_reconstructed
)
SELECT
    a.application_id, r.resume_id, r.title, r.description,
    l.experience_text, l.education_text,
    r.target_position, r.city, r.employment_type, r.work_format,
    r.work_schedule, r.salary_from, r.portfolio_url, r.languages,
    TRUE
FROM application a
JOIN resume r ON r.resume_id = a.resume_id
LEFT JOIN resume_legacy_content l ON l.resume_id = r.resume_id;

INSERT INTO application_snapshot_experience (
    application_id, position, company_name, position_title,
    started_on, ended_on, is_current, achievements
)
SELECT a.application_id, e.position, e.company_name, e.position_title,
       e.started_on, e.ended_on, e.is_current, e.achievements
FROM application a
JOIN resume_experience e ON e.resume_id = a.resume_id;

INSERT INTO application_snapshot_education (
    application_id, position, record_type, institution, completion_year, specialization
)
SELECT a.application_id, e.position, e.record_type, e.institution,
       e.completion_year, e.specialization
FROM application a
JOIN resume_education e ON e.resume_id = a.resume_id;

INSERT INTO application_snapshot_skill (application_id, position, name)
SELECT a.application_id, s.position, s.name
FROM application a
JOIN resume_skill s ON s.resume_id = a.resume_id;

INSERT INTO application_snapshot_project (application_id, position, title, url, description)
SELECT a.application_id, p.position, p.title, p.url, p.description
FROM application a
JOIN resume_project p ON p.resume_id = a.resume_id;

INSERT INTO application_snapshot_category (application_id, category_name)
SELECT a.application_id, c.name
FROM application a
JOIN resume_category rc ON rc.resume_id = a.resume_id
JOIN category c ON c.category_id = rc.category_id;

UPDATE application_resume_snapshot
   SET is_sealed = TRUE;

INSERT INTO application_status_history (application_id, status, changed_at)
SELECT a.application_id, 'created', a.created_at
FROM application a
WHERE NOT EXISTS (
    SELECT 1
    FROM application_status_history h
    WHERE h.application_id = a.application_id
);

ALTER TABLE application
    DROP CONSTRAINT application_vacancy_resume_unique,
    DROP CONSTRAINT application_resume_fk,
    DROP COLUMN resume_id,
    ADD CONSTRAINT application_vacancy_seeker_unique
        UNIQUE (vacancy_id, seeker_id);

CREATE OR REPLACE FUNCTION check_message_sender()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM chat c
        JOIN application a ON a.application_id = c.application_id
        JOIN vacancy v ON v.vacancy_id = a.vacancy_id
        WHERE c.chat_id = NEW.chat_id
          AND NEW.sender_id IN (v.employer_id, a.seeker_id)
    ) THEN
        RAISE EXCEPTION 'user % is not a participant of chat %',
            NEW.sender_id, NEW.chat_id
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TABLE application_note (
    note_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id BIGINT NOT NULL
        REFERENCES application (application_id) ON DELETE CASCADE,
    author_id BIGINT
        REFERENCES app_user (user_id) ON DELETE SET NULL,
    body TEXT NOT NULL CHECK (length(btrim(body)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE FUNCTION resume_lock_related_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    old_resume_id BIGINT;
    new_resume_id BIGINT;
BEGIN
    IF TG_OP <> 'INSERT' THEN
        old_resume_id := OLD.resume_id;
    END IF;
    IF TG_OP <> 'DELETE' THEN
        new_resume_id := NEW.resume_id;
    END IF;
    PERFORM 1
    FROM resume r
    WHERE r.resume_id = old_resume_id OR r.resume_id = new_resume_id
    ORDER BY r.resume_id
    FOR UPDATE;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER resume_experience_lock_trigger
BEFORE INSERT OR UPDATE OR DELETE ON resume_experience
FOR EACH ROW EXECUTE FUNCTION resume_lock_related_change();

CREATE TRIGGER resume_education_lock_trigger
BEFORE INSERT OR UPDATE OR DELETE ON resume_education
FOR EACH ROW EXECUTE FUNCTION resume_lock_related_change();

CREATE TRIGGER resume_skill_lock_trigger
BEFORE INSERT OR UPDATE OR DELETE ON resume_skill
FOR EACH ROW EXECUTE FUNCTION resume_lock_related_change();

CREATE TRIGGER resume_project_lock_trigger
BEFORE INSERT OR UPDATE OR DELETE ON resume_project
FOR EACH ROW EXECUTE FUNCTION resume_lock_related_change();

CREATE TRIGGER resume_category_lock_trigger
BEFORE INSERT OR UPDATE OR DELETE ON resume_category
FOR EACH ROW EXECUTE FUNCTION resume_lock_related_change();

CREATE FUNCTION create_application(
    p_vacancy_id BIGINT,
    p_vacancy_version INTEGER,
    p_resume_id BIGINT,
    p_seeker_id BIGINT,
    p_cover_letter TEXT
)
RETURNS TABLE (new_application_id BIGINT, created_on TIMESTAMPTZ, updated_on TIMESTAMPTZ)
LANGUAGE plpgsql
AS $$
DECLARE
    created_application_id BIGINT;
    created_application_at TIMESTAMPTZ;
    updated_application_at TIMESTAMPTZ;
BEGIN
    PERFORM 1
    FROM resume r
    WHERE r.resume_id = p_resume_id
      AND r.seeker_id = p_seeker_id
      AND r.deleted_at IS NULL
    FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Resume does not exist or belongs to another seeker'
            USING ERRCODE = '23503';
    END IF;

    PERFORM 1
    FROM vacancy v
    WHERE v.vacancy_id = p_vacancy_id
    FOR SHARE;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Vacancy does not exist'
            USING ERRCODE = '23503';
    END IF;

    PERFORM 1
    FROM vacancy_history h
    WHERE h.vacancy_id = p_vacancy_id
      AND h.version = p_vacancy_version
      AND h.version = (
          SELECT MAX(h2.version)
          FROM vacancy_history h2
          WHERE h2.vacancy_id = p_vacancy_id
      );

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Vacancy version is not current'
            USING ERRCODE = '23514';
    END IF;

    INSERT INTO application (vacancy_id, vacancy_version, seeker_id, cover_letter)
    VALUES (p_vacancy_id, p_vacancy_version, p_seeker_id, p_cover_letter)
    RETURNING application_id, created_at, updated_at
    INTO created_application_id, created_application_at, updated_application_at;

    INSERT INTO application_resume_snapshot (
        application_id, source_resume_id, title, description,
        legacy_experience_text, legacy_education_text, target_position,
        city, employment_type, work_format, work_schedule,
        salary_from, portfolio_url, languages
    )
    SELECT created_application_id, r.resume_id, r.title, r.description,
           l.experience_text, l.education_text, r.target_position,
           r.city, r.employment_type, r.work_format, r.work_schedule,
           r.salary_from, r.portfolio_url, r.languages
    FROM resume r
    LEFT JOIN resume_legacy_content l ON l.resume_id = r.resume_id
    WHERE r.resume_id = p_resume_id;

    INSERT INTO application_snapshot_experience (
        application_id, position, company_name, position_title,
        started_on, ended_on, is_current, achievements
    )
    SELECT created_application_id, position, company_name, position_title,
           started_on, ended_on, is_current, achievements
    FROM resume_experience
    WHERE resume_id = p_resume_id;

    INSERT INTO application_snapshot_education (
        application_id, position, record_type, institution, completion_year, specialization
    )
    SELECT created_application_id, position, record_type, institution,
           completion_year, specialization
    FROM resume_education
    WHERE resume_id = p_resume_id;

    INSERT INTO application_snapshot_skill (application_id, position, name)
    SELECT created_application_id, position, name
    FROM resume_skill
    WHERE resume_id = p_resume_id;

    INSERT INTO application_snapshot_project (application_id, position, title, url, description)
    SELECT created_application_id, position, title, url, description
    FROM resume_project
    WHERE resume_id = p_resume_id;

    INSERT INTO application_snapshot_category (application_id, category_name)
    SELECT created_application_id, c.name
    FROM resume_category rc
    JOIN category c ON c.category_id = rc.category_id
    WHERE rc.resume_id = p_resume_id;

    UPDATE application_resume_snapshot
       SET is_sealed = TRUE
     WHERE application_id = created_application_id;

    INSERT INTO application_status_history (application_id, status)
    VALUES (created_application_id, 'created');

    RETURN QUERY SELECT created_application_id, created_application_at, updated_application_at;
END;
$$;

CREATE FUNCTION application_require_snapshot()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
          FROM application_resume_snapshot
         WHERE application_id = NEW.application_id
           AND is_sealed
    ) THEN
        RAISE EXCEPTION 'Application % must have a sealed resume snapshot', NEW.application_id
            USING ERRCODE = '23514';
    END IF;
    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER application_snapshot_required_trigger
AFTER INSERT ON application
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION application_require_snapshot();

CREATE FUNCTION application_guard_resume_snapshot()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    candidate application_resume_snapshot%ROWTYPE;
BEGIN
    IF TG_OP = 'DELETE' THEN
        IF NOT EXISTS (
            SELECT 1 FROM application WHERE application_id = OLD.application_id
        ) THEN
            RETURN OLD;
        END IF;
    END IF;

    IF TG_OP <> 'UPDATE' THEN
        RAISE EXCEPTION 'Application resume snapshots cannot be removed'
            USING ERRCODE = '23514';
    END IF;

    candidate := NEW;
    IF NEW.is_sealed AND NOT OLD.is_sealed THEN
        candidate.is_sealed := OLD.is_sealed;
    END IF;

    IF NEW.source_resume_id IS NULL
       AND OLD.source_resume_id IS NOT NULL
       AND NOT EXISTS (
           SELECT 1 FROM resume WHERE resume_id = OLD.source_resume_id
       ) THEN
        candidate.source_resume_id := OLD.source_resume_id;
    END IF;

    IF candidate IS DISTINCT FROM OLD THEN
        RAISE EXCEPTION 'Application resume snapshots cannot be changed'
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION application_guard_snapshot_item()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM 1
           FROM application_resume_snapshot
          WHERE application_id = NEW.application_id
            AND NOT is_sealed
            FOR SHARE;
        IF FOUND THEN
            RETURN NEW;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF NOT EXISTS (
            SELECT 1 FROM application_resume_snapshot WHERE application_id = OLD.application_id
        ) THEN
            RETURN OLD;
        END IF;
    END IF;

    RAISE EXCEPTION 'Application resume snapshot items can only be inserted before sealing'
        USING ERRCODE = '23514';
END;
$$;

CREATE TRIGGER application_resume_snapshot_guard
BEFORE UPDATE OR DELETE ON application_resume_snapshot
FOR EACH ROW EXECUTE FUNCTION application_guard_resume_snapshot();

CREATE TRIGGER application_resume_snapshot_truncate_guard
BEFORE TRUNCATE ON application_resume_snapshot
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_resume_snapshot();

CREATE TRIGGER application_snapshot_item_guard
BEFORE INSERT OR UPDATE OR DELETE ON application_snapshot_experience
FOR EACH ROW EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_truncate_guard
BEFORE TRUNCATE ON application_snapshot_experience
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_guard
BEFORE INSERT OR UPDATE OR DELETE ON application_snapshot_education
FOR EACH ROW EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_truncate_guard
BEFORE TRUNCATE ON application_snapshot_education
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_guard
BEFORE INSERT OR UPDATE OR DELETE ON application_snapshot_skill
FOR EACH ROW EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_truncate_guard
BEFORE TRUNCATE ON application_snapshot_skill
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_guard
BEFORE INSERT OR UPDATE OR DELETE ON application_snapshot_project
FOR EACH ROW EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_truncate_guard
BEFORE TRUNCATE ON application_snapshot_project
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_guard
BEFORE INSERT OR UPDATE OR DELETE ON application_snapshot_category
FOR EACH ROW EXECUTE FUNCTION application_guard_snapshot_item();

CREATE TRIGGER application_snapshot_item_truncate_guard
BEFORE TRUNCATE ON application_snapshot_category
FOR EACH STATEMENT EXECUTE FUNCTION application_guard_snapshot_item();

CREATE FUNCTION application_prevent_delete()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM vacancy WHERE vacancy_id = OLD.vacancy_id) THEN
        RAISE EXCEPTION 'Applications can only be deleted with their vacancy'
            USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
END;
$$;

CREATE TRIGGER application_delete_prohibited_trigger
BEFORE DELETE ON application
FOR EACH ROW EXECUTE FUNCTION application_prevent_delete();

CREATE FUNCTION application_prevent_identity_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.vacancy_id IS DISTINCT FROM OLD.vacancy_id
       OR NEW.vacancy_version IS DISTINCT FROM OLD.vacancy_version
       OR NEW.seeker_id IS DISTINCT FROM OLD.seeker_id THEN
        RAISE EXCEPTION 'Application identity cannot be changed'
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER application_identity_immutable_trigger
BEFORE UPDATE OF vacancy_id, vacancy_version, seeker_id ON application
FOR EACH ROW EXECUTE FUNCTION application_prevent_identity_change();

COMMIT;
