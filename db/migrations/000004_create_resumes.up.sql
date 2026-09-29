CREATE TABLE resume
(
    resume_id   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    seeker_id   BIGINT       NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    experience  TEXT,
    education   TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT resume_seeker_fk
        FOREIGN KEY (seeker_id)
            REFERENCES seeker_profile (user_id)
            ON DELETE CASCADE
);

CREATE TABLE resume_category
(
    resume_id   BIGINT NOT NULL,
    category_id BIGINT NOT NULL,

    PRIMARY KEY (resume_id, category_id),

    CONSTRAINT resume_category_resume_fk
        FOREIGN KEY (resume_id)
            REFERENCES resume (resume_id)
            ON DELETE CASCADE,

    CONSTRAINT resume_category_category_fk
        FOREIGN KEY (category_id)
            REFERENCES category (category_id)
            ON DELETE CASCADE
);