# Запуск OnlineConf

Сначала инициализируйте БД, примените миграции и назначьте первого владельца по инструкции в [db/onlineconf/README.md](../db/onlineconf/README.md).
Для раскаток, групп и прав требуется миграция `000002`. Порядок работы с ними описан в [руководстве OnlineConf](onlineconf-advanced.md).

Для сервиса нужны следующие переменные окружения.

| Переменная | Значение |
|---|---|
| `ONLINECONF_DATABASE_URL` | DSN базы `onlineconf` для роли `onlineconf_runtime`. |
| `ONLINECONF_PRODUCT_READER_DATABASE_URL` | DSN отдельной read-only роли в продуктовой БД. |
| `JWT_SECRET` | Тот же ключ проверки access JWT, что использует HireNoon. |
| `ONLINECONF_SNAPSHOT_SECRET` | Отдельный случайный Bearer-секрет для внутреннего снимка. |
| `ONLINECONF_ADDR` | Адрес standalone-сервиса. По умолчанию `:8081`. |
| `ONLINECONF_PORT` | Порт хоста при запуске через Compose. По умолчанию `8081`. |

Для продуктового подключения подготовьте отдельную учётную запись с `CONNECT` к продуктовой БД, `USAGE` на схему `public` и только `SELECT (user_id)` на `public.app_user`. Если часть этих прав уже даёт `PUBLIC`, повторные grants не нужны. Адаптер устанавливает read-only режим и проверяет эффективные права на `app_user`. Подготовку роли выполняет администратор БД после согласования. OnlineConf не меняет продуктовые ACL и не использует `onlineconf_runtime` для продуктового подключения.

Для локального запуска передайте DSN, доступные с хоста, и выполните `make onlineconf-run`. Для Compose используйте DSN с адресом `postgres:5432`, если обе БД в существующем контейнере PostgreSQL, и запустите:

```bash
docker compose --profile onlineconf up -d --build onlineconf
```

`GET /health` проверяет обе зависимости. Внутренний `GET /internal/onlineconf/snapshot` принимает `Authorization: Bearer <ONLINECONF_SNAPSHOT_SECRET>`. Передавайте секрет только доверенным сервисам по защищённому каналу. Контракт административного API и снимка находится в [onlineconf.openapi.yaml](onlineconf.openapi.yaml).

Изолированные тесты HTTP и запуска сервиса:

```bash
ONLINECONF_CHECK_HTTP=1 bash db/onlineconf/checks/run.sh
```
