CREATE TABLE employer_profile
(
    user_id      BIGINT PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    description  TEXT,
    website      VARCHAR(255),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT employer_profile_user_fk
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
            ON DELETE CASCADE
);

CREATE TABLE seeker_profile
(
    user_id    BIGINT PRIMARY KEY,
    first_name VARCHAR(255),
    last_name  VARCHAR(255),
    phone      VARCHAR(50),
    about      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT seeker_profile_user_fk
        FOREIGN KEY (user_id)
            REFERENCES users (user_id)
            ON DELETE CASCADE
);