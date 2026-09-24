# Архитектура backend

Основной поток обработки запроса:

```text
HTTP request
    ↓
handler
    ↓
service
    ↓
repository
    ↓
database
````

## Структура

```text
cmd/server/       — точка входа приложения

internal/
├── handler/      — обработка HTTP-запросов и формирование ответов
├── server/       — настройка HTTP-сервера и маршрутов
├── service/      — бизнес-логика приложения
├── repository/   — работа с базой данных
├── model/        — основные структуры данных
└── middleware/   — общая обработка запросов, например авторизация и CORS
```

Сейчас реализован минимальный HTTP-сервис с ручками:

```http
GET /api/health     - состояние сервиса 
POST /api/register  - ручка регистрации
```
