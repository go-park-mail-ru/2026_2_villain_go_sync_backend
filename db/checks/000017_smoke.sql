BEGIN;

DO $$
DECLARE
    v_seeker BIGINT;
    v_employer BIGINT;
    v_resume BIGINT;
    v_other_resume BIGINT;
    v_vacancy BIGINT;
    v_application BIGINT;
    v_chat BIGINT;
    v_title TEXT;
    v_statement TEXT;
    v_table TEXT;
    v_blocked BOOLEAN;
BEGIN
    INSERT INTO app_user (email, password_hash, role)
    VALUES ('smoke-seeker-' || txid_current() || '@example.invalid', 'test-hash', 'seeker')
    RETURNING user_id INTO v_seeker;

    INSERT INTO app_user (email, password_hash, role)
    VALUES ('smoke-employer-' || txid_current() || '@example.invalid', 'test-hash', 'employer')
    RETURNING user_id INTO v_employer;

    INSERT INTO seeker_profile (user_id, first_name, last_name)
    VALUES (v_seeker, 'Test', 'Candidate');

    INSERT INTO employer_profile (user_id, company_name)
    VALUES (v_employer, 'Test Company');

    INSERT INTO resume (seeker_id, title, target_position)
    VALUES (v_seeker, 'Go Developer', 'Go Developer')
    RETURNING resume_id INTO v_resume;

    INSERT INTO resume_experience (resume_id, company_name, position_title, position)
    VALUES (v_resume, 'Test Company', 'Developer', 1);

    INSERT INTO resume_education (resume_id, institution, position)
    VALUES (v_resume, 'Test University', 1);

    INSERT INTO resume_skill (resume_id, name, position)
    VALUES (v_resume, 'Go', 1);

    INSERT INTO resume_project (resume_id, title, position)
    VALUES (v_resume, 'Test Project', 1);

    INSERT INTO category (name) VALUES ('Smoke test category');

    INSERT INTO resume_category (resume_id, category_id)
    SELECT v_resume, category_id
      FROM category
     WHERE name = 'Smoke test category';

    INSERT INTO vacancy (employer_id, apply_until)
    VALUES (v_employer, CURRENT_DATE - 1)
    RETURNING vacancy_id INTO v_vacancy;

    FOREACH v_statement IN ARRAY ARRAY[
        'INSERT INTO vacancy_history (vacancy_id, version, title) VALUES ($1, 1, ''Incomplete vacancy'')',
        'INSERT INTO vacancy_history (vacancy_id, version, title, requirements) VALUES ($1, 1, ''Incomplete vacancy'', ''Know Go'')',
        'INSERT INTO vacancy_history (vacancy_id, version, title, responsibilities, requirements)
         VALUES ($1, 1, ''Incomplete vacancy'', ''Develop services'', chr(9) || chr(160))'
    ] LOOP
        v_blocked := FALSE;
        BEGIN
            EXECUTE v_statement USING v_vacancy;
        EXCEPTION WHEN check_violation THEN
            v_blocked := TRUE;
        END;
        IF NOT v_blocked THEN
            RAISE EXCEPTION 'Vacancy version without mandatory fields was accepted';
        END IF;
    END LOOP;

    INSERT INTO vacancy_history (
        vacancy_id, version, title, responsibilities, requirements
    )
    VALUES (
        v_vacancy, 1, 'Go Developer', 'Develop services', 'Know Go'
    );

    SELECT new_application_id INTO v_application
    FROM create_application(v_vacancy, 1, v_resume, v_seeker, NULL);

    IF NOT EXISTS (
        SELECT 1
        FROM application_resume_snapshot
        WHERE application_id = v_application
    ) OR NOT EXISTS (
        SELECT 1
        FROM application_snapshot_experience
        WHERE application_id = v_application
    ) THEN
        RAISE EXCEPTION 'Application snapshot is incomplete';
    END IF;

    INSERT INTO resume (seeker_id, title, target_position)
    VALUES (v_seeker, 'Second CV', 'Go Developer')
    RETURNING resume_id INTO v_other_resume;

    v_blocked := FALSE;
    BEGIN
        PERFORM new_application_id
        FROM create_application(v_vacancy, 1, v_other_resume, v_seeker, NULL);
    EXCEPTION WHEN unique_violation THEN
        v_blocked := TRUE;
    END;
    IF NOT v_blocked THEN
        RAISE EXCEPTION 'Duplicate application was accepted';
    END IF;

    v_blocked := FALSE;
    BEGIN
        INSERT INTO application_status_history (application_id, status)
        VALUES (v_application, 'rejected');
    EXCEPTION WHEN check_violation THEN
        v_blocked := TRUE;
    END;
    IF NOT v_blocked THEN
        RAISE EXCEPTION 'Rejection without reason was accepted';
    END IF;

    INSERT INTO application_status_history (
        application_id, status, reason_code, comment, changed_by
    ) VALUES (
        v_application, 'rejected', 'skills_mismatch', 'Need more Go experience', v_employer
    );

    UPDATE resume SET title = 'Changed Resume' WHERE resume_id = v_resume;

    SELECT title INTO v_title
    FROM application_resume_snapshot
    WHERE application_id = v_application;

    IF v_title <> 'Go Developer' THEN
        RAISE EXCEPTION 'Snapshot changed together with source resume';
    END IF;

    INSERT INTO favorite_vacancy (user_id, vacancy_id)
    VALUES (v_seeker, v_vacancy);

    INSERT INTO vacancy_note (user_id, vacancy_id, body)
    VALUES (v_seeker, v_vacancy, 'Contact later');

    DELETE FROM favorite_vacancy
    WHERE user_id = v_seeker AND vacancy_id = v_vacancy;

    IF NOT EXISTS (
        SELECT 1 FROM vacancy_note
        WHERE user_id = v_seeker AND vacancy_id = v_vacancy
    ) THEN
        RAISE EXCEPTION 'Note disappeared after removing favorite';
    END IF;

    v_blocked := FALSE;
    BEGIN
        UPDATE application_resume_snapshot
           SET source_resume_id = NULL
         WHERE application_id = v_application;
    EXCEPTION WHEN check_violation THEN
        v_blocked := TRUE;
    END;
    IF NOT v_blocked THEN
        RAISE EXCEPTION 'Snapshot source could be detached while the resume exists';
    END IF;

    DELETE FROM resume WHERE resume_id = v_resume;

    IF NOT EXISTS (
        SELECT 1 FROM application_resume_snapshot
        WHERE application_id = v_application AND source_resume_id IS NULL
    ) THEN
        RAISE EXCEPTION 'Snapshot disappeared after source resume deletion';
    END IF;

    SET CONSTRAINTS application_snapshot_required_trigger IMMEDIATE;

    FOREACH v_statement IN ARRAY ARRAY[
        'UPDATE application_resume_snapshot SET title = ''Changed snapshot'' WHERE application_id = $1',
        'UPDATE application_resume_snapshot SET is_sealed = FALSE WHERE application_id = $1',
        'DELETE FROM application_resume_snapshot WHERE application_id = $1',
        'INSERT INTO application_snapshot_skill (application_id, position, name) VALUES ($1, 2, ''Unexpected skill'')',
        'TRUNCATE application_snapshot_skill'
    ] LOOP
        v_blocked := FALSE;
        BEGIN
            EXECUTE v_statement USING v_application;
        EXCEPTION WHEN check_violation THEN
            v_blocked := TRUE;
        END;
        IF NOT v_blocked THEN
            RAISE EXCEPTION 'Snapshot mutation was accepted: %', v_statement;
        END IF;
    END LOOP;

    FOREACH v_table IN ARRAY ARRAY[
        'application_snapshot_experience',
        'application_snapshot_education',
        'application_snapshot_skill',
        'application_snapshot_project',
        'application_snapshot_category'
    ] LOOP
        v_blocked := FALSE;
        BEGIN
            EXECUTE format('DELETE FROM %I WHERE application_id = $1', v_table)
            USING v_application;
        EXCEPTION WHEN check_violation THEN
            v_blocked := TRUE;
        END;
        IF NOT v_blocked THEN
            RAISE EXCEPTION 'Snapshot items could be deleted from %', v_table;
        END IF;
    END LOOP;

    v_blocked := FALSE;
    BEGIN
        UPDATE application_snapshot_skill
           SET name = 'Changed skill'
         WHERE application_id = v_application AND position = 1;
    EXCEPTION WHEN check_violation THEN
        v_blocked := TRUE;
    END;
    IF NOT v_blocked THEN
        RAISE EXCEPTION 'Snapshot skill could be changed';
    END IF;

    v_blocked := FALSE;
    BEGIN
        DELETE FROM application WHERE application_id = v_application;
    EXCEPTION WHEN check_violation THEN
        v_blocked := TRUE;
    END;
    IF NOT v_blocked THEN
        RAISE EXCEPTION 'Historical application was deleted';
    END IF;

    INSERT INTO favorite_vacancy (user_id, vacancy_id)
    VALUES (v_seeker, v_vacancy);

    INSERT INTO vacancy_category (vacancy_id, category_id)
    SELECT v_vacancy, category_id
      FROM category
     WHERE name = 'Smoke test category';

    INSERT INTO vacancy_version_skill (vacancy_id, version, name, position)
    VALUES (v_vacancy, 1, 'Go', 1);

    INSERT INTO application_note (application_id, author_id, body)
    VALUES (v_application, v_employer, 'Interview notes');

    INSERT INTO chat (application_id)
    VALUES (v_application)
    RETURNING chat_id INTO v_chat;

    INSERT INTO message (chat_id, sender_id, body)
    VALUES (v_chat, v_seeker, 'Hello');

    DELETE FROM vacancy WHERE vacancy_id = v_vacancy;

    FOREACH v_table IN ARRAY ARRAY[
        'vacancy', 'vacancy_history', 'vacancy_category', 'vacancy_version_skill',
        'favorite_vacancy', 'vacancy_note', 'application'
    ] LOOP
        EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE vacancy_id = $1)', v_table)
        INTO v_blocked USING v_vacancy;
        IF v_blocked THEN
            RAISE EXCEPTION 'Vacancy data remained in %', v_table;
        END IF;
    END LOOP;

    FOREACH v_table IN ARRAY ARRAY[
        'application_resume_snapshot', 'application_snapshot_experience',
        'application_snapshot_education', 'application_snapshot_skill',
        'application_snapshot_project', 'application_snapshot_category',
        'application_status_history', 'application_note', 'chat'
    ] LOOP
        EXECUTE format('SELECT EXISTS (SELECT 1 FROM %I WHERE application_id = $1)', v_table)
        INTO v_blocked USING v_application;
        IF v_blocked THEN
            RAISE EXCEPTION 'Application data remained in %', v_table;
        END IF;
    END LOOP;

    IF EXISTS (SELECT 1 FROM message WHERE chat_id = v_chat) THEN
        RAISE EXCEPTION 'Chat messages remained after vacancy deletion';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM resume WHERE resume_id = v_other_resume)
       OR NOT EXISTS (SELECT 1 FROM category WHERE name = 'Smoke test category') THEN
        RAISE EXCEPTION 'Vacancy deletion removed independent data';
    END IF;
END;
$$;

ROLLBACK;
