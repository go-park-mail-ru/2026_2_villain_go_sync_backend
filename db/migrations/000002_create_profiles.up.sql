CREATE TABLE employer_profile
(
    user_id      BIGINT PRIMARY KEY,
    role         TEXT        NOT NULL DEFAULT 'employer'
        CHECK (role = 'employer'),
    company_name TEXT        NOT NULL,
    description  TEXT,
    website      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT employer_profile_user_fk
        FOREIGN KEY (user_id, role)
            REFERENCES app_user (user_id, role)
            ON DELETE CASCADE
);

CREATE TABLE seeker_profile
(
    user_id    BIGINT PRIMARY KEY,
    role       TEXT        NOT NULL DEFAULT 'seeker'
        CHECK (role = 'seeker'),
    first_name TEXT,
    last_name  TEXT,
    phone      TEXT,
    about      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT seeker_profile_user_fk
        FOREIGN KEY (user_id, role)
            REFERENCES app_user (user_id, role)
            ON DELETE CASCADE
);
