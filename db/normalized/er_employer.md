# HireNoon — Employer + Favorites

```mermaid
erDiagram
    APP_USER ||--o| EMPLOYER_PROFILE : profile
    EMPLOYER_PROFILE ||--o{ EMPLOYER_TECHNOLOGY : technology
    EMPLOYER_PROFILE ||--o{ EMPLOYER_CULTURE : culture
    EMPLOYER_PROFILE ||--o{ VACANCY : vacancy
    VACANCY ||--o{ VACANCY_HISTORY : version
    VACANCY ||--o{ VACANCY_CATEGORY : tag
    CATEGORY ||--o{ VACANCY_CATEGORY : tag
    VACANCY_HISTORY ||--o{ VACANCY_VERSION_SKILL : skill
    APP_USER ||--o{ FAVORITE_VACANCY : favorite
    VACANCY ||--o{ FAVORITE_VACANCY : favorite
    APP_USER ||--o{ VACANCY_NOTE : note
    VACANCY ||--o{ VACANCY_NOTE : note
    APP_USER {
        attr user_id PK
        attr email UK
        attr password_hash
        attr role
        attr created_at
        attr updated_at
    }
    EMPLOYER_PROFILE {
        attr user_id PK,FK
        attr company_name
        attr description
        attr website
        attr logo_key
        attr created_at
        attr updated_at
        attr industry
        attr team_size
        attr main_city
        attr benefits
        attr hiring_email
        attr office_address
    }
    EMPLOYER_TECHNOLOGY {
        attr technology_id PK
        attr employer_id FK
        attr name
        attr position
        attr created_at
        attr updated_at
    }
    EMPLOYER_CULTURE {
        attr culture_id PK
        attr employer_id FK
        attr title
        attr short_description
        attr description
        attr position
        attr created_at
        attr updated_at
    }
    CATEGORY {
        attr category_id PK
        attr name UK
        attr created_at
        attr updated_at
    }
    VACANCY {
        attr vacancy_id PK
        attr employer_id FK
        attr created_at
        attr updated_at
        attr apply_until
    }
    VACANCY_HISTORY {
        attr vacancy_id PK,FK
        attr version PK
        attr title
        attr description
        attr salary_from
        attr salary_to
        attr changed_at
        attr currency
        attr responsibilities
        attr requirements
        attr offer
        attr experience_level
        attr employment_type
        attr work_format
        attr work_schedule
        attr work_hours
        attr city
        attr office_address
    }
    VACANCY_CATEGORY {
        attr vacancy_id PK,FK
        attr category_id PK,FK
    }
    VACANCY_VERSION_SKILL {
        attr skill_id PK
        attr vacancy_id FK
        attr version FK
        attr name
        attr position
        attr created_at
        attr updated_at
    }
    FAVORITE_VACANCY {
        attr user_id PK,FK
        attr vacancy_id PK,FK
        attr created_at
        attr updated_at
    }
    VACANCY_NOTE {
        attr user_id PK,FK
        attr vacancy_id PK,FK
        attr body
        attr created_at
        attr updated_at
    }
```

Связи показаны для переходной единой БД. При физическом разделении межсервисные ссылки превращаются в ссылки по ID и gRPC-вызовы; MinIO/S3 хранит файлы по ключам из профилей и `pdf_document`.
