BEGIN;

CREATE TABLE vacancy_note (
    user_id BIGINT NOT NULL
        REFERENCES app_user (user_id) ON DELETE CASCADE,
    vacancy_id BIGINT NOT NULL
        REFERENCES vacancy (vacancy_id) ON DELETE CASCADE,
    body TEXT NOT NULL CHECK (length(btrim(body)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, vacancy_id)
);

ALTER TABLE vacancy
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE pdf_document
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE favorite_vacancy
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DROP INDEX IF EXISTS idx_vacancy_employer_id;
DROP INDEX IF EXISTS idx_resume_seeker_id;
DROP INDEX IF EXISTS idx_vacancy_category_category_id;
DROP INDEX IF EXISTS idx_resume_category_category_id;
DROP INDEX IF EXISTS idx_application_vacancy_version;
DROP INDEX IF EXISTS idx_application_status_history_application_changed;
DROP INDEX IF EXISTS idx_notification_user_created;
DROP INDEX IF EXISTS idx_message_chat_created;
DROP INDEX IF EXISTS idx_favorite_vacancy_vacancy_id;
DROP INDEX IF EXISTS idx_notification_topic_id;
DROP INDEX IF EXISTS idx_pdf_document_resume_id;
DROP INDEX IF EXISTS idx_message_sender_id;

COMMIT;
