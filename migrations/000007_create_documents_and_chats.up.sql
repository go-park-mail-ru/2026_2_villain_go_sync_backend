CREATE TABLE pdf_document
(
    document_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resume_id   BIGINT       NOT NULL,
    file_path   VARCHAR(500) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

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

    CONSTRAINT message_chat_fk
        FOREIGN KEY (chat_id)
            REFERENCES chat (chat_id)
            ON DELETE CASCADE,

    CONSTRAINT message_sender_fk
        FOREIGN KEY (sender_id)
            REFERENCES users (user_id)
);