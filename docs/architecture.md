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

### POST /api/register

Принимает JSON:

```json
{
  "email": "user@example.com",
  "password": "Password123"
}
```

Требования к паролю:

- длина от 8 до 72 байт
- хотя бы одна заглавная латинская буква
- хотя бы одна строчная латинская буква
- хотя бы одна цифра
- только ASCII-символы (без пробелов и кириллицы)

Возвращает JSON:

```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

## Токены

Используются два типа JWT, подписанных алгоритмом HS256:

- **access token** — короткоживущий, прикладывается к каждому защищённому запросу.
- **refresh token** — долгоживущий, используется только для получения новой пары токенов.

Внутри токена лежат: `uid` (id пользователя), `typ` (`access` или `refresh`), а также стандартные `exp`, `iat`, `sub`. При проверке контролируется подпись, срок жизни и соответствие типа ожидаемому.

## Конфигурация

Читается из переменных окружения:

- `JWT_SECRET` — секрет для подписи JWT (обязателен)
- `ACCESS_TOKEN_TTL` — время жизни access-токена (по умолчанию 15m)
- `REFRESH_TOKEN_TTL` — время жизни refresh-токена (по умолчанию 168h)

## Запуск

Задать `JWT_SECRET` и запустить сервер.

Linux / macOS:

```bash
JWT_SECRET=your-secret go run cmd/server/main.go
```

PowerShell:

```powershell
$env:JWT_SECRET="your-secret"
go run cmd/server/main.go
```

CMD:

```cmd
set JWT_SECRET=your-secret
go run cmd/server/main.go
```

Сервер слушает `:8080`.

## Пример запроса

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password123"}'
```