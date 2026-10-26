ALTER TABLE application
DROP
CONSTRAINT application_vacancy_version_fk;

ALTER TABLE application
    ADD CONSTRAINT application_vacancy_version_fk
        FOREIGN KEY (vacancy_id, vacancy_version)
            REFERENCES vacancy_history (vacancy_id, version)
            ON DELETE CASCADE;
