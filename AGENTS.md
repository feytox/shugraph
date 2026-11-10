# AGENTS.md

- Задачи в YouTrack `GSD-<n>`. Ветка `GSD-<n>` от `main`, без суффиксов: по имени ветки lefthook подставляет префикс в коммит.
- Если задача опирается на невлитую ветку другой задачи, её ветка отходит от этой ветки, а PR смотрит в неё же, пока та не влита.
- Коммит `[GSD-<n>] <type>: <description>` на английском, одной строкой, типы `feat`, `fix`, `refactor`, `docs`, `chore`, `test`.
- Контекст продукта: `docs/rfc/01-zeroth-rfc/README.md`, термины: `docs/definitions.md`.
- Перед работой в `backend/`, `frontend/` или `infra/` прочитай `AGENTS.md` в этой папке.
