# Автотесты

Отдельный Playwright-проект для E2E, UI и security-проверок схемы «ферма». В CI пока не подключён: запускается вручную против stage.

## Виды тестов в проекте

| Вид | Где | Инструмент | Что проверяет и почему так |
|---|---|---|---|
| Юнит-тесты backend | `backend/*/internal` | `go test` | Домен, риск, парсеры без БД. Быстрые и изолированные |
| Интеграционные backend | `backend/graph-api/internal/app/*_integration_test.go` | `make integration-test`, testcontainers + Postgres | HTTP-ручки, SQL, сборка групп и решение на реальной БД; сценарии GSD-28 читаются из JSON |
| Фронтовые компонентные | `frontend/graph-frontapp/src/*.test.tsx` | Vitest + Testing Library + MSW + jsdom | Список, карточка, граф, polling и ошибки на моках. Быстро и без окружения |
| E2E API | `tests/api` | Playwright `APIRequestContext` | Реальный stage: admin-ручка запускает сценарий, graph-api создаёт группу, решение пишется в БД. Ловит расхождения между сервисами |
| E2E UI | `tests/e2e` | Playwright, Chromium | Путь аналитика в браузере: список → граф → карточка → решение |
| UI / кросс-браузер | `tests/ui` | Playwright Chromium + Firefox + WebKit, `toHaveScreenshot` | Вёрстка, скроллы, панели и пустые/ошибочные состояния в трёх движках |
| Security | `tests/security` + инструменты | Playwright, OWASP ZAP, Trivy, `govulncheck`, `pnpm audit` | Валидация входа и XSS автоматизированы; сканеры зависимостей и DAST запускаются отдельно |

Почему Playwright: один раннер для API и браузера, три движка, авто-ожидания, trace и видео при падении, встроенные скриншоты и TypeScript, как у фронта. Для E2E не нужен отдельный сервис. Юнит и интеграционные тесты команды не дублируем.

## Запуск

```bash
cd test
pnpm install
pnpm run install:browsers
BASE_URL=http://pr-40.shugraph.feytox.ru pnpm test
```

В PowerShell:

```powershell
$env:BASE_URL = "http://pr-40.shugraph.feytox.ru"
pnpm test
```

На удалённом stage бывают обрывы сети — можно включить один ретрай:

```powershell
$env:PW_RETRIES = "1"
```

Если браузеры Playwright не скачаны, Chromium-проекты можно запустить в системном Chrome:

```powershell
$env:PW_CHANNEL = "chrome"
pnpm test:e2e
```

Firefox и WebKit так не подменяются — для визуальных тестов нужны браузеры Playwright.

Отдельные наборы:

```bash
pnpm test:api          # без браузера
pnpm test:e2e          # путь в Chromium
pnpm test:visual       # три браузера и скриншоты
pnpm test:security
pnpm test:list         # проверить, что тесты собрались
```

Первый запуск визуальных тестов создаёт эталоны:

```bash
pnpm test:visual -- --update-snapshots
```

Скриншоты — локальный артефакт, в git не коммитим: они привязаны к браузеру, ОС и viewport. Когда появится стабильная CI-среда, эталоны сгенерируем в ней и только тогда закоммитим. Сейчас каталог `tests/ui/*-snapshots/` в `.gitignore`.

## Ограничения

- Только stage (`pr-<n>.shugraph.feytox.ru`), не prod. Stage одноразовый: своя БД на PR, удаляется после закрытия.
- У сценариев фиксированные `event_id`, а очистки БД в API нет. Для проверки клика по решению нужен свежий stage; на повторном прогоне шаг пропускается с пометкой.
- Сейчас собирается только ферма: в `farm_and_address` и `farm_and_chain` проверяем появление фермы, вторую группу не ждём до реализации схем 2/3.
- XSS проверяется на подменённом ответе API: комментарий с HTML отрисовывается как текст, данные stage не меняются.
- Тесты идут с `workers: 1`, чтобы сценарии не пересекались. Порядок проектов: API → visual → E2E → security.
- Граф — canvas: данные узлов и рёбер проверяются через API, в браузере — только отрисовка, легенда и подписи.
- Три движка Playwright — это Chromium, Firefox и WebKit. Реальный Safari/iOS при необходимости проверяется отдельно через BrowserStack.
- Активные security-сканы и ZAP запускаем только на одноразовом stage, не на prod.

## Security отдельно

В коде автоматизированы только безопасные проверки: XSS в комментарии, битый JSON, кривые `cursor`/`group_id`, пустой комментарий, длинный `event_id`. Это не заменяет DAST.

- OWASP ZAP baseline/API scan — заголовки, CORS, ошибки, доступные ручки.
- Trivy — CVEs в образах.
- `govulncheck` и `pnpm audit` уже есть в CI (`checks.yml`), их предупреждения разбираем по мере появления.

## Результаты прогонов

25.11.2026, локальный стенд: 27 из 28. Stage `pr-49.shugraph.feytox.ru`: 25 из 28, два security-теста упали по таймауту сети и прошли при повторном прогоне.

- `SEC-3` падает намеренно и на локальном стенде, и на stage: `PUT /api/v1/events/.../{event_id}` принимает `event_id` длиной 4096 символов (`201`, повтор — `204`). Ограничения длины нет, ожидается `4xx`. Нужна задача на лимит.
- `SEC-2` уточнён: битый JSON даёт `422`, а не `400`; тест принимает любой `4xx` без `500`.
- В комбинированных сценариях пока создаётся только ферма — это ожидаемое состояние текущего спринта.
