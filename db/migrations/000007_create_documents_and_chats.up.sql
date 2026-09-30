CREATE TABLE pdf_document
(
    document_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id   BIGINT      NOT NULL,
    file_path   TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pdf_document_resume_fk
        FOREIGN KEY (resume_id)
            REFERENCES resume (resume_id)
            ON DELETE CASCADE
);

CREATE TABLE chat
(
    chat_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id BIGINT      NOT NULL UNIQUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chat_application_fk
        FOREIGN KEY (application_id)
            REFERENCES application (application_id)
            ON DELETE CASCADE
);

CREATE TABLE message
(
    message_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    chat_id    BIGINT      NOT NULL,
    sender_id  BIGINT      NOT NULL,
    body       TEXT        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT message_chat_fk
        FOREIGN KEY (chat_id)
            REFERENCES chat (chat_id)
            ON DELETE CASCADE,

    CONSTRAINT message_sender_fk
        FOREIGN KEY (sender_id)
            REFERENCES app_user (user_id)
);

CREATE FUNCTION check_message_sender()
    RETURNS TRIGGER AS $$
BEGIN
    IF
NOT EXISTS (
        SELECT 1
        FROM chat c
        JOIN application a
            ON a.application_id = c.application_id
        JOIN vacancy v
            ON v.vacancy_id = a.vacancy_id
        JOIN resume r
            ON r.resume_id = a.resume_id
        WHERE c.chat_id = NEW.chat_id
          AND (
              NEW.sender_id = v.employer_id
              OR NEW.sender_id = r.seeker_id
          )
    ) THEN
        RAISE EXCEPTION
            'user % is not a participant of chat %',
            NEW.sender_id,
            NEW.chat_id;
END IF;

RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER message_sender_check
    BEFORE INSERT OR
UPDATE OF chat_id, sender_id
ON message
    FOR EACH ROW
    EXECUTE FUNCTION check_message_sender();
