CREATE TABLE favorite_vacancy
(
    user_id    BIGINT      NOT NULL,
    vacancy_id BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, vacancy_id),

    CONSTRAINT favorite_vacancy_user_fk
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
            ON DELETE CASCADE,

    CONSTRAINT favorite_vacancy_vacancy_fk
        FOREIGN KEY (vacancy_id)
            REFERENCES vacancy (vacancy_id)
            ON DELETE CASCADE
);

CREATE TABLE notification_topic
(
    topic_id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code          VARCHAR(100) NOT NULL UNIQUE,
    title         VARCHAR(255) NOT NULL,
    body_template TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE notification
(
    notification_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         BIGINT      NOT NULL,
    topic_id        BIGINT      NOT NULL,
    rendered_text   TEXT        NOT NULL,
    is_read         BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notification_user_fk
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
            ON DELETE CASCADE,

    CONSTRAINT notification_topic_fk
        FOREIGN KEY (topic_id)
            REFERENCES notification_topic (topic_id)
);