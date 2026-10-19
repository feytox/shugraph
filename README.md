# Граф связей для выявления мошеннических групп

## Git-хуки

В репозитории настроен [lefthook](https://github.com/evilmartians/lefthook). Хук `commit-msg` добавляет к сообщению коммита префикс задачи из имени ветки: на ветке `GSD-N` коммит `feat: x` превращается в `[GSD-N] feat: x`. На `main` и ветках с другими именами сообщение не меняется.

**Варианты установки по кастам:**
- ЦАРЬ:
    ```bash
    brew install lefthook && lefthook install
    ```
- ГОЙДА ЖИТЕЛЬ:
    ```bash
    pnpm setup && pnpm add -g lefthook && lefthook install
    ```
- НИЩУК:
    ```bash
    go install github.com/evilmartians/lefthook/v2@latest && lefthook install
    ```
