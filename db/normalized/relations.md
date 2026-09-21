# Описание таблиц

---

## USER

**Назначение:** базовая учётная запись. Хранит данные аутентификации и роль.

**Функциональные зависимости:**

```
{user_id} -> email, password_hash, role, created_at, updated_at
{email}   -> user_id, password_hash, role, created_at, updated_at
```

**Нормальные формы:**
- 1НФ: все значения атомарны
- 2НФ: PK не составной, частичных зависимостей нет
- 3НФ: нет транзитивных зависимостей
- НФБК: детерминанты `user_id` и `email` — потенциальные ключи

---

## EMPLOYER_PROFILE

**Назначение:** профиль работодателя (компании). 1:1 к `USER` при `role = 'employer'`.

**Функциональные зависимости:**

```
{user_id}      -> company_name, description, website, created_at, updated_at
{company_name} -> user_id, description, website, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## SEEKER_PROFILE

**Назначение:** профиль соискателя. 1:1 к `USER` при `role = 'seeker'`.

**Функциональные зависимости:**

```
{user_id} -> first_name, last_name, phone, about, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## CATEGORY

**Назначение:** справочник категорий для вакансий и резюме.

**Функциональные зависимости:**

```
{category_id} -> name, created_at, updated_at
{name}        -> category_id, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## VACANCY

**Назначение:** инвариантные поля вакансии. Изменяемые данные (title, description, salary) хранятся в `VACANCY_HISTORY`.

**Функциональные зависимости:**

```
{vacancy_id} -> employer_id, created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## VACANCY_CATEGORY

**Назначение:** связующая таблица M:N между `VACANCY` и `CATEGORY`.

**Функциональные зависимости:** нет неключевых атрибутов.

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## VACANCY_HISTORY

**Назначение:** все версии вакансии. Актуальная версия — с максимальным `version`.

**Функциональные зависимости:**

```
{vacancy_id, version} -> title, description, salary_from, salary_to, changed_at
```

**Нормальные формы:**
- 1НФ
- 2НФ: все неключевые атрибуты зависят от полного ключа `{vacancy_id, version}`
- 3НФ: нет зависимостей между неключевыми атрибутами
- НФБК: единственный детерминант — потенциальный ключ

---

## RESUME

**Назначение:** резюме соискателя.

**Функциональные зависимости:**

```
{resume_id} -> seeker_id, title, description, experience, education, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## RESUME_CATEGORY

**Назначение:** связующая таблица M:N между `RESUME` и `CATEGORY`.

**Функциональные зависимости:** нет неключевых атрибутов.

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## APPLICATION

**Назначение:** отклик соискателя на конкретную версию вакансии.

**Функциональные зависимости:**

```
{application_id} -> vacancy_id, vacancy_version, resume_id, cover_letter, created_at, updated_at
{vacancy_id, resume_id} -> application_id, vacancy_version, cover_letter, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## APPLICATION_STATUS_HISTORY

**Назначение:** журнал смены статусов отклика.

**Функциональные зависимости:**

```
{status_history_id} -> application_id, status, comment, changed_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## FAVORITE_VACANCY

**Назначение:** избранные вакансии пользователя.

**Функциональные зависимости:**

```
{user_id, vacancy_id} -> created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## NOTIFICATION_TOPIC

**Назначение:** справочник тем уведомлений и шаблонов их текстов.

**Функциональные зависимости:**

```
{topic_id} -> code, title, body_template, created_at, updated_at
{code}     -> topic_id, title, body_template, created_at, updated_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## NOTIFICATION

**Назначение:** конкретное уведомление пользователя.

**Функциональные зависимости:**

```
{notification_id} -> user_id, topic_id, rendered_text, is_read, created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## PDF_DOCUMENT

**Назначение:** сгенерированные PDF-файлы резюме. Файл хранится в S3/Minio, в БД — путь.

**Функциональные зависимости:**

```
{document_id} -> resume_id, file_path, created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## CHAT

**Назначение:** чат по конкретному отклику. 1:1 к `APPLICATION`.

**Функциональные зависимости:**

```
{chat_id}        -> application_id, created_at
{application_id} -> chat_id, created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.

---

## MESSAGE

**Назначение:** сообщения чата.

**Функциональные зависимости:**

```
{message_id} -> chat_id, sender_id, body, is_read, created_at
```

**Нормальные формы:** 1НФ, 2НФ, 3НФ, НФБК.