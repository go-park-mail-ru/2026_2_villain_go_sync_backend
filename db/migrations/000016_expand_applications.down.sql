BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM application) THEN
        RAISE EXCEPTION 'Cannot roll back 000016 while applications exist';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM application_status_history
        WHERE status IN ('reviewing', 'offer') OR reason_code IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Cannot roll back 000016 while new status data exists';
    END IF;
    IF EXISTS (SELECT 1 FROM application_note) THEN
        RAISE EXCEPTION 'Cannot roll back 000016 while notes exist';
    END IF;
END;
$$;

DROP TRIGGER application_delete_prohibited_trigger ON application;
DROP FUNCTION application_prevent_delete();

DROP TRIGGER application_identity_immutable_trigger ON application;
DROP FUNCTION application_prevent_identity_change();
DROP TRIGGER application_snapshot_required_trigger ON application;
DROP FUNCTION application_require_snapshot();
DROP FUNCTION create_application(BIGINT, INTEGER, BIGINT, BIGINT, TEXT);

DROP TRIGGER resume_category_lock_trigger ON resume_category;
DROP TRIGGER resume_project_lock_trigger ON resume_project;
DROP TRIGGER resume_skill_lock_trigger ON resume_skill;
DROP TRIGGER resume_education_lock_trigger ON resume_education;
DROP TRIGGER resume_experience_lock_trigger ON resume_experience;
DROP FUNCTION resume_lock_related_change();

DROP TABLE application_note;
DROP TABLE application_snapshot_category;
DROP TABLE application_snapshot_project;
DROP TABLE application_snapshot_skill;
DROP TABLE application_snapshot_education;
DROP TABLE application_snapshot_experience;
DROP TABLE application_resume_snapshot;

DROP FUNCTION application_guard_snapshot_item();
DROP FUNCTION application_guard_resume_snapshot();

ALTER TABLE application
    DROP CONSTRAINT application_vacancy_seeker_unique,
    DROP CONSTRAINT application_seeker_fk,
    DROP COLUMN seeker_id,
    ADD COLUMN resume_id BIGINT NOT NULL,
    ADD CONSTRAINT application_resume_fk
        FOREIGN KEY (resume_id) REFERENCES resume (resume_id) ON DELETE CASCADE,
    ADD CONSTRAINT application_vacancy_resume_unique
        UNIQUE (vacancy_id, resume_id);

ALTER TABLE application
    DROP CONSTRAINT application_vacancy_version_fk;

ALTER TABLE application
    ADD CONSTRAINT application_vacancy_version_fk
        FOREIGN KEY (vacancy_id, vacancy_version)
        REFERENCES vacancy_history (vacancy_id, version) ON DELETE CASCADE;

CREATE INDEX idx_application_resume_id ON application (resume_id);

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
        JOIN resume r ON r.resume_id = a.resume_id
        WHERE c.chat_id = NEW.chat_id
          AND NEW.sender_id IN (v.employer_id, r.seeker_id)
    ) THEN
        RAISE EXCEPTION 'user % is not a participant of chat %',
            NEW.sender_id, NEW.chat_id
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

ALTER TABLE application_status_history
    DROP CONSTRAINT application_rejection_required_check,
    DROP CONSTRAINT application_rejection_status_check,
    DROP CONSTRAINT application_rejection_code_check,
    DROP CONSTRAINT application_status_history_status_check;

ALTER TABLE application_status_history
    ADD CONSTRAINT application_status_history_status_check
        CHECK (status IN ('created', 'viewed', 'interview', 'accepted', 'rejected'));

ALTER TABLE application_status_history
    DROP COLUMN changed_by,
    DROP COLUMN reason_code;

COMMIT;
