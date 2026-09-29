CREATE TABLE employer_profile
(
    user_id      BIGINT PRIMARY KEY,
    company_name TEXT        NOT NULL,
    description  TEXT,
    website      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT employer_profile_user_fk
        FOREIGN KEY (user_id)
            REFERENCES app_user (user_id)
            ON DELETE CASCADE
);

CREATE TABLE seeker_profile
(
    user_id    BIGINT PRIMARY KEY,
    first_name TEXT,
    last_name  TEXT,
    phone      TEXT,
    about      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT seeker_profile_user_fk
        FOREIGN KEY (user_id)
            REFERENCES app_user (user_id)
            ON DELETE CASCADE
);

CREATE FUNCTION check_profile_role()
    RETURNS TRIGGER AS $$
BEGIN
    IF
NOT EXISTS (
        SELECT 1
        FROM app_user
        WHERE user_id = NEW.user_id
          AND role = TG_ARGV[0]
    ) THEN
        RAISE EXCEPTION 'user % must have role %', NEW.user_id, TG_ARGV[0];
END IF;

RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER employer_profile_role_check
    BEFORE INSERT OR
UPDATE ON employer_profile
    FOR EACH ROW
    EXECUTE FUNCTION check_profile_role('employer');

CREATE TRIGGER seeker_profile_role_check
    BEFORE INSERT OR
UPDATE ON seeker_profile
    FOR EACH ROW
    EXECUTE FUNCTION check_profile_role('seeker');

CREATE FUNCTION check_app_user_role_change()
    RETURNS TRIGGER AS $$
BEGIN
    IF
NEW.role = 'employer'
       AND EXISTS (
           SELECT 1
           FROM seeker_profile
           WHERE user_id = NEW.user_id
       )
    THEN
        RAISE EXCEPTION 'user % already has seeker profile', NEW.user_id;
END IF;

    IF
NEW.role = 'seeker'
       AND EXISTS (
           SELECT 1
           FROM employer_profile
           WHERE user_id = NEW.user_id
       )
    THEN
        RAISE EXCEPTION 'user % already has employer profile', NEW.user_id;
END IF;

RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER app_user_role_change_check
    BEFORE UPDATE OF role
    ON app_user
    FOR EACH ROW
    EXECUTE FUNCTION check_app_user_role_change();
