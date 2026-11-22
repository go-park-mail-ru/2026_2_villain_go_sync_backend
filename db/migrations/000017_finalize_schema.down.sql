BEGIN;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM vacancy_note) THEN
        RAISE EXCEPTION 'Cannot roll back 000017 while vacancy notes exist';
    END IF;
END;
$$;

DROP TABLE vacancy_note;

ALTER TABLE favorite_vacancy
    DROP COLUMN updated_at;

ALTER TABLE pdf_document
    DROP COLUMN updated_at;

ALTER TABLE vacancy
    DROP COLUMN updated_at;

CREATE INDEX idx_vacancy_employer_id ON vacancy (employer_id);
CREATE INDEX idx_resume_seeker_id ON resume (seeker_id);
CREATE INDEX idx_vacancy_category_category_id ON vacancy_category (category_id);
CREATE INDEX idx_resume_category_category_id ON resume_category (category_id);
CREATE INDEX idx_application_vacancy_version ON application (vacancy_id, vacancy_version);
CREATE INDEX idx_application_status_history_application_changed
    ON application_status_history (application_id, changed_at);
CREATE INDEX idx_notification_user_created ON notification (user_id, created_at);
CREATE INDEX idx_message_chat_created ON message (chat_id, created_at);
CREATE INDEX idx_favorite_vacancy_vacancy_id ON favorite_vacancy (vacancy_id);
CREATE INDEX idx_notification_topic_id ON notification (topic_id);
CREATE INDEX idx_pdf_document_resume_id ON pdf_document (resume_id);
CREATE INDEX idx_message_sender_id ON message (sender_id);

COMMIT;
