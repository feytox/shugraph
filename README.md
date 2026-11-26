# Граф связей для выявления мошеннических групп

## Локальный запуск сценариев

Запустите `docker compose --env-file .env.example up -d --build --wait`. В [админке](http://localhost:3001/admin/) запустите `device_farm). После обработки событий в [рабочем месте аналитика](http://localhost:3000/) появится группа «Ферма».

## Git-хуки

В репозитории настроен [lefthook](https://github.com/evilmartians/lefthook). Хук `commit-msg` добавляет к сообщению коммита префикс задачи из имени ветки: на ветке `GSD-N` коммит `feat: x` превращается в `[GSD-N] feat: x`. На `main` и ветках с другими именами сообщение не меняется.

**Установка** (любой вариант, остальные способы: https://lefthook.dev/install/):
- macOS / Linux с Homebrew:
    ```bash
    brew install lefthook && lefthook install
    ```
- через pnpm:
    ```bash
    pnpm setup && pnpm add -g lefthook && lefthook install
    ```
- через Go:
    ```bash
    go install github.com/evilmartians/lefthook/v2@latest && lefthook install
    ```
