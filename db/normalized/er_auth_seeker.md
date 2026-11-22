# HireNoon — Auth + Seeker

```mermaid
erDiagram
    APP_USER ||--o| SEEKER_PROFILE : profile
    SEEKER_PROFILE ||--o{ SEEKER_EXPERIENCE : experience
    SEEKER_PROFILE ||--o{ SEEKER_EDUCATION : education
    SEEKER_PROFILE ||--o{ SEEKER_SKILL : skill
    SEEKER_PROFILE ||--o{ SEEKER_PROJECT : project
    SEEKER_PROFILE ||--o{ RESUME : resume
    RESUME ||--o| RESUME_LEGACY_CONTENT : archive
    RESUME ||--o{ RESUME_EXPERIENCE : experience
    RESUME ||--o{ RESUME_EDUCATION : education
    RESUME ||--o{ RESUME_SKILL : skill
    RESUME ||--o{ RESUME_PROJECT : project
    RESUME ||--o{ RESUME_CATEGORY : tag
    CATEGORY ||--o{ RESUME_CATEGORY : tag
    RESUME ||--o{ PDF_DOCUMENT : pdf
    APP_USER {
        attr user_id PK
        attr email UK
        attr password_hash
        attr role
        attr created_at
        attr updated_at
    }
    SEEKER_PROFILE {
        attr user_id PK,FK
        attr first_name
        attr last_name
        attr phone
        attr about
        attr avatar_key
        attr created_at
        attr updated_at
        attr city
        attr messenger
        attr contact_email
        attr preferred_position
        attr preferred_cities
        attr work_format
    }
    SEEKER_EXPERIENCE {
        attr experience_id PK
        attr seeker_id FK
        attr company_name
        attr position_title
        attr started_on
        attr ended_on
        attr is_current
        attr achievements
        attr position
        attr created_at
        attr updated_at
    }
    SEEKER_EDUCATION {
        attr education_id PK
        attr seeker_id FK
        attr record_type
        attr institution
        attr completion_year
        attr specialization
        attr position
        attr created_at
        attr updated_at
    }
    SEEKER_SKILL {
        attr skill_id PK
        attr seeker_id FK
        attr name
        attr position
        attr created_at
        attr updated_at
    }
    SEEKER_PROJECT {
        attr project_id PK
        attr seeker_id FK
        attr title
        attr url
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
    RESUME {
        attr resume_id PK
        attr seeker_id FK
        attr title
        attr description
        attr created_at
        attr updated_at
        attr deleted_at
        attr target_position
        attr city
        attr employment_type
        attr work_format
        attr work_schedule
        attr salary_from
        attr portfolio_url
        attr languages
    }
    RESUME_LEGACY_CONTENT {
        attr resume_id PK,FK
        attr experience_text
        attr education_text
        attr created_at
    }
    RESUME_CATEGORY {
        attr resume_id PK,FK
        attr category_id PK,FK
    }
    RESUME_EXPERIENCE {
        attr experience_id PK
        attr resume_id FK
        attr company_name
        attr position_title
        attr started_on
        attr ended_on
        attr is_current
        attr achievements
        attr position
        attr created_at
        attr updated_at
    }
    RESUME_EDUCATION {
        attr education_id PK
        attr resume_id FK
        attr record_type
        attr institution
        attr completion_year
        attr specialization
        attr position
        attr created_at
        attr updated_at
    }
    RESUME_SKILL {
        attr skill_id PK
        attr resume_id FK
        attr name
        attr position
        attr created_at
        attr updated_at
    }
    RESUME_PROJECT {
        attr project_id PK
        attr resume_id FK
        attr title
        attr url
        attr description
        attr position
        attr created_at
        attr updated_at
    }
    PDF_DOCUMENT {
        attr document_id PK
        attr resume_id FK
        attr file_path
        attr created_at
        attr updated_at
    }
```

Связи показаны для переходной единой БД. При физическом разделении межсервисные ссылки превращаются в ссылки по ID и gRPC-вызовы; MinIO/S3 хранит файлы по ключам из профилей и `pdf_document`.
