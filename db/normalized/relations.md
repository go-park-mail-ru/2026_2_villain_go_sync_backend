# HireNoon — нормализованная схема БД

Целевая переходная PostgreSQL-схема после миграций `000001`–`000017`: **37 таблиц**. Для Auth, Seeker, Employer и Application показаны доменные границы, но до выделения микросервисов это **одна БД** с междоменными внешними ключами.

## Обоснование нормальных форм

1НФ: значения полей атомарны; повторяющиеся элементы (опыт, образование, навыки, проекты, категории, заметки) вынесены в отдельные отношения. Свободный текст описания/предпочтительных городов — один пользовательский текст, не SQL-массив. `RESUME_LEGACY_CONTENT` хранит свободный текст опыта и образования. Отдельные записи опыта и образования хранятся в своих отношениях.

2НФ: все неключевые атрибуты полностью зависят от соответствующего кандидатного ключа; в отношениях со составным ключом нет зависимости неключевого поля от его части.

3НФ: нет транзитивных зависимостей между неключевыми атрибутами. Например, в `APPLICATION` нет `resume_id`, `employer_id` и `current_status`, чтобы не повторять определения владельца и последнего события.

НФБК: все указанные нетривиальные функциональные зависимости имеют слева **потенциальный ключ**. Снимки — отдельные факты о состоянии резюме в момент конкретного отклика, а не кэш текущего резюме.

Необязательные поля без `DEFAULT` имеют неявное значение `NULL`. Для обязательных ключей связей, порядковых номеров, статусов событий и содержимого (например, email, роль, название, текст сообщения, ключ файла) значения задаёт вызывающая операция: они зависят от конкретного объекта или действия. Суррогатные ключи генерирует `IDENTITY`. Репозитории обновляют `updated_at` при записи изменений. У связей M:N и неизменяемых исторических фактов отсутствие `updated_at` объясняется ниже.

Денежные суммы хранятся в `NUMERIC(12, 2)` для точного хранения десятичных значений. Календарные даты и сроки без времени суток хранятся в `DATE`, моменты событий — в `TIMESTAMPTZ`. Для текстовых полей используется `TEXT`. Бизнес-ограничения на содержимое задаются через `CHECK`.

## Описание отношений

### APP_USER

Аккаунт, email для входа, хэш пароля, роль seeker/employer.

