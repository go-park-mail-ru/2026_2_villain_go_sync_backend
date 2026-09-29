CREATE TABLE category
(
    category_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vacancy
(
    vacancy_id  BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    employer_id BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vacancy_employer_fk
        FOREIGN KEY (employer_id)
            REFERENCES employer_profile (user_id)
            ON DELETE CASCADE
);

CREATE TABLE vacancy_history
(
    vacancy_id  BIGINT      NOT NULL,
    version     INT         NOT NULL
        CHECK (version > 0),
    title       TEXT        NOT NULL,
    description TEXT,
    salary_from NUMERIC(12, 2),
    salary_to   NUMERIC(12, 2),
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (vacancy_id, version),

    CONSTRAINT vacancy_history_vacancy_fk
        FOREIGN KEY (vacancy_id)
            REFERENCES vacancy (vacancy_id)
            ON DELETE CASCADE,

    CONSTRAINT vacancy_salary_from_nonnegative
        CHECK (salary_from IS NULL OR salary_from >= 0),

    CONSTRAINT vacancy_salary_to_nonnegative
        CHECK (salary_to IS NULL OR salary_to >= 0),

    CONSTRAINT vacancy_salary_range
        CHECK (
            salary_from IS NULL
                OR salary_to IS NULL
                OR salary_from <= salary_to
            )
);

CREATE TABLE vacancy_category
(
    vacancy_id  BIGINT NOT NULL,
    category_id BIGINT NOT NULL,

    PRIMARY KEY (vacancy_id, category_id),

    CONSTRAINT vacancy_category_vacancy_fk
        FOREIGN KEY (vacancy_id)
            REFERENCES vacancy (vacancy_id)
            ON DELETE CASCADE,

    CONSTRAINT vacancy_category_category_fk
        FOREIGN KEY (category_id)
            REFERENCES category (category_id)
            ON DELETE CASCADE
);
