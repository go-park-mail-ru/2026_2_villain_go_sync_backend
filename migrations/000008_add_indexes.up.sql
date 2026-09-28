CREATE INDEX idx_vacancy_employer_id
    ON vacancy (employer_id);

CREATE INDEX idx_resume_seeker_id
    ON resume (seeker_id);

CREATE INDEX idx_vacancy_category_category_id
    ON vacancy_category (category_id);

CREATE INDEX idx_resume_category_category_id
    ON resume_category (category_id);

CREATE INDEX idx_application_resume_id
    ON application (resume_id);

CREATE INDEX idx_application_vacancy_version
    ON application (vacancy_id, vacancy_version);

CREATE INDEX idx_application_status_history_application_changed
    ON application_status_history (application_id, changed_at);

CREATE INDEX idx_notification_user_created
    ON notification (user_id, created_at);

CREATE INDEX idx_message_chat_created
    ON message (chat_id, created_at);