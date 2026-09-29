# ER-диаграмма базы данных

```mermaid
erDiagram
    USERS ||--o| EMPLOYER_PROFILE : "has employer profile"
    USERS ||--o| SEEKER_PROFILE : "has seeker profile"
    USERS ||--o{ NOTIFICATION : receives
    USERS ||--o{ FAVORITE_VACANCY : adds
    USERS ||--o{ MESSAGE : sends

    EMPLOYER_PROFILE ||--o{ VACANCY : creates
    SEEKER_PROFILE ||--o{ RESUME : owns

    VACANCY ||--o{ VACANCY_CATEGORY : "tagged with"
    RESUME ||--o{ RESUME_CATEGORY : "tagged with"

    CATEGORY ||--o{ VACANCY_CATEGORY : "applies to"
    CATEGORY ||--o{ RESUME_CATEGORY : "applies to"

    VACANCY ||--o{ FAVORITE_VACANCY : favorited
    VACANCY ||--o{ VACANCY_HISTORY : "has versions"

    VACANCY_HISTORY ||--o{ APPLICATION : receives

    APPLICATION ||--o{ APPLICATION_STATUS_HISTORY : "has status history"
    APPLICATION ||--o| CHAT : opens

    CHAT ||--o{ MESSAGE : contains

    RESUME ||--o{ APPLICATION : used_in
    RESUME ||--o{ PDF_DOCUMENT : generates

    NOTIFICATION_TOPIC ||--o{ NOTIFICATION : "instantiated as"

    USERS {
        bigint user_id PK
        varchar email UK
        varchar password_hash
        varchar role
        timestamp created_at
        timestamp updated_at
    }

    EMPLOYER_PROFILE {
        bigint user_id PK, FK
        varchar company_name
        text description
        varchar website
        timestamp created_at
        timestamp updated_at
    }

    SEEKER_PROFILE {
        bigint user_id PK, FK
        varchar first_name
        varchar last_name
        varchar phone
        text about
        timestamp created_at
        timestamp updated_at
    }

    CATEGORY {
        bigint category_id PK
        varchar name UK
        timestamp created_at
        timestamp updated_at
    }

    VACANCY {
        bigint vacancy_id PK
        bigint employer_id FK
        timestamp created_at
    }

    VACANCY_CATEGORY {
        bigint vacancy_id PK, FK
        bigint category_id PK, FK
    }

    VACANCY_HISTORY {
        bigint vacancy_id PK, FK
        int version PK
        varchar title
        text description
        numeric salary_from
        numeric salary_to
        timestamp changed_at
    }

    RESUME {
        bigint resume_id PK
        bigint seeker_id FK
        varchar title
        text description
        text experience
        text education
        timestamp created_at
        timestamp updated_at
    }

    RESUME_CATEGORY {
        bigint resume_id PK, FK
        bigint category_id PK, FK
    }

    APPLICATION {
        bigint application_id PK
        bigint vacancy_id FK
        int vacancy_version FK
        bigint resume_id FK
        text cover_letter
        timestamp created_at
        timestamp updated_at
    }

    APPLICATION_STATUS_HISTORY {
        bigint status_history_id PK
        bigint application_id FK
        varchar status
        text comment
        timestamp changed_at
    }

    FAVORITE_VACANCY {
        bigint user_id PK, FK
        bigint vacancy_id PK, FK
        timestamp created_at
    }

    NOTIFICATION_TOPIC {
        bigint topic_id PK
        varchar code UK
        varchar title
        text body_template
        timestamp created_at
        timestamp updated_at
    }

    NOTIFICATION {
        bigint notification_id PK
        bigint user_id FK
        bigint topic_id FK
        text rendered_text
        boolean is_read
        timestamp created_at
    }

    PDF_DOCUMENT {
        bigint document_id PK
        bigint resume_id FK
        varchar file_path
        timestamp created_at
    }

    CHAT {
        bigint chat_id PK
        bigint application_id UK, FK
        timestamp created_at
    }

    MESSAGE {
        bigint message_id PK
        bigint chat_id FK
        bigint sender_id FK
        text body
        boolean is_read
        timestamp created_at
    }
```

> `APPLICATION` хранит конкретную версию вакансии:
> `(vacancy_id, vacancy_version) -> VACANCY_HISTORY(vacancy_id, version)`.

> Для `APPLICATION` также действует ограничение
> `UNIQUE(vacancy_id, resume_id)`.
