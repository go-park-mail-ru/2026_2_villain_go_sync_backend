# OnlineConf

OnlineConf использует отдельную базу `onlineconf` на том же PostgreSQL 17. Инициализация запускается после старта PostgreSQL и подходит для нового и уже используемого тома. Продуктовые миграции запускаются прежними командами. Все команды ниже выполняются из корня репозитория.

Замените `cluster_admin`, `product_reader` и `product_db` на свои значения. Для bootstrap учётной записи `product_reader` нужен доступ на чтение `public.app_user.user_id`. Пароли можно передать через `PGPASSFILE` с правами `0600`. Для удалённого PostgreSQL задайте режим TLS по требованиям окружения.

```bash
docker compose up -d postgres
export ONLINECONF_ADMIN_DATABASE_URL='postgresql://cluster_admin@127.0.0.1:5432/postgres?sslmode=disable'
export ONLINECONF_MIGRATE_DATABASE_URL='postgresql://onlineconf_migrator@127.0.0.1:5432/onlineconf?sslmode=disable'
export ONLINECONF_PRODUCT_DATABASE_URL='postgresql://product_reader@127.0.0.1:5432/product_db?sslmode=disable'
make onlineconf-db-init
```

Роли `onlineconf_migrator` и `onlineconf_runtime` создаются без паролей. При первом запуске настройте их аутентификацию. Для парольной схемы установите отдельные пароли:

```bash
psql -X "$ONLINECONF_ADMIN_DATABASE_URL" -c '\password onlineconf_migrator'
psql -X "$ONLINECONF_ADMIN_DATABASE_URL" -c '\password onlineconf_runtime'
```

Добавьте учётные данные мигратора в используемый файл паролей. Приложение будет подключаться к `onlineconf` под ролью `onlineconf_runtime`. Повторная инициализация сохраняет пароли и данные.

```bash
make onlineconf-migrate-up
make onlineconf-migrate-version
make onlineconf-admin-bootstrap ONLINECONF_OWNER_USER_ID=123
```

Замените `123` на ID существующего продуктового пользователя. Bootstrap проверяет ID в продуктовой БД и атомарно создаёт первого владельца с записью в истории. При наличии администраторов или истории административных изменений bootstrap отклоняется. Последнего владельца нельзя удалить или понизить до `admin`. При конкурентных изменениях PostgreSQL может откатить одну транзакцию из-за конфликта блокировок или сериализации.

Ключ и подтип создаются в одной транзакции. Запись истории обычного изменения должна выполняться приложением в той же транзакции. Для персонального исключения `change_history.resource_id` содержит ID ключа, а снимок содержит `user_id`.

`make onlineconf-migrate-down` откатывает одну миграцию. Для `000002` откат отклоняется при наличии раскаток, регламентов, групп или дополнительных прав. Для `000001` откат отклоняется при наличии истории, ключей, администраторов или изменений исходных scope. После отказа `golang-migrate` помечает версию как `dirty`, хотя SQL-транзакция сохраняет данные. Перед исправлением состояния сверяйте фактическую схему и версию.

## Проверка изоляции

Роли OnlineConf получают доступ к продуктовой БД через `PUBLIC`, если у `PUBLIC` разрешены `CONNECT` или `TEMPORARY`. PostgreSQL суммирует эти права с персональными grants. Инициализация OnlineConf оставляет ACL продуктовой БД в прежнем состоянии.

Чтобы проверить эффективные права на конкретном экземпляре, запустите запросы от учётной записи с доступом к каталогам продуктовой БД:

```bash
psql -X "$ONLINECONF_PRODUCT_DATABASE_URL" -v ON_ERROR_STOP=1 -f db/onlineconf/checks/product_privileges.sql
```

Если нужен полный запрет подключения, потребуется согласовать продуктовые ACL: сначала выдать `CONNECT` необходимым продуктовым ролям, затем отозвать его у `PUBLIC`. Также можно обеспечить запрет подключения отдельно через `pg_hba.conf` или сетевой прокси.

Изолированный прогон выполняется командой `bash db/onlineconf/checks/run.sh`. Он запускает временный PostgreSQL 17 с отдельной тестовой продуктовой БД. Нужны Docker, `psql`, OpenSSL и `golang-migrate`. Путь к бинарнику миграций можно задать через `ONLINECONF_MIGRATE_BIN`.
