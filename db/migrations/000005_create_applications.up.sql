CREATE TABLE application
(
    application_id  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    vacancy_id      BIGINT      NOT NULL,
    vacancy_version INT         NOT NULL,
    resume_id       BIGINT      NOT NULL,
    cover_letter    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT application_vacancy_version_fk
        FOREIGN KEY (vacancy_id, vacancy_version)
            REFERENCES vacancy_history (vacancy_id, version),

    CONSTRAINT application_resume_fk
        FOREIGN KEY (resume_id)
            REFERENCES resume (resume_id)
            ON DELETE CASCADE,

    CONSTRAINT application_vacancy_resume_unique
        UNIQUE (vacancy_id, resume_id)
);

CREATE TABLE application_status_history
(
    status_history_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id    BIGINT      NOT NULL,
    status            VARCHAR(50) NOT NULL,
    comment           TEXT,
    changed_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT application_status_application_fk
        FOREIGN KEY (application_id)
            REFERENCES application (application_id)
            ON DELETE CASCADE
);