**Атрибуты:** `user_id`, `email`, `password_hash`, `role`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(user_id)`, `(email)`.

**Функциональные зависимости:**

```text
{user_id} -> email, password_hash, role, created_at, updated_at
{email} -> user_id, password_hash, role, created_at, updated_at
```

**Внешние ключи:** нет.

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### SEEKER_PROFILE

Профиль соискателя хранит контактную почту отдельно от email авторизации. Фамилия обязательна при сохранении, желаемая должность в профиле необязательна. Профили без фамилии, созданные до миграции `000013`, доступны для чтения. Фамилию нужно указать при следующем сохранении.

**Атрибуты:** `user_id`, `first_name`, `last_name`, `phone`, `about`, `avatar_key`, `created_at`, `updated_at`, `city`, `messenger`, `contact_email`, `preferred_position`, `preferred_cities`, `work_format`.

**Потенциальные ключи:** `(user_id)`.

**Функциональные зависимости:**

```text
{user_id} -> first_name, last_name, phone, about, avatar_key, created_at, updated_at, city, messenger, contact_email, preferred_position, preferred_cities, work_format
```

**Внешние ключи:** `user_id` → `APP_USER(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### SEEKER_EXPERIENCE

До 10 записей об опыте соискателя; is_current соответствует галочке «По настоящее время». Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `experience_id`, `seeker_id`, `company_name`, `position_title`, `started_on`, `ended_on`, `is_current`, `achievements`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(experience_id)`, `(seeker_id, position)`.

**Функциональные зависимости:**

```text
{experience_id} -> seeker_id, company_name, position_title, started_on, ended_on, is_current, achievements, position, created_at, updated_at
{seeker_id, position} -> experience_id, company_name, position_title, started_on, ended_on, is_current, achievements, created_at, updated_at
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### SEEKER_EDUCATION

До 10 записей об образовании и курсах соискателя. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `education_id`, `seeker_id`, `record_type`, `institution`, `completion_year`, `specialization`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(education_id)`, `(seeker_id, position)`.

**Функциональные зависимости:**

```text
{education_id} -> seeker_id, record_type, institution, completion_year, specialization, position, created_at, updated_at
{seeker_id, position} -> education_id, record_type, institution, completion_year, specialization, created_at, updated_at
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### SEEKER_SKILL

Список навыков соискателя, имена уникальны без учёта регистра и крайних пробелов. Уникальность названия без учёта регистра и крайних пробелов обеспечивает уникальный функциональный индекс.

**Атрибуты:** `skill_id`, `seeker_id`, `name`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(skill_id)`, `(seeker_id, position)`.

**Функциональные зависимости:**

```text
{skill_id} -> seeker_id, name, position, created_at, updated_at
{seeker_id, position} -> skill_id, name, created_at, updated_at
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### SEEKER_PROJECT

До 5 проектов в профиле соискателя. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `project_id`, `seeker_id`, `title`, `url`, `description`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(project_id)`, `(seeker_id, position)`.

**Функциональные зависимости:**

```text
{project_id} -> seeker_id, title, url, description, position, created_at, updated_at
{seeker_id, position} -> project_id, title, url, description, created_at, updated_at
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### EMPLOYER_PROFILE

Профиль компании: одно описание, контакты, условия и ссылка на логотип.

**Атрибуты:** `user_id`, `company_name`, `description`, `website`, `logo_key`, `created_at`, `updated_at`, `industry`, `team_size`, `main_city`, `benefits`, `hiring_email`, `office_address`.

**Потенциальные ключи:** `(user_id)`.

**Функциональные зависимости:**

```text
{user_id} -> company_name, description, website, logo_key, created_at, updated_at, industry, team_size, main_city, benefits, hiring_email, office_address
```

**Внешние ключи:** `user_id` → `APP_USER(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### EMPLOYER_TECHNOLOGY

Направления и технологии компании. Уникальность названия без учёта регистра и крайних пробелов обеспечивает уникальный функциональный индекс.

**Атрибуты:** `technology_id`, `employer_id`, `name`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(technology_id)`, `(employer_id, position)`.

**Функциональные зависимости:**

```text
{technology_id} -> employer_id, name, position, created_at, updated_at
{employer_id, position} -> technology_id, name, created_at, updated_at
```

**Внешние ключи:** `employer_id` → `EMPLOYER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### EMPLOYER_CULTURE

До двух принципов культуры компании: заголовок, краткая характеристика и описание. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `culture_id`, `employer_id`, `title`, `short_description`, `description`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(culture_id)`, `(employer_id, position)`.

**Функциональные зависимости:**

```text
{culture_id} -> employer_id, title, short_description, description, position, created_at, updated_at
{employer_id, position} -> culture_id, title, short_description, description, created_at, updated_at
```

**Внешние ключи:** `employer_id` → `EMPLOYER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### CATEGORY

Справочник категорий, общий в переходном монолите.

**Атрибуты:** `category_id`, `name`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(category_id)`, `(name)`.

**Функциональные зависимости:**

```text
{category_id} -> name, created_at, updated_at
{name} -> category_id, created_at, updated_at
```

**Внешние ключи:** нет.

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### VACANCY

Вакансия работодателя с отметками времени создания и обновления. Необязательная дата `apply_until` хранится справочно. Её наступление не меняет видимость вакансии, возможность редактирования или приём откликов.

**Атрибуты:** `vacancy_id`, `employer_id`, `created_at`, `apply_until`, `updated_at`.

**Потенциальные ключи:** `(vacancy_id)`.

**Функциональные зависимости:**

```text
{vacancy_id} -> employer_id, created_at, apply_until, updated_at
```

**Внешние ключи:** `employer_id` → `EMPLOYER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### VACANCY_HISTORY

Исторические версии вакансии, включая условия, навыки через отдельную таблицу.

**Атрибуты:** `vacancy_id`, `version`, `title`, `description`, `salary_from`, `salary_to`, `changed_at`, `currency`, `responsibilities`, `requirements`, `offer`, `experience_level`, `employment_type`, `work_format`, `work_schedule`, `work_hours`, `city`, `office_address`.

**Потенциальные ключи:** `(vacancy_id, version)`.

**Функциональные зависимости:**

```text
{vacancy_id, version} -> title, description, salary_from, salary_to, changed_at, currency, responsibilities, requirements, offer, experience_level, employment_type, work_format, work_schedule, work_hours, city, office_address
```

**Внешние ключи:** `vacancy_id` → `VACANCY(vacancy_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### VACANCY_CATEGORY

Связь категории с вакансией.

**Атрибуты:** `vacancy_id`, `category_id`.

**Потенциальные ключи:** `(vacancy_id, category_id)`.

**Функциональные зависимости:**

```text
Нетривиальных зависимостей нет: отношение состоит только из атрибутов составного ключа.
```

**Внешние ключи:** `vacancy_id` → `VACANCY(vacancy_id)` (`ON DELETE CASCADE`); `category_id` → `CATEGORY(category_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### VACANCY_VERSION_SKILL

Навыки конкретной версии вакансии, сохраняющие историю. Уникальность названия без учёта регистра и крайних пробелов обеспечивает уникальный функциональный индекс.

**Атрибуты:** `skill_id`, `vacancy_id`, `version`, `name`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(skill_id)`, `(vacancy_id, version, position)`.

**Функциональные зависимости:**

```text
{skill_id} -> vacancy_id, version, name, position, created_at, updated_at
{vacancy_id, version, position} -> skill_id, name, created_at, updated_at
```

**Внешние ключи:** `(vacancy_id,version)` → `VACANCY_HISTORY((vacancy_id,version))` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME

Резюме с условиями желаемой работы и логическим удалением; содержательные секции в дочерних таблицах.

**Атрибуты:** `resume_id`, `seeker_id`, `title`, `description`, `created_at`, `updated_at`, `deleted_at`, `target_position`, `city`, `employment_type`, `work_format`, `work_schedule`, `salary_from`, `portfolio_url`, `languages`.

**Потенциальные ключи:** `(resume_id)`.

**Функциональные зависимости:**

```text
{resume_id} -> seeker_id, title, description, created_at, updated_at, deleted_at, target_position, city, employment_type, work_format, work_schedule, salary_from, portfolio_url, languages
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_LEGACY_CONTENT

Свободный текст опыта и образования из прежней схемы и действующего API. Структурированные записи хранятся в `RESUME_EXPERIENCE` и `RESUME_EDUCATION`.

**Атрибуты:** `resume_id`, `experience_text`, `education_text`, `created_at`.

**Потенциальные ключи:** `(resume_id)`.

**Функциональные зависимости:**

```text
{resume_id} -> experience_text, education_text, created_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_CATEGORY

Связь категории с резюме.

**Атрибуты:** `resume_id`, `category_id`.

**Потенциальные ключи:** `(resume_id, category_id)`.

**Функциональные зависимости:**

```text
Нетривиальных зависимостей нет: отношение состоит только из атрибутов составного ключа.
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`); `category_id` → `CATEGORY(category_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_EXPERIENCE

Опыт конкретного резюме, независимый от профиля соискателя. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `experience_id`, `resume_id`, `company_name`, `position_title`, `started_on`, `ended_on`, `is_current`, `achievements`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(experience_id)`, `(resume_id, position)`.

**Функциональные зависимости:**

```text
{experience_id} -> resume_id, company_name, position_title, started_on, ended_on, is_current, achievements, position, created_at, updated_at
{resume_id, position} -> experience_id, company_name, position_title, started_on, ended_on, is_current, achievements, created_at, updated_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_EDUCATION

Образование конкретного резюме, независимое от профиля соискателя. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `education_id`, `resume_id`, `record_type`, `institution`, `completion_year`, `specialization`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(education_id)`, `(resume_id, position)`.

**Функциональные зависимости:**

```text
{education_id} -> resume_id, record_type, institution, completion_year, specialization, position, created_at, updated_at
{resume_id, position} -> education_id, record_type, institution, completion_year, specialization, created_at, updated_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_SKILL

Навыки конкретного резюме, независимые от профиля соискателя. Уникальность названия без учёта регистра и крайних пробелов обеспечивает уникальный функциональный индекс.

**Атрибуты:** `skill_id`, `resume_id`, `name`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(skill_id)`, `(resume_id, position)`.

**Функциональные зависимости:**

```text
{skill_id} -> resume_id, name, position, created_at, updated_at
{resume_id, position} -> skill_id, name, created_at, updated_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### RESUME_PROJECT

До пяти проектов конкретного резюме, независимых от профиля соискателя. Ограничение диапазона `position` вместе с уникальностью `(владелец, position)` реализует лимит формы.

**Атрибуты:** `project_id`, `resume_id`, `title`, `url`, `description`, `position`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(project_id)`, `(resume_id, position)`.

**Функциональные зависимости:**

```text
{project_id} -> resume_id, title, url, description, position, created_at, updated_at
{resume_id, position} -> project_id, title, url, description, created_at, updated_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### PDF_DOCUMENT

Метаданные PDF резюме; файл в MinIO/S3 по ключу file_path.

**Атрибуты:** `document_id`, `resume_id`, `file_path`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(document_id)`.

**Функциональные зависимости:**

```text
{document_id} -> resume_id, file_path, created_at, updated_at
```

**Внешние ключи:** `resume_id` → `RESUME(resume_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION

Отклик соискателя на определённую версию вакансии: один отклик на пару (vacancy_id,seeker_id).

**Атрибуты:** `application_id`, `vacancy_id`, `vacancy_version`, `seeker_id`, `cover_letter`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(application_id)`, `(vacancy_id, seeker_id)`.

**Функциональные зависимости:**

```text
{application_id} -> vacancy_id, vacancy_version, seeker_id, cover_letter, created_at, updated_at
{vacancy_id, seeker_id} -> application_id, vacancy_version, cover_letter, created_at, updated_at
```

**Внешние ключи:** `seeker_id` → `SEEKER_PROFILE(user_id)` (`ON DELETE RESTRICT`); `(vacancy_id,vacancy_version)` → `VACANCY_HISTORY((vacancy_id,version))` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_STATUS_HISTORY

Неизменяемые события статусов, включая обязательный код причины отказа и необязательный комментарий для кандидата.

**Атрибуты:** `status_history_id`, `application_id`, `status`, `comment`, `changed_at`, `reason_code`, `changed_by`.

**Потенциальные ключи:** `(status_history_id)`.

**Функциональные зависимости:**

```text
{status_history_id} -> application_id, status, comment, changed_at, reason_code, changed_by
```

**Внешние ключи:** `application_id` → `APPLICATION(application_id)` (`ON DELETE CASCADE`); `changed_by` → `APP_USER(user_id)` (`ON DELETE SET NULL`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_RESUME_SNAPSHOT

Снимок данных резюме при отклике; источник может быть впоследствии удалён.

**Атрибуты:** `application_id`, `source_resume_id`, `title`, `description`, `legacy_experience_text`, `legacy_education_text`, `target_position`, `city`, `employment_type`, `work_format`, `work_schedule`, `salary_from`, `portfolio_url`, `languages`, `captured_at`, `is_reconstructed`, `is_sealed`.

**Потенциальные ключи:** `(application_id)`.

**Функциональные зависимости:**

```text
{application_id} -> source_resume_id, title, description, legacy_experience_text, legacy_education_text, target_position, city, employment_type, work_format, work_schedule, salary_from, portfolio_url, languages, captured_at, is_reconstructed, is_sealed
```

**Внешние ключи:** `application_id` → `APPLICATION(application_id)` (`ON DELETE CASCADE`); `source_resume_id` → `RESUME(resume_id)` (`ON DELETE SET NULL`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_SNAPSHOT_EXPERIENCE

Опыт, скопированный в исторический снимок резюме.

**Атрибуты:** `application_id`, `position`, `company_name`, `position_title`, `started_on`, `ended_on`, `is_current`, `achievements`.

**Потенциальные ключи:** `(application_id, position)`.

**Функциональные зависимости:**

```text
{application_id, position} -> company_name, position_title, started_on, ended_on, is_current, achievements
```

**Внешние ключи:** `application_id` → `APPLICATION_RESUME_SNAPSHOT(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_SNAPSHOT_EDUCATION

Образование, скопированное в исторический снимок резюме.

**Атрибуты:** `application_id`, `position`, `record_type`, `institution`, `completion_year`, `specialization`.

**Потенциальные ключи:** `(application_id, position)`.

**Функциональные зависимости:**

```text
{application_id, position} -> record_type, institution, completion_year, specialization
```

**Внешние ключи:** `application_id` → `APPLICATION_RESUME_SNAPSHOT(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_SNAPSHOT_SKILL

Навыки, скопированные в исторический снимок резюме.

**Атрибуты:** `application_id`, `position`, `name`.

**Потенциальные ключи:** `(application_id, position)`.

**Функциональные зависимости:**

```text
{application_id, position} -> name
```

**Внешние ключи:** `application_id` → `APPLICATION_RESUME_SNAPSHOT(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_SNAPSHOT_PROJECT

Проекты, скопированные в исторический снимок резюме.

**Атрибуты:** `application_id`, `position`, `title`, `url`, `description`.

**Потенциальные ключи:** `(application_id, position)`.

**Функциональные зависимости:**

```text
{application_id, position} -> title, url, description
```

**Внешние ключи:** `application_id` → `APPLICATION_RESUME_SNAPSHOT(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_SNAPSHOT_CATEGORY

Названия категорий на момент отклика, независимо от последующего изменения справочника.

**Атрибуты:** `application_id`, `category_name`.

**Потенциальные ключи:** `(application_id, category_name)`.

**Функциональные зависимости:**

```text
Нетривиальных зависимостей нет: отношение состоит только из атрибутов составного ключа.
```

**Внешние ключи:** `application_id` → `APPLICATION_RESUME_SNAPSHOT(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### APPLICATION_NOTE

Внутренние заметки работодателя о кандидате; никогда не видны соискателю.

**Атрибуты:** `note_id`, `application_id`, `author_id`, `body`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(note_id)`.

**Функциональные зависимости:**

```text
{note_id} -> application_id, author_id, body, created_at, updated_at
```

**Внешние ключи:** `application_id` → `APPLICATION(application_id)` (`ON DELETE CASCADE`); `author_id` → `APP_USER(user_id)` (`ON DELETE SET NULL`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### FAVORITE_VACANCY

Факт добавления вакансии пользователем в избранное.

**Атрибуты:** `user_id`, `vacancy_id`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(user_id, vacancy_id)`.

**Функциональные зависимости:**

```text
{user_id, vacancy_id} -> created_at, updated_at
```

**Внешние ключи:** `user_id` → `APP_USER(user_id)` (`ON DELETE CASCADE`); `vacancy_id` → `VACANCY(vacancy_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### VACANCY_NOTE

Личная заметка пользователя к вакансии, не зависит от нахождения вакансии в избранном.

**Атрибуты:** `user_id`, `vacancy_id`, `body`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(user_id, vacancy_id)`.

**Функциональные зависимости:**

```text
{user_id, vacancy_id} -> body, created_at, updated_at
```

**Внешние ключи:** `user_id` → `APP_USER(user_id)` (`ON DELETE CASCADE`); `vacancy_id` → `VACANCY(vacancy_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### NOTIFICATION_TOPIC

Шаблоны типов уведомлений.

**Атрибуты:** `topic_id`, `code`, `title`, `body_template`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(topic_id)`, `(code)`.

**Функциональные зависимости:**

```text
{topic_id} -> code, title, body_template, created_at, updated_at
{code} -> topic_id, title, body_template, created_at, updated_at
```

**Внешние ключи:** нет.

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### NOTIFICATION

Созданные уведомления получателя и состояние прочтения.

**Атрибуты:** `notification_id`, `user_id`, `topic_id`, `rendered_text`, `is_read`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(notification_id)`.

**Функциональные зависимости:**

```text
{notification_id} -> user_id, topic_id, rendered_text, is_read, created_at, updated_at
```

**Внешние ключи:** `user_id` → `APP_USER(user_id)` (`ON DELETE CASCADE`); `topic_id` → `NOTIFICATION_TOPIC(topic_id)` (`ON DELETE NO ACTION`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### CHAT

Единственный чат, привязанный к конкретному отклику.

**Атрибуты:** `chat_id`, `application_id`, `created_at`.

**Потенциальные ключи:** `(chat_id)`, `(application_id)`.

**Функциональные зависимости:**

```text
{chat_id} -> application_id, created_at
{application_id} -> chat_id, created_at
```

**Внешние ключи:** `application_id` → `APPLICATION(application_id)` (`ON DELETE CASCADE`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

### MESSAGE

Сообщение в чате, отправитель проверяется по участникам отклика.

**Атрибуты:** `message_id`, `chat_id`, `sender_id`, `body`, `is_read`, `created_at`, `updated_at`.

**Потенциальные ключи:** `(message_id)`.

**Функциональные зависимости:**

```text
{message_id} -> chat_id, sender_id, body, is_read, created_at, updated_at
```

**Внешние ключи:** `chat_id` → `CHAT(chat_id)` (`ON DELETE CASCADE`); `sender_id` → `APP_USER(user_id)` (`ON DELETE NO ACTION`).

**Нормальные формы:** 1НФ, 2НФ, 3НФ и НФБК; все нетривиальные зависимости определяются кандидатными ключами.

---

## Дополнительные правила целостности и удаления

- `APPLICATION` хранит `(vacancy_id, vacancy_version, seeker_id)` и имеет `UNIQUE(vacancy_id, seeker_id)`: один соискатель не может дважды откликнуться на одну вакансию, даже с другим резюме. Сведения об отправленном резюме находятся в `APPLICATION_RESUME_SNAPSHOT`.
- `APPLICATION_RESUME_SNAPSHOT.source_resume_id` — необязательная FK-ссылка с `ON DELETE SET NULL`; физическое удаление исходного резюме **не удаляет** отклик или снимок. `APPLICATION.seeker_id` имеет FK с `RESTRICT`, поэтому удаление пользователя с действующей историей откликов запрещено до отдельной процедуры архивирования/обезличивания.
- API физически удаляет вакансию. `ON DELETE CASCADE` удаляет её версии, связи с категориями, записи избранного, заметки и отклики вместе со снимками, статусами и чатами. Триггеры разрешают удалять отклики и их снимки только в этом каскаде.
- `APPLICATION_STATUS_HISTORY` — последовательность событий, текущий статус определяется последней записью по `(changed_at, status_history_id)`.
- Новые отказы обязаны содержать один из семи кодов `reason_code`. Для исторических отказов без кода используется `CHECK ... NOT VALID`; новые вставки проверяются. Причина и комментарий отображаются соискателю, `APPLICATION_NOTE` остаётся доступной только работодателю.
- При создании нового отклика функция `create_application` проверяет владельца резюме, наличие вакансии и её актуальную версию, вставляет отклик, нормализованный снимок и событие `created` в одной транзакции. `UNIQUE` защищает от конкурентных повторов. При создании отклика функция выставляет `is_sealed` после копирования всех дочерних записей. Отложенный триггер проверяет наличие снимка с завершённой записью на `COMMIT`. Триггеры запрещают править содержимое снимка, удалять его отдельно от отклика, добавлять дочерние записи после создания снимка и очищать таблицы снимков через `TRUNCATE`. При удалении исходного резюме `ON DELETE SET NULL` обнуляет только ссылку `source_resume_id`.
- При параллельных изменениях резюме триггеры дочерних записей блокируют родительскую строку `resume`. Та же строка блокируется при создании снимка, чтобы не получить смесь разных версий в переходном монолите.
- `VACANCY_NOTE` не зависит от `FAVORITE_VACANCY`: удаление записи избранного не удаляет заметку.
- Исторические `VACANCY_HISTORY`, `APPLICATION_STATUS_HISTORY`, `APPLICATION_RESUME_SNAPSHOT` и дочерние снимки не редактируются в продуктовой логике. Их `changed_at`/`captured_at` эквивалентны времени создания, `updated_at` отсутствует по смыслу. У `RESUME_LEGACY_CONTENT` собственный `created_at`. Время изменения текста фиксируется в `RESUME.updated_at` в той же операции. У `CHAT` только факт создания, так что `created_at` достаточен.
- Таблицы `VACANCY_CATEGORY` и `RESUME_CATEGORY` — чистые пары FK с составным PK; их существование и удаление обозначают изменение связи, отдельные `created_at`/`updated_at` не добавляются. Для `APPLICATION_SNAPSHOT_*` не вводятся timestamps обновления, так как это неизменяемые части исторического факта.
- Указанные индексы в `000013`–`000017` служат только уникальности и ключам. Старые необязательные индексы из `000008` удалены миграцией `000017`; перед будущей оптимизацией используем `EXPLAIN ANALYZE` и метрики.

## Физические и межсервисные границы

| Область | Собственная БД в будущем | Основные таблицы |
|---|---|---|
| Auth | `auth_db` | `app_user` |
| Seeker | `seeker_db` | `seeker_profile`, `resume`, дочерние таблицы, `pdf_document` |
| Employer | `employer_db` | `employer_profile`, `vacancy`, `vacancy_history`, `category`, `favorite_vacancy`, `vacancy_note` |
| Application | `application_db` | `application`, снимки, статусы, заметки, чаты, уведомления |

После разделения базы не смогут иметь FK на таблицы других сервисов. Понадобятся gRPC-проверки идентификаторов, сервисная операция создания снимка и изменение триггера проверки участников сообщений. Справочник `category` между резюме и вакансиями сейчас общий: перед физическим разделением ему понадобится отдельная политика владения/синхронизации.

## Данные вне PostgreSQL

- MinIO/S3 хранит бинарные файлы аватаров, логотипов и PDF. Поля `seeker_profile.avatar_key`, `employer_profile.logo_key`, `pdf_document.file_path` хранят только объектные ключи.
- JWT выдаёт Auth Service; пароль хранится только в `app_user.password_hash`. Если будет подключён Redis для сессий, его структуры и ключи нужно отдельно документировать; текущая схема Redis не требует.
- WebSocket-чаты используют таблицы `chat` и `message`; сообщения остаются в PostgreSQL, а сетевое соединение не является реляционной сущностью.

## Совместимость с текущим API

- API профиля соискателя и резюме принимает и возвращает списки `experience_items`, `education_items`, `skills` и `projects`. API профиля работодателя работает со списками `technologies` и `culture`. Порядок элементов в запросе задаёт `position`. При обновлении переданный список заменяется целиком, пустой список очищает раздел, а пропущенный список сохраняет данные. Основная запись и все переданные списки изменяются одной транзакцией. Даты опыта передаются в формате `YYYY-MM-DD`. Текстовые поля резюме `experience` и `education` по-прежнему доступны.
- Поле `target_position` обязательно при создании и обновлении резюме. Старые резюме без целевой должности доступны для чтения. При обновлении API сохраняет значения необязательных полей, если они пропущены в запросе. Пустая строка очищает необязательные текстовые поля, `null` очищает `salary_from`. Поле `deleted_at` используется внутри репозитория для фильтрации. Свободные тексты опыта и образования доступны через прежние поля API и записываются атомарно с резюме.
- При создании вакансии и сохранении новой версии обязательны непустые `responsibilities` и `requirements`. Версии, существовавшие до миграции `000015`, остаются доступными для чтения. `CHECK ... NOT VALID` сохраняет эти строки и проверяет все новые вставки и обновления. При частичном обновлении API берёт значения из текущей версии для полей, отсутствующих в запросе. Чтобы сохранить новую версию старой вакансии с незаполненными полями, их нужно передать в запросе. При создании и изменении вакансии API принимает также условия, уровень опыта, тип занятости, формат и график работы, рабочие часы, город и адрес офиса. Пропущенные поля и навыки предыдущей версии сохраняются. Пустая строка очищает необязательное текстовое поле. Срок откликов возвращается в ответе как справочная дата.
- Список откликов читает заголовок резюме из снимка. После удаления исходного резюме `resume_id` в списке имеет значение `null`.
- Снимки старых откликов восстановлены из данных на момент миграции и помечены `is_reconstructed = TRUE`. Миграция `000016` останавливается, если у одного соискателя есть несколько старых откликов на одну вакансию.


## Статусы откликов в API

Работодатель меняет статус отклика на свою вакансию через `PATCH /api/employer/applications/{id}/status`. Тело запроса содержит `status`, при отказе — обязательный `reason_code`, и может содержать `comment`. Доступные статусы — `reviewing`, `interview`, `offer` и `rejected`. Обновление `updated_at` у отклика и запись события выполняются атомарно. `created` устанавливается автоматически

Соискатель получает свои отклики через `GET /api/seeker/applications`. Ответ содержит последний `status`, `reason_code`, `status_comment` и `status_changed_at`. Код причины и комментарий могут быть `null`, в том числе у старых отказов без кода.
