# Runtime Config для frontend HireNoon

`GET /api/runtime-config` возвращает только настройки, которые продуктовый backend разрешил для frontend. Контракт доступен в [runtime-config.openapi.yaml](runtime-config.openapi.yaml).

Запрашивайте endpoint с cookie `access_token`. Сервер берёт ID пользователя из JWT и возвращает готовые boolean для флагов. Результат можно использовать, чтобы скрыть или отключить кнопку создания вакансии или отправки отклика. Backend также проверяет флаг при записи и отвечает `403 {"error":"feature disabled"}`, если флаг успел измениться между чтением и отправкой формы.

Загрузите настройки сразу после входа и обновляйте их каждые 15 секунд. Сохраняйте предыдущие настройки при временной ошибке сети. При смене пользователя или выходе остановите опрос и очистите настройки предыдущей сессии.

```js
function watchRuntimeConfig(onConfig, onUnauthorized) {
  const controller = new AbortController()
  let pending = false

  async function refresh() {
    if (pending) return
    pending = true
    try {
      const response = await fetch('/api/runtime-config', {
        credentials: 'same-origin',
        cache: 'no-store',
        signal: controller.signal,
      })
      if (response.status === 401) {
        stop()
        onUnauthorized()
        return
      }
      if (response.ok) {
        const { data } = await response.json()
        if (!controller.signal.aborted) onConfig(data)
      }
    } catch {
    } finally {
      pending = false
    }
  }

  function stop() {
    clearInterval(timer)
    controller.abort()
  }

  const timer = setInterval(refresh, 15_000)
  void refresh()
  return stop
}
```

## Настройка монолита

Задайте `ONLINECONF_URL` и `ONLINECONF_SNAPSHOT_SECRET` в окружении backend. Секрет должен совпадать с секретом внутреннего endpoint OnlineConf. При Compose используйте адрес `http://onlineconf:8081` и запускайте оба сервиса с профилем `onlineconf`.

```bash
export ONLINECONF_URL=http://onlineconf:8081
docker compose --profile onlineconf up -d --build backend onlineconf
```

Пустой `ONLINECONF_URL` отключает обращения к OnlineConf. Монолит продолжает использовать лимит пяти категорий и разрешает обе операции по умолчанию. При включённом OnlineConf монолит пытается получить снимок при запуске и затем обновляет его каждые 15 секунд. HTTP-запросы продукта читают локальный кеш.

`vacancy.max_categories` принимает JSON-целое от 1 до 100. Значение другого типа или вне диапазона заменяется дефолтом `5`. Уменьшение лимита не удаляет существующие категории. Новый лимит применяется при создании вакансии и при явной замене списка категорий.
